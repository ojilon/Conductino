export type ThemeName = 'aurora-dark' | 'aurora-light'
export type EngineId = 'duckduckgo' | 'google' | 'bing' | 'startpage'

export interface AppSettings {
  theme: ThemeName
  engine: EngineId
}

const KEY = 'conductino.settings'

export const ENGINES: Record<
  EngineId,
  { name: string; url: string }
> = {
  duckduckgo: { name: 'DuckDuckGo', url: 'https://duckduckgo.com/?q=%s' },
  google: { name: 'Google', url: 'https://www.google.com/search?q=%s' },
  bing: { name: 'Bing', url: 'https://www.bing.com/search?q=%s' },
  startpage: {
    name: 'Startpage',
    url: 'https://www.startpage.com/sp/search?query=%s',
  },
}

export function loadSettings(): AppSettings {
  try {
    const raw = localStorage.getItem(KEY)
    if (raw) {
      const p = JSON.parse(raw) as Partial<AppSettings>
      return {
        theme: p.theme === 'aurora-light' ? 'aurora-light' : 'aurora-dark',
        engine: (p.engine && ENGINES[p.engine] ? p.engine : 'duckduckgo') as EngineId,
      }
    }
  } catch {
    /* ignore */
  }
  return { theme: 'aurora-dark', engine: 'duckduckgo' }
}

export function saveSettings(s: AppSettings) {
  try {
    localStorage.setItem(KEY, JSON.stringify(s))
  } catch {
    /* ignore */
  }
}

export function applyTheme(theme: ThemeName) {
  document.documentElement.setAttribute('data-theme', theme)
}

export function looksLikeUrl(input: string): boolean {
  const s = input.trim()
  if (!s || /\s/.test(s)) return false
  if (/^(https?:\/\/|about:|file:)/i.test(s)) return true
  if (/^[a-z0-9.-]+\.[a-z]{2,}([\/:].*)?$/i.test(s)) return true
  if (/^localhost(:\d+)?([\/:].*)?$/i.test(s)) return true
  return false
}

export function normalizeUrl(input: string): string {
  const s = input.trim()
  if (/^(https?:\/\/|about:|file:)/i.test(s)) return s
  return 'https://' + s
}

export function buildSearchUrl(query: string, engine: EngineId): string {
  const eng = ENGINES[engine] || ENGINES.duckduckgo
  return eng.url.replace('%s', encodeURIComponent(query))
}
