import { useCallback, useRef, useState } from 'react'
import { wails } from '../lib/wails'
import {
  createRegistry,
  loadAiConfigs,
} from '../lib/ai/provider'
import {
  assemblePayload,
  chunkText,
  windowsFor,
} from '../lib/ai/chunker'
import {
  appendBlockToMarkdown,
  makeBlock,
  type KnowledgeBlock,
} from '../lib/ai/output_parser'

interface Doc {
  id: string
  title: string
  text: string
  sourceType: string
  pathOrUrl: string
}

export default function Study() {
  const [docs, setDocs] = useState<Doc[]>([])
  const [activeId, setActiveId] = useState<string | null>(null)
  const [status, setStatus] = useState('')
  const [blocks, setBlocks] = useState<KnowledgeBlock[]>([])
  const [knowledgeMd, setKnowledgeMd] = useState('')
  const sourceRef = useRef<HTMLDivElement>(null)

  const active = docs.find((d) => d.id === activeId) || null

  const addDoc = useCallback((doc: Doc) => {
    setDocs((prev) => [...prev, doc])
    setActiveId(doc.id)
    setStatus(`Loaded: ${doc.title} (${doc.text.length} chars)`)
  }, [])

  const selectionText = () => {
    const sel = window.getSelection()
    if (!sel || sel.isCollapsed) return ''
    if (sourceRef.current && sourceRef.current.contains(sel.anchorNode)) {
      return String(sel)
    }
    return String(sel)
  }

  const openFile = async () => {
    let path = await wails.openFile()
    if (!path) {
      path = window.prompt('File path (native dialog unavailable).', '') || ''
    }
    if (!path) return

    let text = ''
    try {
      text = await wails.extractDocument(path)
    } catch (e) {
      setStatus('Extract failed: ' + String(e))
    }

    if (!text) {
      setStatus('Path recorded. Use Paste text if extract is unavailable.')
      addDoc({
        id: 'doc-' + Date.now().toString(36),
        title: path.split(/[/\\]/).pop() || path,
        text: '',
        sourceType: 'external',
        pathOrUrl: path,
      })
      return
    }

    addDoc({
      id: 'doc-' + Date.now().toString(36),
      title: path.split(/[/\\]/).pop() || path,
      text,
      sourceType: 'external',
      pathOrUrl: path,
    })
  }

  const pasteText = () => {
    const t = window.prompt('Paste document text')
    if (!t) return
    addDoc({
      id: 'doc-' + Date.now().toString(36),
      title: 'Pasted text',
      text: t,
      sourceType: 'paste',
      pathOrUrl: '',
    })
  }

  const transferExact = () => {
    const text = selectionText()
    if (!text.trim()) {
      setStatus('Select text in the source pane first')
      return
    }
    const block = makeBlock(text.trim(), {
      type: 'exact',
      sourceId: active?.id || 'sel',
      provider: 'local',
      model: 'exact-transfer',
    })
    setBlocks((b) => [...b, block])
    const sourceMap = active
      ? {
          [active.id]: {
            title: active.title,
            pathOrUrl: active.pathOrUrl,
          },
        }
      : {}
    setKnowledgeMd((md) => appendBlockToMarkdown(md, block, sourceMap))
    setStatus('Transferred exact selection')
  }

  const summarize = async () => {
    const configs = loadAiConfigs()
    if (!configs.length) {
      setStatus('Add an API key in Settings → AI providers')
      return
    }
    const registry = createRegistry(configs)
    let text = selectionText()
    if (!text && active) text = active.text
    if (!text?.trim()) {
      setStatus('Nothing to summarize — select text or open a document')
      return
    }

    const chunks = chunkText(text, {
      sourceId: active?.id || 'sel',
      targetTokens: 1600,
      maxTokens: 2200,
    })
    const windows = windowsFor(chunks, 6000, 500)
    setStatus(
      `Summarizing ${chunks.length} chunk(s) in ${windows.length} window(s)…`,
    )

    const system =
      'You are an academic research assistant. Summarise clearly for a university student. ' +
      'Preserve technical terms and numbers. Do not invent facts. Use short paragraphs. ' +
      'Mark key claims with [1], [2] when useful.'

    for (let w = 0; w < windows.length; w++) {
      const payload = assemblePayload(windows[w], {
        instruction: 'Summarise the chunks above.',
      })
      try {
        const result = await registry.completeWithFailover({
          system,
          user: payload,
        })
        let pageRange: number[] | null = null
        if (windows[w].length) {
          const pages = windows[w].map((c) => c.approxPage)
          pageRange = [Math.min(...pages), Math.max(...pages)]
        }
        const block = makeBlock(result.text, {
          type: 'summary',
          sourceId: active?.id || 'sel',
          chunkIds: windows[w].map((c) => c.id),
          pageRange,
          model: configs[0]?.model || '',
          provider: result.providerName || result.providerId,
        })
        setBlocks((b) => [...b, block])
        const sourceMap = active
          ? {
              [active.id]: {
                title: active.title,
                pathOrUrl: active.pathOrUrl,
              },
            }
          : {}
        setKnowledgeMd((md) => appendBlockToMarkdown(md, block, sourceMap))
        setStatus(
          `Inserted summary from ${result.providerName} (window ${w + 1}/${windows.length})`,
        )
      } catch (e) {
        setStatus('LLM error: ' + (e instanceof Error ? e.message : String(e)))
        return
      }
    }
  }

  const showChunkInfo = () => {
    if (!active) {
      setStatus('Open a document first')
      return
    }
    const chunks = chunkText(active.text, { sourceId: active.id })
    const tokens = chunks.reduce((a, c) => a + (c.tokenEstimate || 0), 0)
    setStatus(`${chunks.length} chunks · ~${tokens} tokens total`)
  }

  const exportMd = () => {
    const blob = new Blob([knowledgeMd || ''], {
      type: 'text/markdown;charset=utf-8',
    })
    const a = document.createElement('a')
    a.href = URL.createObjectURL(blob)
    a.download =
      'conductino-knowledge-' + new Date().toISOString().slice(0, 10) + '.md'
    a.click()
    URL.revokeObjectURL(a.href)
    setStatus('Export started')
  }

  const clearKnowledge = () => {
    setBlocks([])
    setKnowledgeMd('')
    setStatus('Knowledge pane cleared')
  }

  return (
    <div className="grid h-full min-h-0 grid-cols-1 grid-rows-2 bg-bg text-fg md:grid-cols-2 md:grid-rows-1">
      <div className="flex min-h-0 flex-col border-r border-border">
        <div className="flex flex-wrap items-center gap-1.5 border-b border-border bg-elev px-3 py-2">
          <button type="button" className="primary-btn !mt-0" onClick={openFile}>
            Open file…
          </button>
          <button type="button" className="tool-btn !w-auto px-2" onClick={pasteText}>
            Paste text
          </button>
          <select
            className="max-w-[12rem] rounded border border-border bg-bg px-1.5 py-1 text-fg"
            aria-label="Open documents"
            value={activeId || ''}
            onChange={(e) => setActiveId(e.target.value)}
          >
            {docs.map((d) => (
              <option key={d.id} value={d.id}>
                {d.title}
              </option>
            ))}
          </select>
        </div>
        <div
          ref={sourceRef}
          className="flex-1 overflow-auto whitespace-pre-wrap p-3 text-[0.92rem] leading-relaxed select-text"
          tabIndex={0}
        >
          {active?.text || (
            <p className="text-muted">
              Open or paste a document. Select text then use the actions below.
            </p>
          )}
        </div>
        <div className="flex flex-wrap gap-1.5 border-t border-border bg-elev px-3 py-2">
          <button type="button" className="tool-btn !w-auto px-2" onClick={transferExact}>
            Transfer exact
          </button>
          <button type="button" className="primary-btn !mt-0" onClick={summarize}>
            Summarize &amp; transfer
          </button>
          <button type="button" className="tool-btn !w-auto px-2" onClick={showChunkInfo}>
            Chunk info
          </button>
        </div>
        <div className="min-h-[1.4rem] px-3 pb-2 pt-1 text-xs text-muted">{status}</div>
      </div>

      <div className="flex min-h-0 flex-col">
        <div className="flex flex-wrap items-center gap-1.5 border-b border-border bg-elev px-3 py-2">
          <strong>Knowledge document</strong>
          <button type="button" className="tool-btn !w-auto px-2" onClick={clearKnowledge}>
            Clear
          </button>
          <button type="button" className="tool-btn !w-auto px-2" onClick={exportMd}>
            Export MD
          </button>
        </div>
        <div className="flex-1 overflow-auto p-3 select-text">
          {blocks.length === 0 && (
            <p className="text-muted">Summaries and exact transfers appear here.</p>
          )}
          {blocks.map((b) => (
            <article
              key={b.id}
              className="mb-3 overflow-hidden rounded-md border border-border bg-elev"
            >
              <header className="border-b border-border px-2.5 py-1.5 text-xs opacity-75">
                {[b.type, b.provider, b.model, b.pageRange && `pp. ${b.pageRange.join('-')}`]
                  .filter(Boolean)
                  .join(' · ')}
              </header>
              <div className="whitespace-pre-wrap px-3 py-2">{b.text}</div>
            </article>
          ))}
        </div>
      </div>
    </div>
  )
}
