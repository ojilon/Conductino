/**
 * Unified LLM provider adapter (API-agnostic).
 */
(function (global) {
  "use strict";

  function defaultHeaders(cfg) {
    var h = Object.assign({ "Content-Type": "application/json" }, cfg.headers || {});
    if (cfg.style === "openai" || /openrouter|groq|openai/i.test(cfg.endpoint || "")) {
      h["Authorization"] = "Bearer " + cfg.apiKey;
    }
    return h;
  }

  function guessStyle(endpoint) {
    if (/generativelanguage\.googleapis|gemini/i.test(endpoint || "")) return "google";
    return "openai";
  }

  function buildBody(cfg, system, user) {
    var style = cfg.style || guessStyle(cfg.endpoint);
    if (style === "google") {
      var text = system ? system + "\n\n" + user : user;
      return {
        contents: [{ role: "user", parts: [{ text: text }] }],
        generationConfig: { maxOutputTokens: cfg.maxTokens || 2048, temperature: 0.3 },
      };
    }
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

  function extractText(cfg, data) {
    var style = cfg.style || guessStyle(cfg.endpoint);
    try {
      if (style === "google") {
        return (
          (data.candidates &&
            data.candidates[0] &&
            data.candidates[0].content &&
            data.candidates[0].content.parts &&
            data.candidates[0].content.parts[0] &&
            data.candidates[0].content.parts[0].text) ||
          ""
        );
      }
      return (
        (data.choices && data.choices[0] && data.choices[0].message && data.choices[0].message.content) ||
        ""
      );
    } catch (e) {
      return "";
    }
  }

  function createProvider(cfg) {
    if (!cfg || !cfg.endpoint) throw new Error("provider: endpoint required");
    return {
      id: cfg.id || "default",
      name: cfg.name || cfg.id || "LLM",
      config: cfg,
      isAvailable: function () {
        return !!(cfg.apiKey && cfg.endpoint);
      },
      complete: async function (opts) {
        if (!opts || !opts.user) throw new Error("user text required");
        if (!this.isAvailable()) throw new Error("missing apiKey or endpoint");
        var url = cfg.endpoint;
        if ((cfg.style || guessStyle(cfg.endpoint)) === "google" && cfg.apiKey) {
          url += (url.indexOf("?") >= 0 ? "&" : "?") + "key=" + encodeURIComponent(cfg.apiKey);
        }
        var res = await fetch(url, {
          method: "POST",
          headers: defaultHeaders(cfg),
          body: JSON.stringify(buildBody(cfg, opts.system || "", opts.user)),
          signal: opts.signal,
        });
        if (!res.ok) {
          var errText = "";
          try {
            errText = await res.text();
          } catch (_) {}
          var err = new Error("LLM HTTP " + res.status + ": " + errText.slice(0, 400));
          err.status = res.status;
          throw err;
        }
        var data = await res.json();
        var text = extractText(cfg, data);
        if (!text) throw new Error("empty model response");
        return text;
      },
    };
  }

  function createRegistry(configs) {
    var providers = (configs || []).map(createProvider);
    return {
      list: function () {
        return providers.slice();
      },
      completeWithFailover: async function (opts) {
        var lastErr = null;
        for (var i = 0; i < providers.length; i++) {
          var p = providers[i];
          if (!p.isAvailable()) continue;
          try {
            return { text: await p.complete(opts), providerId: p.id, providerName: p.name };
          } catch (e) {
            lastErr = e;
            continue;
          }
        }
        throw lastErr || new Error("No available LLM provider");
      },
    };
  }

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
