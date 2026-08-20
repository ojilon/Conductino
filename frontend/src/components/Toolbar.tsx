interface Props {
  omnibox: string
  onOmniboxChange: (v: string) => void
  onSubmit: () => void
  onSidebarToggle: () => void
  sidebarOpen: boolean
  onStudy: () => void
}

function schemeIcon(url: string) {
  if (!url) return '🔍'
  if (/^https:\/\//i.test(url)) return '🔒'
  if (/^http:\/\//i.test(url)) return '⚠'
  return '📄'
}

export default function Toolbar({
  omnibox,
  onOmniboxChange,
  onSubmit,
  onSidebarToggle,
  sidebarOpen,
  onStudy,
}: Props) {
  return (
    <div className="flex h-toolbar items-center gap-2 border-b border-border bg-elev px-2.5">
      <div className="flex shrink-0 gap-0.5">
        <button type="button" className="tool-btn" title="Back" disabled>
          ◀
        </button>
        <button type="button" className="tool-btn" title="Forward" disabled>
          ▶
        </button>
        <button type="button" className="tool-btn" title="Reload">
          ⟳
        </button>
      </div>

      <form
        className="flex h-8 min-w-0 flex-1 items-center gap-2 rounded-full border border-border bg-bg px-3 focus-within:border-accent2 focus-within:shadow-[0_0_0_2px_color-mix(in_srgb,var(--accent2)_25%,transparent)]"
        onSubmit={(e) => {
          e.preventDefault()
          onSubmit()
        }}
        autoComplete="off"
      >
        <span className="shrink-0 text-xs opacity-85" title="Scheme">
          {schemeIcon(omnibox)}
        </span>
        <input
          className="min-w-0 flex-1 border-0 bg-transparent text-[13px] text-fg outline-none placeholder:text-muted select-text"
          value={omnibox}
          onChange={(e) => onOmniboxChange(e.target.value)}
          spellCheck={false}
          autoComplete="off"
          placeholder="Search or enter address"
          aria-label="Address and search"
        />
      </form>

      <div className="flex shrink-0 gap-0.5">
        <button
          type="button"
          className="tool-btn"
          title="Study workspace"
          aria-label="Study workspace"
          onClick={onStudy}
        >
          📚
        </button>
        <button
          type="button"
          className="tool-btn"
          title="Toggle sidebar"
          aria-label="Toggle sidebar"
          aria-pressed={sidebarOpen}
          onClick={onSidebarToggle}
        >
          ☰
        </button>
      </div>
    </div>
  )
}
