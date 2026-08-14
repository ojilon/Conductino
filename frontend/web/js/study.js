/**
 * Study workspace controller — split view, open/paste, summarize, transfer.
 */
(function () {
  "use strict";

  var docs = []; // { id, title, text, sourceType, pathOrUrl }
  var activeDocId = null;
  var knowledgeMd = "";

  var el = {};

  function $(id) {
    return document.getElementById(id);
  }

  function status(msg) {
    if (el.status) el.status.textContent = msg || "";
  }

  function AI() {
    return window.ConductinoAI || null;
  }

  function B() {
    return window.ConductinoBridge || null;
  }

  function showStudy() {
    if (window.ConductinoChrome && window.ConductinoChrome.showPanel) {
      window.ConductinoChrome.showPanel("study");
    } else {
      ["welcome", "settings-panel", "stub-panel", "study-panel"].forEach(function (id) {
        var n = $(id);
        if (!n) return;
        var on = id === "study-panel";
        n.classList.toggle("active", on);
        if (on) n.removeAttribute("hidden");
        else n.setAttribute("hidden", "");
      });
    }
  }

  function activeDoc() {
    for (var i = 0; i < docs.length; i++) if (docs[i].id === activeDocId) return docs[i];
    return null;
  }

  function renderDocList() {
    if (!el.docList) return;
    el.docList.innerHTML = "";
    docs.forEach(function (d) {
      var opt = document.createElement("option");
      opt.value = d.id;
      opt.textContent = d.title || d.id;
      if (d.id === activeDocId) opt.selected = true;
      el.docList.appendChild(opt);
    });
  }

  function setActiveDoc(id) {
    activeDocId = id;
    var d = activeDoc();
    if (el.source) {
      el.source.textContent = d ? d.text : "";
    }
    renderDocList();
  }

  function addDoc(doc) {
    docs.push(doc);
    setActiveDoc(doc.id);
    status("Loaded: " + (doc.title || doc.id) + " (" + (doc.text || "").length + " chars)");
  }

  function selectionText() {
    var sel = window.getSelection();
    if (!sel || sel.isCollapsed) return "";
    // Prefer selection inside study-source
    if (el.source && el.source.contains(sel.anchorNode)) {
      return String(sel);
    }
    return String(sel);
  }

  function loadAiConfig() {
    try {
      var raw = localStorage.getItem("conductino.ai");
      if (!raw) return [];
      var parsed = JSON.parse(raw);
      return Array.isArray(parsed) ? parsed : parsed ? [parsed] : [];
    } catch (e) {
      return [];
    }
  }

  function saveAiConfig(list) {
    localStorage.setItem("conductino.ai", JSON.stringify(list));
    var ai = AI();
    if (ai && ai.setProviders) ai.setProviders(list);
  }

  function presetEndpoint(name) {
    if (name === "google") {
      return {
        id: "google",
        name: "Google AI Studio",
        endpoint:
          "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent",
        model: "gemini-2.0-flash",
        style: "google",
      };
    }
    if (name === "openrouter") {
      return {
        id: "openrouter",
        name: "OpenRouter",
        endpoint: "https://openrouter.ai/api/v1/chat/completions",
        model: "google/gemini-2.0-flash-001",
        style: "openai",
      };
    }
    if (name === "groq") {
      return {
        id: "groq",
        name: "Groq",
        endpoint: "https://api.groq.com/openai/v1/chat/completions",
        model: "llama-3.3-70b-versatile",
        style: "openai",
      };
    }
    return {
      id: "custom",
      name: "Custom",
      endpoint: "",
      model: "",
      style: "openai",
    };
  }

  async function summarizeSelectionOrDoc() {
    var ai = AI();
    if (!ai) {
      status("AI modules not loaded");
      return;
    }
    var configs = loadAiConfig();
    if (!configs.length) {
      status("Add an API key in Settings → AI providers");
      return;
    }
    ai.setProviders(configs);

    var text = selectionText();
    var doc = activeDoc();
    if (!text && doc) text = doc.text;
    if (!text || !text.trim()) {
      status("Nothing to summarize — select text or open a document");
      return;
    }

    var chunks = ai.chunkText(text, {
      sourceId: doc ? doc.id : "sel",
      targetTokens: 1600,
      maxTokens: 2200,
    });
    var windows = ai.windowsFor(chunks, 6000, 500);
    status("Summarizing " + chunks.length + " chunk(s) in " + windows.length + " window(s)…");

    var system =
      "You are an academic research assistant. Summarise clearly for a university student. " +
      "Preserve technical terms and numbers. Do not invent facts. Use short paragraphs. " +
      "Mark key claims with [1], [2] when useful.";

    for (var w = 0; w < windows.length; w++) {
      var payload = ai.assemblePayload(windows[w], {
        instruction: "Summarise the chunks above.",
      });
      try {
        var result = await ai.registry.completeWithFailover({
          system: system,
          user: payload,
        });
        var pageRange = null;
        if (windows[w].length) {
          var pages = windows[w].map(function (c) {
            return c.approxPage;
          });
          pageRange = [Math.min.apply(null, pages), Math.max.apply(null, pages)];
        }
        var block = ai.makeBlock(result.text, {
          type: "summary",
          sourceId: doc ? doc.id : "sel",
          chunkIds: windows[w].map(function (c) {
            return c.id;
          }),
          pageRange: pageRange,
          model: (configs[0] && configs[0].model) || "",
          provider: result.providerName || result.providerId,
        });
        var sourceMap = {};
        if (doc) {
          sourceMap[doc.id] = {
            title: doc.title,
            pathOrUrl: doc.pathOrUrl || "",
            id: doc.id,
          };
        }
        ai.insertBlockIntoDOM(el.knowledge, block, {
          appendCitations: true,
          sourceMap: sourceMap,
        });
        knowledgeMd = ai.appendBlockToMarkdown(knowledgeMd, block, sourceMap);
        status("Inserted summary from " + result.providerName + " (window " + (w + 1) + "/" + windows.length + ")");
      } catch (e) {
        status("LLM error: " + (e && e.message ? e.message : String(e)));
        return;
      }
    }
  }

  function transferExact() {
    var ai = AI();
    var text = selectionText();
    if (!text.trim()) {
      status("Select text in the source pane first");
      return;
    }
    var doc = activeDoc();
    var block = ai
      ? ai.makeBlock(text.trim(), {
          type: "exact",
          sourceId: doc ? doc.id : "sel",
          provider: "local",
          model: "exact-transfer",
        })
      : {
          id: "block-" + Date.now(),
          type: "exact",
          text: text.trim(),
          sourceId: doc ? doc.id : "",
          createdAt: new Date().toISOString(),
          citations: [],
        };
    if (ai) {
      var sourceMap = {};
      if (doc) sourceMap[doc.id] = { title: doc.title, pathOrUrl: doc.pathOrUrl || "", id: doc.id };
      ai.insertBlockIntoDOM(el.knowledge, block, { appendCitations: false, sourceMap: sourceMap });
      knowledgeMd = ai.appendBlockToMarkdown(knowledgeMd, block, sourceMap);
    } else {
      var pre = document.createElement("pre");
      pre.textContent = text.trim();
      el.knowledge.appendChild(pre);
    }
    status("Transferred exact selection");
  }

  function showChunkInfo() {
    var ai = AI();
    var doc = activeDoc();
    if (!ai || !doc) {
      status("Open a document first");
      return;
    }
    var chunks = ai.chunkText(doc.text, { sourceId: doc.id });
    status(
      chunks.length +
        " chunks · ~" +
        chunks.reduce(function (a, c) {
          return a + (c.tokenEstimate || 0);
        }, 0) +
        " tokens total"
    );
  }

  async function openFile() {
    var bridge = B();
    var path = "";
    if (bridge && typeof bridge.openFile === "function") {
      path = await Promise.resolve(bridge.openFile());
    } else if (window.hostOpenFile) {
      path = await Promise.resolve(window.hostOpenFile());
    }
    if (!path) {
      // Fallback: paste path manually
      path = window.prompt("File path (native dialog not bound yet). Or use Paste text.", "");
    }
    if (!path) return;

    // Prefer native extract when available
    var text = "";
    if (bridge && typeof bridge.extractDocument === "function") {
      try {
        text = await Promise.resolve(bridge.extractDocument(path));
      } catch (e) {
        status("Extract failed: " + e);
      }
    }
    if (!text) {
      // Last resort: user pastes after opening externally
      status("Path recorded. Native extract not available — use Paste text for now.");
      addDoc({
        id: "doc-" + Date.now().toString(36),
        title: path.split(/[/\\]/).pop() || path,
        text: "",
        sourceType: "external",
        pathOrUrl: path,
      });
      return;
    }
    addDoc({
      id: "doc-" + Date.now().toString(36),
      title: path.split(/[/\\]/).pop() || path,
      text: text,
      sourceType: "external",
      pathOrUrl: path,
    });
  }

  function pasteText() {
    var t = window.prompt("Paste document text");
    if (!t) return;
    addDoc({
      id: "doc-" + Date.now().toString(36),
      title: "Pasted text",
      text: t,
      sourceType: "paste",
      pathOrUrl: "",
    });
  }

  function exportMd() {
    var blob = new Blob([knowledgeMd || el.knowledge.innerText || ""], {
      type: "text/markdown;charset=utf-8",
    });
    var a = document.createElement("a");
    a.href = URL.createObjectURL(blob);
    a.download = "conductino-knowledge-" + new Date().toISOString().slice(0, 10) + ".md";
    a.click();
    URL.revokeObjectURL(a.href);
    status("Export started");
  }

  function clearKnowledge() {
    if (el.knowledge) el.knowledge.innerHTML = "";
    knowledgeMd = "";
    status("Knowledge pane cleared");
  }

  function bindSettingsAi() {
    var btn = $("btn-ai-save");
    if (!btn) return;
    btn.addEventListener("click", function () {
      var preset = ($("setting-ai-preset") || {}).value || "google";
      var base = presetEndpoint(preset);
      var key = (($("setting-ai-key") || {}).value || "").trim();
      var endpoint = (($("setting-ai-endpoint") || {}).value || "").trim();
      var model = (($("setting-ai-model") || {}).value || "").trim();
      if (endpoint) base.endpoint = endpoint;
      if (model) base.model = model;
      base.apiKey = key;
      if (!base.apiKey || !base.endpoint) {
        alert("API key and endpoint required");
        return;
      }
      var list = loadAiConfig().filter(function (c) {
        return c.id !== base.id;
      });
      list.unshift(base);
      saveAiConfig(list);
      status("AI config saved: " + base.name);
      alert("Saved " + base.name);
    });
  }

  function init() {
    el.source = $("study-source");
    el.knowledge = $("study-knowledge");
    el.docList = $("study-doc-list");
    el.status = $("study-status");

    var openStudy = function () {
      showStudy();
    };
    ["btn-study", "btn-open-study"].forEach(function (id) {
      var n = $(id);
      if (n) n.addEventListener("click", openStudy);
    });

    document.querySelectorAll('.sidebar-item[data-action="study"]').forEach(function (btn) {
      btn.addEventListener("click", function () {
        openStudy();
        var side = $("sidebar");
        if (side) {
          side.setAttribute("hidden", "");
          side.classList.remove("open");
        }
      });
    });

    if ($("btn-doc-open")) $("btn-doc-open").addEventListener("click", openFile);
    if ($("btn-doc-paste")) $("btn-doc-paste").addEventListener("click", pasteText);
    if ($("btn-transfer-exact")) $("btn-transfer-exact").addEventListener("click", transferExact);
    if ($("btn-summarize")) $("btn-summarize").addEventListener("click", summarizeSelectionOrDoc);
    if ($("btn-chunk-info")) $("btn-chunk-info").addEventListener("click", showChunkInfo);
    if ($("btn-knowledge-clear")) $("btn-knowledge-clear").addEventListener("click", clearKnowledge);
    if ($("btn-knowledge-export")) $("btn-knowledge-export").addEventListener("click", exportMd);
    if (el.docList) {
      el.docList.addEventListener("change", function () {
        setActiveDoc(el.docList.value);
      });
    }

    bindSettingsAi();
    var configs = loadAiConfig();
    var ai = AI();
    if (ai && configs.length) ai.setProviders(configs);
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init);
  } else {
    init();
  }

  // Expose for app.js panel routing
  window.ConductinoStudy = {
    show: showStudy,
    addDoc: addDoc,
  };
})();
