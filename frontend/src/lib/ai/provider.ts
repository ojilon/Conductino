export type ProviderStyle = 'google' | 'openai' | 'raw'

export interface ProviderConfig {
  id: string
  name: string
  endpoint: string
  apiKey: string
  model?: string
  headers?: Record<string, string>
  maxTokens?: number
  style?: ProviderStyle
}

function guessStyle(endpoint: string): ProviderStyle {
  if (/generativelanguage\.googleapis|gemini/i.test(endpoint)) return 'google'
  return 'openai'
}

function defaultHeaders(cfg: ProviderConfig): Record<string, string> {
  const h: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(cfg.headers || {}),
  }
  const style = cfg.style || guessStyle(cfg.endpoint)
  if (style === 'openai' || /openrouter|groq|openai/i.test(cfg.endpoint)) {
    h['Authorization'] = 'Bearer ' + cfg.apiKey
  }
  return h
}

function buildBody(cfg: ProviderConfig, system: string, user: string): unknown {
  const style = cfg.style || guessStyle(cfg.endpoint)
  if (style === 'google') {
    const text = system ? system + '\n\n' + user : user
    return {
      contents: [{ role: 'user', parts: [{ text }] }],
      generationConfig: {
        maxOutputTokens: cfg.maxTokens || 2048,
        temperature: 0.3,
      },
    }
  }
  const messages: { role: string; content: string }[] = []
  if (system) messages.push({ role: 'system', content: system })
  messages.push({ role: 'user', content: user })
  return {
    model: cfg.model,
    messages,
    max_tokens: cfg.maxTokens || 2048,
    temperature: 0.3,
    stream: false,
  }
}

function extractText(cfg: ProviderConfig, data: unknown): string {
  const style = cfg.style || guessStyle(cfg.endpoint)
  try {
    const d = data as Record<string, unknown>
    if (style === 'google') {
      const candidates = d.candidates as { content?: { parts?: { text?: string }[] } }[] | undefined
      return candidates?.[0]?.content?.parts?.[0]?.text || ''
    }
    const choices = d.choices as { message?: { content?: string } }[] | undefined
    return choices?.[0]?.message?.content || ''
  } catch {
    return ''
  }
}

export function createProvider(cfg: ProviderConfig) {
  if (!cfg?.endpoint) throw new Error('provider: endpoint required')

  return {
    id: cfg.id || 'default',
    name: cfg.name || cfg.id || 'LLM',
    config: cfg,

    isAvailable() {
      return !!(cfg.apiKey && cfg.endpoint)
    },

    async complete(opts: { system?: string; user: string; signal?: AbortSignal }): Promise<string> {
      if (!opts?.user) throw new Error('provider.complete: user text required')
      if (!this.isAvailable()) throw new Error('provider: missing apiKey or endpoint')

      let url = cfg.endpoint
      if ((cfg.style || guessStyle(cfg.endpoint)) === 'google' && cfg.apiKey) {
        url += (url.includes('?') ? '&' : '?') + 'key=' + encodeURIComponent(cfg.apiKey)
      }

      const res = await fetch(url, {
        method: 'POST',
        headers: defaultHeaders(cfg),
        body: JSON.stringify(buildBody(cfg, opts.system || '', opts.user)),
        signal: opts.signal,
      })

      if (!res.ok) {
        let errText = ''
        try {
          errText = await res.text()
        } catch {
          /* */
        }
        const err = new Error('LLM HTTP ' + res.status + ': ' + errText.slice(0, 400)) as Error & {
          status?: number
        }
        err.status = res.status
        throw err
      }

      const data = await res.json()
      const text = extractText(cfg, data)
      if (!text) throw new Error('provider: empty model response')
      return text
    },
  }
}

export type Provider = ReturnType<typeof createProvider>

export function createRegistry(configs: ProviderConfig[]) {
  let providers = (configs || []).map(createProvider)
  return {
    list() {
      return providers.slice()
    },
    get(id: string) {
      return providers.find((p) => p.id === id) || null
    },
    async completeWithFailover(opts: { system?: string; user: string; signal?: AbortSignal }) {
      let lastErr: unknown = null
      for (const p of providers) {
        if (!p.isAvailable()) continue
        try {
          return {
            text: await p.complete(opts),
            providerId: p.id,
            providerName: p.name,
          }
        } catch (e) {
          lastErr = e
          continue
        }
      }
      throw lastErr || new Error('No available LLM provider')
    },
    setConfigs(configs: ProviderConfig[]) {
      providers = (configs || []).map(createProvider)
    },
  }
}

const AI_KEY = 'conductino.ai'

export function loadAiConfigs(): ProviderConfig[] {
  try {
    const raw = localStorage.getItem(AI_KEY)
    if (!raw) return []
    const parsed = JSON.parse(raw)
    return Array.isArray(parsed) ? parsed : parsed ? [parsed] : []
  } catch {
    return []
  }
}

export function saveAiConfigs(list: ProviderConfig[]) {
  localStorage.setItem(AI_KEY, JSON.stringify(list))
}

export function presetEndpoint(name: string): Omit<ProviderConfig, 'apiKey'> & { apiKey?: string } {
  if (name === 'google') {
    return {
      id: 'google',
      name: 'Google AI Studio',
      endpoint:
        'https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent',
      model: 'gemini-2.0-flash',
      style: 'google',
    }
  }
  if (name === 'openrouter') {
    return {
      id: 'openrouter',
      name: 'OpenRouter',
      endpoint: 'https://openrouter.ai/api/v1/chat/completions',
      model: 'google/gemini-2.0-flash-001',
      style: 'openai',
    }
  }
  if (name === 'groq') {
    return {
      id: 'groq',
      name: 'Groq',
      endpoint: 'https://api.groq.com/openai/v1/chat/completions',
      model: 'llama-3.3-70b-versatile',
      style: 'openai',
    }
  }
  return {
    id: 'custom',
    name: 'Custom',
    endpoint: '',
    model: '',
    style: 'openai',
  }
}
