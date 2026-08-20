interface Props {
  url: string
  onOpenExternal?: () => void
}

/**
 * In-app content surface. Uses an iframe inside the single Wails webview so
 * chrome (tabs/toolbar) stays visible. Sites that send X-Frame-Options / CSP
 * frame-ancestors may refuse to load; user can open externally as fallback.
 */
export default function BrowserView({ url, onOpenExternal }: Props) {
  return (
    <div className="relative flex h-full min-h-0 flex-col bg-bg">
      <div className="flex shrink-0 items-center gap-2 border-b border-border bg-elev px-3 py-1 text-xs text-muted">
        <span className="min-w-0 flex-1 truncate" title={url}>
          {url}
        </span>
        {onOpenExternal && (
          <button
            type="button"
            className="tool-btn !h-7 !w-auto px-2 text-xs"
            title="Open in system browser (sites that block embedding)"
            onClick={onOpenExternal}
          >
            Open external
          </button>
        )}
      </div>
      <iframe
        key={url}
        src={url}
        title="Content"
        className="min-h-0 w-full flex-1 border-0 bg-white"
        sandbox="allow-scripts allow-same-origin allow-forms allow-popups allow-popups-to-escape-sandbox allow-downloads"
        referrerPolicy="no-referrer-when-downgrade"
      />
    </div>
  )
}
