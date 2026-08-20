import { useState } from 'react'
import {
  ENGINES,
  type AppSettings,
  type EngineId,
  type ThemeName,
} from '../lib/settings'
import {
  loadAiConfigs,
  presetEndpoint,
  saveAiConfigs,
  type ProviderConfig,
} from '../lib/ai/provider'

interface Props {
  settings: AppSettings
  onTheme: (t: ThemeName) => void
  onEngine: (e: EngineId) => void
  onDone: () => void
}

export default function SettingsPanel({
  settings,
  onTheme,
  onEngine,
  onDone,
}: Props) {
  const [preset, setPreset] = useState('google')
  const [apiKey, setApiKey] = useState('')
  const [endpoint, setEndpoint] = useState('')
  const [model, setModel] = useState('')
  const [msg, setMsg] = useState('')

  const saveAi = () => {
    const base = presetEndpoint(preset) as ProviderConfig
    if (endpoint.trim()) base.endpoint = endpoint.trim()
    if (model.trim()) base.model = model.trim()
    base.apiKey = apiKey.trim()
    if (!base.apiKey || !base.endpoint) {
      alert('API key and endpoint required')
      return
    }
    const list = loadAiConfigs().filter((c) => c.id !== base.id)
    list.unshift(base)
    saveAiConfigs(list)
    setMsg('Saved ' + base.name)
  }

  return (
    <div className="p-8">
      <div className="mx-auto mt-10 max-w-lg rounded-app border border-border bg-elev p-7">
        <h2 className="m-0 mb-2 text-[22px] font-semibold">Settings</h2>

        <label className="field">
          <span>Theme</span>
          <select
            value={settings.theme}
            onChange={(e) => onTheme(e.target.value as ThemeName)}
          >
            <option value="aurora-dark">Aurora Dark</option>
            <option value="aurora-light">Aurora Light</option>
          </select>
        </label>

        <label className="field">
          <span>Search engine</span>
          <select
            value={settings.engine}
            onChange={(e) => onEngine(e.target.value as EngineId)}
          >
            {(Object.keys(ENGINES) as EngineId[]).map((id) => (
              <option key={id} value={id}>
                {ENGINES[id].name}
              </option>
            ))}
          </select>
        </label>

        <hr className="my-4 border-border" />

        <h3 className="m-0 text-base font-semibold">AI providers</h3>
        <p className="text-xs text-muted">
          Keys stay in localStorage. Network calls run in the webview.
        </p>

        <label className="field">
          <span>Provider preset</span>
          <select value={preset} onChange={(e) => setPreset(e.target.value)}>
            <option value="google">Google AI Studio (Gemini)</option>
            <option value="openrouter">OpenRouter</option>
            <option value="groq">Groq</option>
            <option value="custom">Custom endpoint</option>
          </select>
        </label>

        <label className="field">
          <span>API key</span>
          <input
            type="password"
            autoComplete="off"
            placeholder="paste key"
            value={apiKey}
            onChange={(e) => setApiKey(e.target.value)}
          />
        </label>

        <label className="field">
          <span>Endpoint (custom)</span>
          <input
            type="text"
            placeholder="https://..."
            value={endpoint}
            onChange={(e) => setEndpoint(e.target.value)}
          />
        </label>

        <label className="field">
          <span>Model</span>
          <input
            type="text"
            placeholder="gemini-2.0-flash"
            value={model}
            onChange={(e) => setModel(e.target.value)}
          />
        </label>

        <button type="button" className="primary-btn" onClick={saveAi}>
          Save AI config
        </button>
        <button type="button" className="primary-btn ml-2" onClick={onDone}>
          Done
        </button>
        {msg && <p className="mt-2 text-xs text-accent">{msg}</p>}
      </div>
    </div>
  )
}
