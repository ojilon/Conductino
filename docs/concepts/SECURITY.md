# Security notes

- **API keys** — stored in browser `localStorage` only; never sent to the C++ core by default.
- **Backend offline** — document extract and storage must not phone home.
- **iframe sandbox** — content iframe uses a restrictive `sandbox` attribute; still treat remote pages as untrusted.
- **External open** — system browser is explicit user action for non-embeddable sites.
- **File access** — only paths chosen via native dialog or user paste.

See also historical detail in older SECURITY drafts if present under archive.
