# Testing

## Smoke (`wails dev`)

1. Window opens with tabs + toolbar.
2. Omnibox: open `https://example.com` → iframe loads in-app.
3. Sidebar (right) → Settings / Study.
4. Study → Paste text → Transfer exact / Chunk info.
5. Settings → save AI key (optional) → Summarize if keyed.

## Build

```bat
cd frontend
wails build
```

Run the binary from `build/bin`.
