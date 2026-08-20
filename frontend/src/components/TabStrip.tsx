export interface Tab {
  id: number
  title: string
  url: string
  active: boolean
}

interface Props {
  tabs: Tab[]
  onNew: () => void
  onClose: (id: number) => void
  onActivate: (id: number) => void
}

export default function TabStrip({ tabs, onNew, onClose, onActivate }: Props) {
  return (
    <div
      className="flex h-tabstrip items-end gap-1 border-b border-border bg-bg px-2 pt-1"
      role="tablist"
      aria-label="Tabs"
    >
      <div className="flex min-w-0 flex-1 gap-1 overflow-x-auto [scrollbar-width:none] [&::-webkit-scrollbar]:hidden">
        {tabs.map((tab) => (
          <button
            key={tab.id}
            type="button"
            role="tab"
            aria-selected={tab.active}
            className={
              'flex h-[30px] max-w-[200px] min-w-[80px] items-center gap-2 rounded-t-lg border border-b-0 px-2.5 ' +
              (tab.active
                ? 'border-border bg-tab-active text-fg'
                : 'border-transparent text-muted hover:bg-elev hover:text-fg')
            }
            onClick={() => onActivate(tab.id)}
          >
            <span className="min-w-0 flex-1 truncate text-left">{tab.title}</span>
            <span
              className="flex h-[18px] w-[18px] shrink-0 items-center justify-center rounded text-xs text-muted hover:bg-elev2 hover:text-fg"
              title="Close tab"
              onClick={(e) => {
                e.stopPropagation()
                onClose(tab.id)
              }}
            >
              ×
            </span>
          </button>
        ))}
      </div>
      <button
        type="button"
        className="mb-0.5 h-7 w-7 shrink-0 rounded-md text-lg leading-none text-muted hover:bg-elev hover:text-fg"
        title="New tab"
        aria-label="New tab"
        onClick={onNew}
      >
        +
      </button>
    </div>
  )
}
