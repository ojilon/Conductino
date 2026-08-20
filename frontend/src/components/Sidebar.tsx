interface Props {
  open: boolean
  onClose: () => void
  onAction: (action: string) => void
}

const ITEMS: { action: string; label: string; stub?: boolean }[] = [
  { action: 'settings', label: '⚙ Settings' },
  { action: 'study', label: '📚 Study workspace' },
  { action: 'downloads', label: '⬇ Downloads', stub: true },
  { action: 'bookmarks', label: '☆ Bookmarks', stub: true },
]

export default function Sidebar({ open, onClose, onAction }: Props) {
  if (!open) return null
  return (
    <aside
      className="absolute bottom-0 right-0 top-0 z-20 flex w-sidebar flex-col border-l border-border bg-elev shadow-[-8px_0_24px_rgba(0,0,0,0.35)]"
      aria-label="Sidebar"
    >
      <div className="flex items-center justify-between px-3 pb-2 pt-3">
        <h2 className="m-0 text-[13px] font-semibold uppercase tracking-wide text-muted">
          Menu
        </h2>
        <button
          type="button"
          className="tool-btn"
          title="Close sidebar"
          onClick={onClose}
        >
          ×
        </button>
      </div>
      <nav className="flex flex-col gap-0.5 px-2">
        {ITEMS.map((item) => (
          <button
            key={item.action}
            type="button"
            className="flex w-full items-center gap-2 rounded-app-sm px-3 py-2.5 text-left text-fg hover:bg-elev2"
            onClick={() => onAction(item.action)}
          >
            {item.label}
            {item.stub && (
              <span className="ml-auto text-[10px] uppercase tracking-wider text-muted">
                soon
              </span>
            )}
          </button>
        ))}
      </nav>
    </aside>
  )
}
