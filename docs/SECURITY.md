# Conductino Desktop — Security notes

Living document. Many items are **future work**; capture risks early so tools (JS injection, logins, WASM) are designed safely.

## Architecture reminders

| Surface | Trust |
|---------|--------|
| Chrome shell (`http://127.0.0.1:8080/`) | App-controlled HTML/JS |
| Content WebView2 | **Untrusted** web content (arbitrary sites) |
| C++ `conductino_core` | Local only — no browsing network |
| Go host binds | Privileged — only expose minimal APIs |

Remote pages must never run with the same privileges as the chrome shell.

---

## JS injection into content (study tools)

Planned: inject helpers (highlight, extract text, annotate) into the **content** WebView after load.

**Risks**

- Page JS can read or redefine injected globals.
- Injected code can be observed by the page (fingerprint / anti-bot).
- Accidental `eval` of page-controlled strings → XSS into the tool channel.

**Guidelines (when you add injection)**

1. Prefer `AddScriptToExecuteOnDocumentCreated` / isolated world if the WebView2 API allows; avoid sharing the page's main world when possible.
2. Never pass unsanitized page HTML/strings into chrome binds.
3. Content → Go messages: allowlist message types; validate JSON schema.
4. Do not expose `hostNavigate`, notes save, or settings from content scripts without origin checks.
5. Log injections in debug builds only; strip verbose logs in release.

---

## Sandbox / WASM (far future)

- Heavy study logic (parsers, models) should run in **WASM or native** behind a narrow API, not as free-form JS in the page.
- WASM still needs a clear capability boundary (no ambient filesystem/network unless explicit).
- Chrome process remains the policy enforcer; WASM is not a security boundary by itself.

---

## Logins, sessions, cookies

**Current:** content WebView2 uses its own user-data folder (`conductino-data/content-webview`), separate from chrome.

**When adding account features**

| Topic | Direction |
|-------|-----------|
| Site logins | Stay in content WebView profile; user expects normal browser cookies |
| App accounts (Conductino sync) | Separate token store in `conductino_core` / OS secret store — not in page JS |
| Clearing data | Explicit UI: clear content profile vs clear notes |
| Phishing | Omnibox shows real URL; never trust page-rendered URL chrome |

Do not put long-lived app secrets into `localStorage` of untrusted pages.

---

## Go ↔ JS binds

- Treat every `webview.Bind` as a **privileged syscall**.
- Prefer explicit methods (`hostSidebarOpen`, `hostNavigate`) over a generic `hostEval`.
- If a bind accepts a URL or path, validate scheme (`https`, `http`, `about`) before use.

---

## Proxy / network

- **Do not** reintroduce Go-side page fetching for display (CORS identity issues; larger attack surface).
- Native WebView2 owns TLS and challenges (Cloudflare, etc.).

---

## Checklist before shipping a content tool

- [ ] Runs only after navigation completed
- [ ] No privileged binds callable from page JS
- [ ] Messages schema-validated
- [ ] Works with third-party cookies / logins without leaking app tokens
- [ ] Documented in this file

---

## Related

- `docs/SHELL.md` — chrome vs content surfaces
- `docs/BRIDGE.md` — bind list
- `docs/BACKEND.md` — native core
