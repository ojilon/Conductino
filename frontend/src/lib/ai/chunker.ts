export interface Chunk {
  id: string
  sourceId: string
  index: number
  text: string
  approxPage: number
  headingPath: string[]
  tokenEstimate: number
  charStart: number
  charEnd: number
}

export function estimateTokens(text: string): number {
  if (!text) return 0
  return Math.ceil(String(text).length / 4)
}

export function chunkText(
  text: string,
  opts?: {
    targetTokens?: number
    maxTokens?: number
    sourceId?: string
    charsPerPage?: number
  },
): Chunk[] {
  const target = opts?.targetTokens ?? 1800
  const maxT = opts?.maxTokens ?? 2400
  const sourceId = opts?.sourceId ?? 'doc-unknown'
  const charsPerPage = opts?.charsPerPage ?? 3000

  const raw = String(text || '').replace(/\r\n/g, '\n')
  let paras = raw
    .split(/\n\s*\n/)
    .map((p) => p.trim())
    .filter(Boolean)
  if (paras.length <= 1) {
    paras = raw
      .split(/\n/)
      .map((p) => p.trim())
      .filter(Boolean)
  }
  if (paras.length === 0 && raw.trim()) paras = [raw.trim()]

  const chunks: Chunk[] = []
  let buf: string[] = []
  let bufTokens = 0
  let charPos = 0
  let index = 0

  function flush() {
    if (!buf.length) return
    const body = buf.join('\n\n')
    let start = charPos - body.length
    if (start < 0) start = 0
    chunks.push({
      id: 'chunk-' + String(index).padStart(5, '0'),
      sourceId,
      index,
      text: body,
      approxPage: Math.floor(start / charsPerPage) + 1,
      headingPath: [],
      tokenEstimate: estimateTokens(body),
      charStart: start,
      charEnd: charPos,
    })
    index += 1
    buf = []
    bufTokens = 0
  }

  for (const p of paras) {
    const t = estimateTokens(p)
    if (bufTokens + t > maxT && buf.length) flush()
    if (t > maxT) {
      const step = Math.floor(maxT * 4 * 0.9)
      for (let s = 0; s < p.length; s += step) {
        const slice = p.slice(s, s + step)
        buf.push(slice)
        bufTokens += estimateTokens(slice)
        charPos += slice.length + 2
        flush()
      }
      continue
    }
    buf.push(p)
    bufTokens += t
    charPos += p.length + 2
    if (bufTokens >= target) flush()
  }
  flush()
  return chunks
}

export function assemblePayload(
  chunks: Chunk[],
  extras?: { preamble?: string; instruction?: string; altTexts?: { placeholder: string; altText: string }[] },
): string {
  const parts: string[] = []
  if (extras?.preamble) parts.push(extras.preamble)
  for (const c of chunks || []) {
    parts.push(`---\nChunk ${c.index} (approx page ${c.approxPage})\n${c.text}`)
  }
  if (extras?.altTexts?.length) {
    parts.push('\n--- Image descriptions ---')
    for (const img of extras.altTexts) {
      parts.push(img.placeholder + ' ' + img.altText)
    }
  }
  if (extras?.instruction) parts.push('\n---\n' + extras.instruction)
  return parts.join('\n\n')
}

export function windowsFor(
  chunks: Chunk[],
  maxTokens = 6000,
  systemOverhead = 400,
): Chunk[][] {
  const budget = maxTokens - systemOverhead
  const windows: Chunk[][] = []
  let cur: Chunk[] = []
  let curT = 0
  for (const c of chunks || []) {
    const t = c.tokenEstimate || estimateTokens(c.text)
    if (curT + t > budget && cur.length) {
      windows.push(cur)
      cur = []
      curT = 0
    }
    cur.push(c)
    curT += t
  }
  if (cur.length) windows.push(cur)
  return windows
}
