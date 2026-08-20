# GUI (React chrome)

## Layout

```
┌─────────────────────────────────────────────┐
│ Tab strip (+ new tab)                       │
├─────────────────────────────────────────────┤
│ Back/Fwd/Reload │ Omnibox │ Study │ Sidebar │
├──────────────────────────────────┬──────────┤
│ Content host                     │ Sidebar  │
│  welcome | browser iframe |      │ (right)  │
│  settings | study | stub         │ optional │
└──────────────────────────────────┴──────────┘
```

- **Sidebar** docks on the **right** when open.
- **browser** panel: in-app iframe for the active tab URL.
- **study** panel: split source / knowledge document.

Source: `frontend/src/App.tsx` and `frontend/src/components/*`.

## Theming

CSS variables on `[data-theme=aurora-dark|aurora-light]` in `src/index.css`, exposed to Tailwind as `bg`, `elev`, `fg`, `muted`, `accent`, etc.

## Extending

1. Add a panel id in `App.tsx`.
2. Add a component under `components/`.
3. Wire a sidebar or toolbar action.

Avoid putting network page-fetch logic in Go.
