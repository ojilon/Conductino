interface Props {
  onOpenStudy: () => void
}

export default function Welcome({ onOpenStudy }: Props) {
  return (
    <div className="p-8">
      <div className="mx-auto mt-10 max-w-lg rounded-app border border-border bg-elev p-7">
        <h1 className="m-0 mb-2 text-[22px] font-semibold">New Tab</h1>
        <p className="text-muted">Search or type a URL in the address bar.</p>
        <p className="text-xs text-muted">
          Pages load inside the app. Study workspace stays local for documents and AI.
        </p>
        <p className="text-xs text-muted">
          <button type="button" className="primary-btn" onClick={onOpenStudy}>
            Open study workspace
          </button>
        </p>
      </div>
    </div>
  )
}
