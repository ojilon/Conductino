/**
 * Study workspace — paste/transfer + resizable split.
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

  function transferExact() {
    var text = selectionText();
    if (!text.trim()) {
      status("Select text in the source pane first");
      return;
    }
    var kn = $("study-knowledge");
    if (!kn) return;
    var article = document.createElement("article");
    article.className = "knowledge-block";
    var header = document.createElement("header");
    header.className = "knowledge-block-meta";
    header.textContent = "exact · local · " + new Date().toISOString().slice(0, 19).replace("T", " ");
    var body = document.createElement("div");
    body.className = "knowledge-block-body";
    body.setAttribute("contenteditable", "true");
    body.style.whiteSpace = "pre-wrap";
    body.textContent = text.trim();
    article.appendChild(header);
    article.appendChild(body);
    kn.appendChild(article);
    knowledgeMd += "\n\n" + text.trim() + "\n";
    status("Transferred exact selection");
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
    var path = "";
    if (b && b.openFile) path = await b.openFile();
    if (!path) {
      path = window.prompt("File path (native dialog in Step 3). Or use Paste text.", "") || "";
    }
    if (!path) return;
    var text = "";
    if (b && b.extractDocument) text = await b.extractDocument(path);
    if (!text) {
      status("Path noted. Extract lands in Step 3 — use Paste text for now.");
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

  /** Drag the middle bar to resize left/right study panes. */
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

    // Touch support
    splitter.addEventListener(
      "touchstart",
      function (e) {
        dragging = true;
        splitter.classList.add("dragging");
      },
      { passive: true }
    );
    window.addEventListener(
      "touchmove",
      function (e) {
        if (!dragging || !e.touches[0]) return;
        onMove(e.touches[0].clientX, e.touches[0].clientY);
      },
      { passive: true }
    );
    window.addEventListener("touchend", function () {
      dragging = false;
      splitter.classList.remove("dragging");
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
    if (summarize) {
      summarize.addEventListener("click", function () {
        status("Summarize wires in Step 4 (AI modules).");
      });
    }
    if (chunk) {
      chunk.addEventListener("click", function () {
        var d = activeDoc();
        if (!d || !d.text) {
          status("Open or paste a document first");
          return;
        }
        status("~" + Math.ceil(d.text.length / 4) + " tokens (rough) · full chunker in Step 4");
      });
    }
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
