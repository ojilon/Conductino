import { useCallback, useEffect, useState } from 'react'
import TabStrip, { type Tab } from './components/TabStrip'
import Toolbar from './components/Toolbar'
import Sidebar from './components/Sidebar'
import Welcome from './components/Welcome'
import SettingsPanel from './components/Settings'
import Study from './components/Study'
import BrowserView from './components/BrowserView'
import {
  applyTheme,
  buildSearchUrl,
  loadSettings,
  looksLikeUrl,
  normalizeUrl,
  saveSettings,
  type AppSettings,
  type EngineId,
  type ThemeName,
} from './lib/settings'
import { wails } from './lib/wails'

type Panel = 'welcome' | 'settings' | 'study' | 'stub' | 'browser'

export default function App() {
  const [settings, setSettings] = useState<AppSettings>(() => loadSettings())
  const [tabs, setTabs] = useState<Tab[]>([
    { id: 1, title: 'New Tab', url: '', active: true },
  ])
  const [nextId, setNextId] = useState(2)
  const [omnibox, setOmnibox] = useState('')
  const [sidebarOpen, setSidebarOpen] = useState(false)
  const [panel, setPanel] = useState<Panel>('welcome')
  const [stub, setStub] = useState({ title: 'Coming soon', body: '' })

  useEffect(() => {
    applyTheme(settings.theme)
    saveSettings(settings)
  }, [settings])

  const active = tabs.find((t) => t.active) || tabs[0]

  useEffect(() => {
    setOmnibox(active?.url || '')
    if (active?.url) setPanel('browser')
  }, [active?.id, active?.url])

  const updateActive = useCallback((patch: Partial<Tab>) => {
    setTabs((prev) =>
      prev.map((t) => (t.active ? { ...t, ...patch } : t)),
    )
  }, [])

  const navigateTo = useCallback(
    (input: string) => {
      const raw = (input || '').trim()
      if (!raw) return
      const url = looksLikeUrl(raw)
        ? normalizeUrl(raw)
        : buildSearchUrl(raw, settings.engine)
      updateActive({
        url,
        title: url.replace(/^https?:\/\//, '').split('/')[0] || 'Tab',
      })
      setOmnibox(url)
      setPanel('browser')
    },
    [settings.engine, updateActive],
  )

  const newTab = () => {
    setTabs((prev) => [
      ...prev.map((t) => ({ ...t, active: false })),
      { id: nextId, title: 'New Tab', url: '', active: true },
    ])
    setNextId((n) => n + 1)
    setPanel('welcome')
    setOmnibox('')
  }

  const closeTab = (id: number) => {
    setTabs((prev) => {
      let next = prev.filter((t) => t.id !== id)
      if (!next.length) {
        next = [{ id: nextId, title: 'New Tab', url: '', active: true }]
        setNextId((n) => n + 1)
        setPanel('welcome')
      } else if (!next.some((t) => t.active)) {
        next = next.map((t, i) => ({ ...t, active: i === 0 }))
      }
      return next
    })
  }

  const activateTab = (id: number) => {
    setTabs((prev) => {
      const next = prev.map((t) => ({ ...t, active: t.id === id }))
      const t = next.find((x) => x.id === id)
      if (t?.url) setPanel('browser')
      else setPanel('welcome')
      return next
    })
  }

  const onSidebarAction = (action: string) => {
    setSidebarOpen(false)
    if (action === 'settings') setPanel('settings')
    else if (action === 'study') setPanel('study')
    else if (action === 'downloads') {
      setStub({ title: 'Downloads', body: 'Download manager will live here.' })
      setPanel('stub')
    } else if (action === 'bookmarks') {
      setStub({ title: 'Bookmarks', body: 'Bookmarks UI will live here.' })
      setPanel('stub')
    }
  }

  return (
    <div className="flex h-full flex-col select-none">
      <TabStrip
        tabs={tabs}
        onNew={newTab}
        onClose={closeTab}
        onActivate={activateTab}
      />
      <Toolbar
        omnibox={omnibox}
        onOmniboxChange={setOmnibox}
        onSubmit={() => navigateTo(omnibox)}
        onSidebarToggle={() => setSidebarOpen((o) => !o)}
        sidebarOpen={sidebarOpen}
        onStudy={() => setPanel('study')}
      />
      <div className="relative flex min-h-0 flex-1">
        <main className="content-host min-w-0 flex-1 overflow-hidden bg-bg">
          {panel === 'welcome' && (
            <Welcome onOpenStudy={() => setPanel('study')} />
          )}
          {panel === 'browser' && active?.url && (
            <BrowserView
              url={active.url}
              onOpenExternal={() => wails.openURL(active.url)}
            />
          )}
          {panel === 'settings' && (
            <SettingsPanel
              settings={settings}
              onTheme={(theme: ThemeName) =>
                setSettings((s) => ({ ...s, theme }))
              }
              onEngine={(engine: EngineId) =>
                setSettings((s) => ({ ...s, engine }))
              }
              onDone={() => setPanel(active?.url ? 'browser' : 'welcome')}
            />
          )}
          {panel === 'study' && <Study />}
          {panel === 'stub' && (
            <div className="overflow-auto p-8">
              <div className="mx-auto mt-10 max-w-lg rounded-app border border-border bg-elev p-7">
                <h2 className="m-0 mb-2 text-[22px] font-semibold">{stub.title}</h2>
                <p className="text-muted">{stub.body}</p>
                <button
                  type="button"
                  className="primary-btn"
                  onClick={() => setPanel(active?.url ? 'browser' : 'welcome')}
                >
                  Back
                </button>
              </div>
            </div>
          )}
        </main>
        <Sidebar
          open={sidebarOpen}
          onClose={() => setSidebarOpen(false)}
          onAction={onSidebarAction}
        />
      </div>
    </div>
  )
}
