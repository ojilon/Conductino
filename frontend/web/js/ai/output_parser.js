/**
 * Parse LLM text output: citations, image placeholders, insert into DOM or markdown.
 */
(function (global) {
  "use strict";

  var CITE_RE = /\[(\d+)\]/g;
  var IMG_RE = /\[\[IMG:([a-zA-Z0-9_-]+)\]\]/g;

  function extractMarkers(text) {
    var citations = [];
    var images = [];
    var seenCite = {};
    var seenImg = {};
    var m;
    CITE_RE.lastIndex = 0;
    while ((m = CITE_RE.exec(text)) !== null) {
      var n = m[1];
      if (!seenCite[n]) {
        seenCite[n] = true;
        citations.push({ label: "[" + n + "]", number: parseInt(n, 10) });
      }
    }
    IMG_RE.lastIndex = 0;
    while ((m = IMG_RE.exec(text)) !== null) {
      var id = m[1];
      if (!seenImg[id]) {
        seenImg[id] = true;
        images.push({ id: id, placeholder: m[0] });
      }
    }
    return { citations: citations, images: images };
  }

  function buildCitationList(markers, sourceMap, blockMeta) {
    sourceMap = sourceMap || {};
    blockMeta = blockMeta || {};
    var entries = [];
    var list = (markers && markers.citations) || [];
    if (!list.length && blockMeta.sourceId) {
      list = [{ label: "[1]", number: 1 }];
    }
    list.forEach(function (c, idx) {
      var src = sourceMap[blockMeta.sourceId] || sourceMap[Object.keys(sourceMap)[0]] || {};
      entries.push({
        id: "cite-" + (c.number || idx + 1),
        label: c.label || ("[" + (idx + 1) + "]"),
        title: src.title || "Untitled source",
        sourceId: blockMeta.sourceId || src.id || "",
        pathOrUrl: src.pathOrUrl || "",
        pages: (blockMeta.pageRange && blockMeta.pageRange.join("-")) || src.pages || "",
      });
    });
    return { entries: entries };
  }

  function citationsToMarkdown(citeList) {
    if (!citeList || !citeList.entries || !citeList.entries.length) return "";
    var lines = ["", "## References", ""];
    citeList.entries.forEach(function (e) {
      var pageBit = e.pages ? ", pp. " + e.pages : "";
      lines.push(e.label + " " + e.title + pageBit + (e.pathOrUrl ? " — " + e.pathOrUrl : ""));
    });
    return lines.join("\n");
  }

  function makeBlock(text, meta) {
    meta = meta || {};
    var markers = extractMarkers(text);
    return {
      id: meta.id || ("block-" + Date.now().toString(36)),
      type: meta.type || "summary",
      text: text,
      sourceId: meta.sourceId || "",
      chunkIds: meta.chunkIds || [],
      pageRange: meta.pageRange || null,
      model: meta.model || "",
      provider: meta.provider || "",
      createdAt: meta.createdAt || new Date().toISOString(),
      citations: markers.citations,
      images: markers.images,
      verified: false,
      flags: [],
    };
  }

  function insertBlockIntoDOM(container, block, opts) {
    if (!container || !block) return null;
    opts = opts || {};

    var article = document.createElement("article");
    article.className = "knowledge-block";
    article.dataset.blockId = block.id;
    article.dataset.sourceId = block.sourceId || "";
    if (block.pageRange) article.dataset.pages = block.pageRange.join("-");

    var header = document.createElement("header");
    header.className = "knowledge-block-meta";
    var metaBits = [];
    if (block.type) metaBits.push(block.type);
    if (block.provider) metaBits.push(block.provider);
    if (block.model) metaBits.push(block.model);
    if (block.pageRange) metaBits.push("pp. " + block.pageRange.join("-"))
    if (block.createdAt) metaBits.push(block.createdAt.slice(0, 19).replace("T", " "));
    header.textContent = metaBits.join(" · ");

    var body = document.createElement("div");
    body.className = "knowledge-block-body";
    body.setAttribute("contenteditable", opts.editable === false ? "false" : "true");
    body.style.whiteSpace = "pre-wrap";
    body.textContent = block.text;

    article.appendChild(header);
    article.appendChild(body);

    if (opts.appendCitations && block.citations && block.citations.length) {
      var citeList = buildCitationList({ citations: block.citations }, opts.sourceMap || {}, block);
      var ref = document.createElement("footer");
      ref.className = "knowledge-block-refs";
      ref.style.whiteSpace = "pre-wrap";
      ref.textContent = citationsToMarkdown(citeList).trim();
      article.appendChild(ref);
    }

    if (opts.prepend) container.insertBefore(article, container.firstChild);
    else container.appendChild(article);
    return article;
  }

  function appendBlockToMarkdown(existingMd, block, sourceMap) {
    var md = existingMd || "";
    var markers = extractMarkers(block.text);
    var citeList = buildCitationList(markers, sourceMap || {}, block);
    var header =
      "\n\n<!-- block:" + block.id + " source:" + (block.sourceId || "") + " -->\n\n";
    var body = block.text.trim() + "\n";
    var refs = citationsToMarkdown(citeList);
    return md + header + body + (refs ? refs + "\n" : "");
  }

  global.ConductinoAI = global.ConductinoAI || {};
  global.ConductinoAI.extractMarkers = extractMarkers;
  global.ConductinoAI.buildCitationList = buildCitationList;
  global.ConductinoAI.citationsToMarkdown = citationsToMarkdown;
  global.ConductinoAI.makeBlock = makeBlock;
  global.ConductinoAI.insertBlockIntoDOM = insertBlockIntoDOM;
  global.ConductinoAI.appendBlockToMarkdown = appendBlockToMarkdown;
})(typeof window !== "undefined" ? window : globalThis);
