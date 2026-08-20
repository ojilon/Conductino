/**
 * Thin wrappers around Wails-generated bindings + runtime.
 * Window chrome uses runtime. In-app navigation uses an iframe in React;
 * BrowserOpenURL is only for the optional "Open external" fallback.
 * File dialog / document extract stay on App bindings.
 */

import {
  BrowserOpenURL,
  WindowMinimise,
  WindowToggleMaximise,
  Quit,
} from '../../wailsjs/runtime/runtime'

async function callGo<T>(fn: () => Promise<T>, fallback?: T): Promise<T> {
  try {
    return await fn()
  } catch (e) {
    console.warn('[wails]', e)
    if (fallback !== undefined) return fallback
    throw e
  }
}

function goApp(): Record<string, (...args: unknown[]) => Promise<unknown>> | null {
  const g = (window as unknown as { go?: { main?: { App?: Record<string, (...a: unknown[]) => Promise<unknown>> } } }).go
  return g?.main?.App ?? null
}

export const wails = {
  /** System browser — only for sites that block iframe embedding. */
  openURL(url: string) {
    try {
      BrowserOpenURL(url)
    } catch (e) {
      console.warn('[wails] BrowserOpenURL', e)
    }
  },

  minimise() {
    try {
      WindowMinimise()
    } catch {
      /* */
    }
  },

  toggleMaximise() {
    try {
      WindowToggleMaximise()
    } catch {
      /* */
    }
  },

  quit() {
    try {
      Quit()
    } catch {
      /* */
    }
  },

  async openFile(): Promise<string> {
    const app = goApp()
    if (app?.OpenFile) {
      const path = await callGo(() => app.OpenFile() as Promise<string>, '')
      return typeof path === 'string' ? path : ''
    }
    return ''
  },

  async extractDocument(path: string): Promise<string> {
    const app = goApp()
    if (app?.ExtractDocument) {
      const text = await callGo(() => app.ExtractDocument(path) as Promise<string>, '')
      return typeof text === 'string' ? text : ''
    }
    return ''
  },

  async importDocument(path: string): Promise<string> {
    const app = goApp()
    if (app?.ImportDocument) {
      const dest = await callGo(() => app.ImportDocument(path) as Promise<string>, '')
      return typeof dest === 'string' ? dest : ''
    }
    return ''
  },

  async greet(name: string): Promise<string> {
    const app = goApp()
    if (app?.Greet) {
      return (await callGo(() => app.Greet(name) as Promise<string>, `Hello ${name}`)) as string
    }
    return `Hello ${name}`
  },
}
