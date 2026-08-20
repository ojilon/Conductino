export interface KnowledgeBlock {
  id: string
  type: string
  text: string
  sourceId: string
  chunkIds: string[]
  pageRange: number[] | null
  model: string
  provider: string
  createdAt: string
  citations: { label: string; number: number }[]
  images: { id: string; placeholder: string }[]
  verified: boolean
  flags: string[]
}

const CITE_RE = /\[(\d+)\]/g
const IMG_RE = /\[\[IMG:([a-zA-Z0-9_-]+)\]\]/g

export function extractMarkers(text: string) {
  const citations: { label: string; number: number }[] = []
  const images: { id: string; placeholder: string }[] = []
  const seenCite: Record<string, boolean> = {}
  const seenImg: Record<string, boolean> = {}
  let m: RegExpExecArray | null
  CITE_RE.lastIndex = 0
  while ((m = CITE_RE.exec(text)) !== null) {
    const n = m[1]
    if (!seenCite[n]) {
      seenCite[n] = true
      citations.push({ label: '[' + n + ']', number: parseInt(n, 10) })
    }
  }
  IMG_RE.lastIndex = 0
  while ((m = IMG_RE.exec(text)) !== null) {
    const id = m[1]
    if (!seenImg[id]) {
      seenImg[id] = true
      images.push({ id, placeholder: m[0] })
    }
  }
  return { citations, images }
}

export function makeBlock(
  text: string,
  meta?: Partial<KnowledgeBlock>,
): KnowledgeBlock {
  const markers = extractMarkers(text)
  return {
    id: meta?.id || 'block-' + Date.now().toString(36),
    type: meta?.type || 'summary',
    text,
    sourceId: meta?.sourceId || '',
    chunkIds: meta?.chunkIds || [],
    pageRange: meta?.pageRange ?? null,
    model: meta?.model || '',
    provider: meta?.provider || '',
    createdAt: meta?.createdAt || new Date().toISOString(),
    citations: markers.citations,
    images: markers.images,
    verified: false,
    flags: [],
  }
}

export function appendBlockToMarkdown(
  existingMd: string,
  block: KnowledgeBlock,
  sourceMap?: Record<string, { title?: string; pathOrUrl?: string }>,
): string {
  const src = sourceMap?.[block.sourceId]
  const header =
    '\n\n<!-- block:' +
    block.id +
    ' source:' +
    (block.sourceId || '') +
    ' -->\n\n'
  let refs = ''
  if (block.citations?.length) {
    refs =
      '\n\n## References\n\n' +
      block.citations
        .map((c) => {
          const title = src?.title || 'Untitled source'
          const path = src?.pathOrUrl ? ' — ' + src.pathOrUrl : ''
          const pages = block.pageRange ? ', pp. ' + block.pageRange.join('-') : ''
          return c.label + ' ' + title + pages + path
        })
        .join('\n') +
      '\n'
  }
  return (existingMd || '') + header + block.text.trim() + '\n' + refs
}
