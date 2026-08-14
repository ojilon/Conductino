/**
 * Unified LLM provider adapter.
 * Switch vendor by changing endpoint / apiKey / model only.
 * Network stays in the webview (fetch). Backend remains no-network.
 */
(function (global) {
  "use strict";

  /**
   * @typedef {Object} ProviderConfig
   * @property {string} id
   * @property {string} name
   * @property {string} endpoint
   * @property {string} apiKey
   * @property {string} [model]
   * @property {Object} [headers]
   * @property {number} [maxTokens]
   * @property {string} [style]  "google" | "openai" | "raw"
   */

  function defaultHeaders(cfg) {
    var h = Object.assign(
      { "Content-Type": "application/json" },
      cfg.headers || {}
    );
    // Common patterns; override via cfg.headers when needed.
    if (cfg.style === "openai" || /openrouter|groq|openai/i.test(cfg.endpoint)) {
      h["Authorization"] = "Bearer " + cfg.apiKey;
    }
    return h;
  }

  function buildBody(cfg, system, user) {
    var style = cfg.style || guessStyle(cfg.endpoint);
    if (style === "google") {
      // Google AI Studio / Gemini generateContent
      var contents = [];
      if (system) {
        contents.push({ role: "user", parts: [{ text: system + "\n\n" + user }] });
      } else {
        contents.push({ role: "user", parts: [{ text: user }] });
      }
      return {
        contents: contents,
        generationConfig: {
          maxOutputTokens: cfg.maxTokens || 2048,
          temperature: 0.3,
        },
      };
    }
    // OpenAI-compatible (OpenRouter, Groq, many others)
    var messages = [];
    if (system) messages.push({ role: "system", content: system });
    messages.push({ role: "user", content: user });
    return {
      model: cfg.model,
      messages: messages,
      max_tokens: cfg.maxTokens || 2048,
      temperature: 0.3,
      stream: false,
    };
  }

  function guessStyle(endpoint) {
    if (/generativelanguage\.googleapis|gemini/i.test(endpoint)) return "google";
    return "openai";
  }

  function extractText(cfg, data) {
    var style = cfg.style || guessStyle(cfg.endpoint);
    try {
      if (style === "google") {
        return (
          data.candidates &&
          data.candidates[0] &&
          data.candidates[0].content &&
          data.candidates[0].content.parts &&
          data.candidates[0].content.parts[0] &&
          data.candidates[0].content.parts[0].text
        ) || "";
      }
      // OpenAI-compatible
      return (
        data.choices &&
        data.choices[0] &&
        data.choices[0].message &&
        data.choices[0].message.content
      ) || "";
    } catch (e) {
      return "";
    }
  }

  /**
   * @param {ProviderConfig} cfg
   */
  function createProvider(cfg) {
    if (!cfg || !cfg.endpoint) {
      throw new Error("provider: endpoint required");
    }

    return {
      id: cfg.id || "default",
      name: cfg.name || cfg.id || "LLM",
      config: cfg,

      isAvailable: function () {
        return !!(cfg.apiKey && cfg.endpoint);
      },

      /**
       * Non-streaming completion. Returns plain text.
       * @param {{ system?: string, user: string, signal?: AbortSignal }} opts
       * @returns {Promise<string>}
       */
      complete: async function (opts) {
        if (!opts || !opts.user) throw new Error("provider.complete: user text required");
        if (!this.isAvailable()) throw new Error("provider: missing apiKey or endpoint");

        var url = cfg.endpoint;
        // Google often wants key as query param
        if ((cfg.style || guessStyle(cfg.endpoint)) === "google" && cfg.apiKey) {
          url += (url.indexOf("?") >= 0 ? "&" : "?") + "key=" + encodeURIComponent(cfg.apiKey);
        }

        var body = buildBody(cfg, opts.system || "", opts.user);
        var res = await fetch(url, {
          method: "POST",
          headers: defaultHeaders(cfg),
          body: JSON.stringify(body),
          signal: opts.signal,
        });

        if (!res.ok) {
          var errText = "";
          try { errText = await res.text(); } catch (_) {}
          var err = new Error("LLM HTTP " + res.status + ": " + errText.slice(0, 400));
          err.status = res.status;
          throw err;
        }

        var data = await res.json();
        var text = extractText(cfg, data);
        if (!text) throw new Error("provider: empty model response");
        return text;
      },
    };
  }

  /**
   * Simple multi-provider registry with ordered failover.
   * @param {ProviderConfig[]} configs
   */
  function createRegistry(configs) {
    var providers = (configs || []).map(createProvider);
    return {
      list: function () { return providers.slice(); },
      get: function (id) {
        for (var i = 0; i < providers.length; i++) {
          if (providers[i].id === id) return providers[i];
        }
        return null;
      },
      /**
       * Try providers in order until one succeeds.
       * Skips 401/403 (bad key) only after trying; retries on 429/5xx.
       */
      completeWithFailover: async function (opts) {
        var lastErr = null;
        for (var i = 0; i < providers.length; i++) {
          var p = providers[i];
          if (!p.isAvailable()) continue;
          try {
            return { text: await p.complete(opts), providerId: p.id, providerName: p.name };
          } catch (e) {
            lastErr = e;
            // continue on rate-limit / server errors
            if (e && (e.status === 429 || (e.status >= 500 && e.status < 600))) continue;
            // for other errors still try next if multiple keys configured
            continue;
          }
        }
        throw lastErr || new Error("No available LLM provider");
      },
    };
  }

  // Default empty registry; app fills from settings / localStorage.
  var registry = createRegistry([]);

  global.ConductinoAI = global.ConductinoAI || {};
  global.ConductinoAI.createProvider = createProvider;
  global.ConductinoAI.createRegistry = createRegistry;
  global.ConductinoAI.registry = registry;
  global.ConductinoAI.setProviders = function (configs) {
    registry = createRegistry(configs);
    global.ConductinoAI.registry = registry;
    return registry;
  };
})(typeof window !== "undefined" ? window : globalThis);
