# In-app browser (real WebView2)

## Model

- **No iframes.** The omnibox navigates the **main WebView2** (`window.location`), the same engine Edge uses.
- Search words (`leaf`, `tea`) → configured search engine results page.
- Site names / URLs → the real site (ResearchGate, Cloudflare challenges, etc. work like a normal browser).
- Study / Library / Settings live on the Conductino **home** surface (asset UI).

## Floating tool bar (on web pages only)

After navigation, Conductino injects a small top bar:

| Button | Action |
|--------|--------|
| Conductino Home | Back to app UI |
| Copy selection | Clipboard |
| Selection → Study | Clipboard + open Study |
| Summarize selection | Clipboard + Study + summarize |
| Hide | Remove bar |

Tools run **only when you click**. Browsing stays normal.

## Limits

- Chrome (tabs/omnibox) is temporarily replaced by the web page until you press **Conductino Home**.
- True dual-pane (persistent chrome + content WebView2) needs a native child WebView2 control — future work.
- Injected bar may appear after 1–3s; click Hide if a site layout conflicts.
