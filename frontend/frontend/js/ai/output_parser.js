/** Parse LLM output: citations, blocks, DOM insert. */
(function (global) {
  "use strict";

  var CITE_RE = /\[(\d+)\]/g;

  function extractMarkers(text) {
    var citations = [];
    var seen = {};
    var m;
    CITE_RE.lastIndex = 0;
    while ((m = CITE_RE.exec(text)) !== null) {
      if (!seen[m[1]]) {
        seen[m[1]] = true;
        citations.push({ label: "[" + m[1] + "]", number: parseInt(m[1], 10) });
      }
    }
    return { citations: citations };
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

    var header = document.createElement("header");
    header.className = "knowledge-block-meta";
    var bits = [];
    if (block.type) bits.push(block.type);
    if (block.provider) bits.push(block.provider);
    if (block.model) bits.push(block.model);
    if (block.pageRange) bits.push("pp. " + block.pageRange.join("-"))
    if (block.createdAt) bits.push(block.createdAt.slice(0, 19).replace("T", " "));
    header.textContent = bits.join(" · ");

    var body = document.createElement("div");
    body.className = "knowledge-block-body";
    body.setAttribute("contenteditable", opts.editable === false ? "false" : "true");
    body.style.whiteSpace = "pre-wrap";
    body.textContent = block.text;

    article.appendChild(header);
    article.appendChild(body);

    if (opts.appendCitations && block.citations && block.citations.length) {
      var ref = document.createElement("footer");
      ref.className = "knowledge-block-refs";
      ref.style.whiteSpace = "pre-wrap";
      var lines = ["References"];
      block.citations.forEach(function (c) {
        lines.push(c.label + " source " + (block.sourceId || ""));
      });
      ref.textContent = lines.join("\n");
      article.appendChild(ref);
    }

    container.appendChild(article);
    return article;
  }

  function appendBlockToMarkdown(existingMd, block) {
    var md = existingMd || "";
    return md + "\n\n<!-- block:" + block.id + " -->\n\n" + (block.text || "").trim() + "\n";
  }

  global.ConductinoAI = global.ConductinoAI || {};
  global.ConductinoAI.extractMarkers = extractMarkers;
  global.ConductinoAI.makeBlock = makeBlock;
  global.ConductinoAI.insertBlockIntoDOM = insertBlockIntoDOM;
  global.ConductinoAI.appendBlockToMarkdown = appendBlockToMarkdown;
})(typeof window !== "undefined" ? window : globalThis);
