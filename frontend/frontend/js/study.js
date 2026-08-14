/**
 * Study workspace — file open, paste, transfer, summarize, resizable split.
 */
(function () {
  "use strict";

  var docs = [];
  var activeDocId = null;
  var knowledgeMd = "";

  function $(id) {
    return document.getElementById(id);
  }

  function status(msg) {
    var s = $("study-status");
    if (s) s.textContent = msg || "";
  }

  function AI() {
    return window.ConductinoAI || null;
  }

  function activeDoc() {
    for (var i = 0; i < docs.length; i++) if (docs[i].id === activeDocId) return docs[i];
    return null;
  }

  function renderDocList() {
    var list = $("study-doc-list");
    if (!list) return;
    list.innerHTML = "";
    docs.forEach(function (d) {
      var opt = document.createElement("option");
      opt.value = d.id;
      opt.textContent = d.title || d.id;
      if (d.id === activeDocId) opt.selected = true;
      list.appendChild(opt);
    });
  }

  function setActiveDoc(id) {
    activeDocId = id;
    var d = activeDoc();
    var src = $("study-source");
    if (src) src.textContent = d ? d.text : "";
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
    var src = $("study-source");
    if (src && src.contains(sel.anchorNode)) return String(sel);
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

  function transferExact() {
    var ai = AI();
    var text = selectionText();
    if (!text.trim()) {
      status("Select text in the source pane first");
      return;
    }
    var kn = $("study-knowledge");
    if (!kn) return;
    var doc = activeDoc();
    if (ai && ai.makeBlock && ai.insertBlockIntoDOM) {
      var block = ai.makeBlock(text.trim(), {
        type: "exact",
        sourceId: doc ? doc.id : "sel",
        provider: "local",
        model: "exact-transfer",
      });
      ai.insertBlockIntoDOM(kn, block, { appendCitations: false });
      knowledgeMd = ai.appendBlockToMarkdown(knowledgeMd, block);
    } else {
      var article = document.createElement("article");
      article.className = "knowledge-block";
      var body = document.createElement("div");
      body.className = "knowledge-block-body";
      body.setAttribute("contenteditable", "true");
      body.style.whiteSpace = "pre-wrap";
      body.textContent = text.trim();
      article.appendChild(body);
      kn.appendChild(article);
      knowledgeMd += "\n\n" + text.trim() + "\n";
    }
    status("Transferred exact selection");
  }

  async function summarizeSelectionOrDoc() {
    var ai = AI();
    if (!ai || !ai.chunkText || !ai.setProviders) {
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

    var kn = $("study-knowledge");
    for (var w = 0; w < windows.length; w++) {
      var payload = ai.assemblePayload(windows[w], { instruction: "Summarise the chunks above." });
      try {
        var result = await ai.registry.completeWithFailover({ system: system, user: payload });
        var pages = windows[w].map(function (c) { return c.approxPage; });
        var pageRange = pages.length ? [Math.min.apply(null, pages), Math.max.apply(null, pages)] : null;
        var block = ai.makeBlock(result.text, {
          type: "summary",
          sourceId: doc ? doc.id : "sel",
          chunkIds: windows[w].map(function (c) { return c.id; }),
          pageRange: pageRange,
          model: (configs[0] && configs[0].model) || "",
          provider: result.providerName || result.providerId,
        });
        if (kn) ai.insertBlockIntoDOM(kn, block, { appendCitations: true });
        knowledgeMd = ai.appendBlockToMarkdown(knowledgeMd, block);
        status(
          "Inserted summary from " +
            result.providerName +
            " (window " +
            (w + 1) +
            "/" +
            windows.length +
            ")"
        );
      } catch (e) {
        status("LLM error: " + (e && e.message ? e.message : String(e)));
        return;
      }
    }
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

  async function openFile() {
    var b = window.ConductinoBridge;
    if (!b) {
      status("Bridge not ready");
      return;
    }
    status("Opening file dialog…");
    var path = "";
    try {
      path = await b.openFile();
    } catch (e) {
      status("Open failed: " + e);
      return;
    }
    if (!path) {
      status("Cancelled");
      return;
    }

    var mode = window.confirm(
      "OK = Import (copy into app data)\nCancel = Link external path only\n\n" + path
    );
    var storePath = path;
    if (mode && b.importDocument) {
      try {
        var imported = await b.importDocument(path);
        if (imported) storePath = imported;
      } catch (_) {}
    }

    status("Extracting…");
    var text = "";
    try {
      text = await b.extractDocument(path);
    } catch (e) {
      status("Extract error: " + e);
      return;
    }
    if (!text) {
      status("No text extracted — try Paste text for PDF/DOCX");
      addDoc({
        id: "doc-" + Date.now().toString(36),
        title: path.split(/[/\\]/).pop() || path,
        text: "",
        sourceType: mode ? "import" : "external",
        pathOrUrl: storePath,
      });
      return;
    }
    addDoc({
      id: "doc-" + Date.now().toString(36),
      title: path.split(/[/\\]/).pop() || path,
      text: text,
      sourceType: mode ? "import" : "external",
      pathOrUrl: storePath,
    });
  }

  function exportMd() {
    var kn = $("study-knowledge");
    var blob = new Blob([knowledgeMd || (kn && kn.innerText) || ""], {
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
    var kn = $("study-knowledge");
    if (kn) kn.innerHTML = "";
    knowledgeMd = "";
    status("Knowledge pane cleared");
  }

  function showChunkInfo() {
    var ai = AI();
    var doc = activeDoc();
    if (!doc || !doc.text) {
      status("Open or paste a document first");
      return;
    }
    if (ai && ai.chunkText) {
      var chunks = ai.chunkText(doc.text, { sourceId: doc.id });
      var total = chunks.reduce(function (a, c) {
        return a + (c.tokenEstimate || 0);
      }, 0);
      status(chunks.length + " chunks · ~" + total + " tokens");
    } else {
      status("~" + Math.ceil(doc.text.length / 4) + " tokens (rough)");
    }
  }

  function initSplitter() {
    var layout = $("study-layout");
    var left = $("study-left");
    var splitter = $("study-splitter");
    if (!layout || !left || !splitter) return;
    var dragging = false;

    function onMove(clientX, clientY) {
      if (!dragging) return;
      var rect = layout.getBoundingClientRect();
      var vertical = window.matchMedia("(max-width: 700px)").matches;
      if (vertical) {
        var y = clientY - rect.top;
        var pct = Math.min(80, Math.max(20, (y / rect.height) * 100));
        left.style.flex = "0 0 " + pct + "%";
        left.style.height = pct + "%";
        left.style.width = "100%";
      } else {
        var x = clientX - rect.left;
        var pctX = Math.min(80, Math.max(20, (x / rect.width) * 100));
        left.style.flex = "0 0 " + pctX + "%";
        left.style.width = pctX + "%";
        left.style.height = "";
      }
    }

    splitter.addEventListener("mousedown", function (e) {
      e.preventDefault();
      dragging = true;
      splitter.classList.add("dragging");
      document.body.style.cursor = window.matchMedia("(max-width: 700px)").matches
        ? "row-resize"
        : "col-resize";
      document.body.style.userSelect = "none";
    });
    window.addEventListener("mousemove", function (e) {
      onMove(e.clientX, e.clientY);
    });
    window.addEventListener("mouseup", function () {
      if (!dragging) return;
      dragging = false;
      splitter.classList.remove("dragging");
      document.body.style.cursor = "";
      document.body.style.userSelect = "";
    });
  }

  function init() {
    var open = $("btn-doc-open");
    var paste = $("btn-doc-paste");
    var transfer = $("btn-transfer-exact");
    var summarize = $("btn-summarize");
    var chunk = $("btn-chunk-info");
    var clear = $("btn-knowledge-clear");
    var exp = $("btn-knowledge-export");
    var list = $("study-doc-list");

    if (open) open.addEventListener("click", openFile);
    if (paste) paste.addEventListener("click", pasteText);
    if (transfer) transfer.addEventListener("click", transferExact);
    if (summarize) summarize.addEventListener("click", summarizeSelectionOrDoc);
    if (chunk) chunk.addEventListener("click", showChunkInfo);
    if (clear) clear.addEventListener("click", clearKnowledge);
    if (exp) exp.addEventListener("click", exportMd);
    if (list) {
      list.addEventListener("change", function () {
        setActiveDoc(list.value);
      });
    }
    initSplitter();
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init);
  } else {
    init();
  }

  window.ConductinoStudy = { addDoc: addDoc };
})();
