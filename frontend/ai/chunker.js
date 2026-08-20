/**
 * Local text chunking + payload assembly.
 * Keeps LLM context windows under free-tier limits.
 */
(function (global) {
  "use strict";

  /** Rough token estimate: ~4 chars per token for English technical text. */
  function estimateTokens(text) {
    if (!text) return 0;
    return Math.ceil(String(text).length / 4);
  }

  /**
   * Split on paragraph boundaries, then pack into chunks near targetTokens.
   * @param {string} text
   * @param {object} [opts]
   * @param {number} [opts.targetTokens=1800]
   * @param {number} [opts.maxTokens=2400]
   * @param {string} [opts.sourceId]
   * @param {number} [opts.charsPerPage=3000]  // rough page estimate
   * @returns {Array<object>} Chunk objects (see docs/ai/ARCHITECTURE.md)
   */
  function chunkText(text, opts) {
    opts = opts || {};
    var target = opts.targetTokens || 1800;
    var maxT = opts.maxTokens || 2400;
    var sourceId = opts.sourceId || "doc-unknown";
    var charsPerPage = opts.charsPerPage || 3000;

    var raw = String(text || "").replace(/\r\n/g, "\n");
    // Prefer double-newline paragraphs; fall back to single newlines / sentences.
    var paras = raw.split(/\n\s*\n/).map(function (p) { return p.trim(); }).filter(Boolean);
    if (paras.length <= 1) {
      paras = raw.split(/\n/).map(function (p) { return p.trim(); }).filter(Boolean);
    }
    if (paras.length === 0 && raw.trim()) paras = [raw.trim()];

    var chunks = [];
    var buf = [];
    var bufTokens = 0;
    var charPos = 0;
    var index = 0;

    function flush() {
      if (!buf.length) return;
      var body = buf.join("\n\n");
      var start = charPos - body.length;
      if (start < 0) start = 0;
      chunks.push({
        id: "chunk-" + String(index).padStart(5, "0"),
        sourceId: sourceId,
        index: index,
        text: body,
        approxPage: Math.floor(start / charsPerPage) + 1,
        headingPath: [],
        tokenEstimate: estimateTokens(body),
        charStart: start,
        charEnd: charPos,
      });
      index += 1;
      buf = [];
      bufTokens = 0;
    }

    for (var i = 0; i < paras.length; i++) {
      var p = paras[i];
      var t = estimateTokens(p);
      if (bufTokens + t > maxT && buf.length) flush();
      if (t > maxT) {
        // Hard-split oversized paragraph by character windows.
        var step = Math.floor(maxT * 4 * 0.9);
        for (var s = 0; s < p.length; s += step) {
          var slice = p.slice(s, s + step);
          buf.push(slice);
          bufTokens += estimateTokens(slice);
          charPos += slice.length + 2;
          flush();
        }
        continue;
      }
      buf.push(p);
      bufTokens += t;
      charPos += p.length + 2;
      if (bufTokens >= target) flush();
    }
    flush();
    return chunks;
  }

  /**
   * Build a single user payload from one or more chunks + optional image alt-text.
   */
  function assemblePayload(chunks, extras) {
    extras = extras || {};
    var parts = [];
    if (extras.preamble) parts.push(extras.preamble);

    (chunks || []).forEach(function (c) {
      parts.push(
        "---\nChunk " + c.index + " (approx page " + c.approxPage + ")\n" + c.text
      );
    });

    if (extras.altTexts && extras.altTexts.length) {
      parts.push("\n--- Image descriptions ---");
      extras.altTexts.forEach(function (img) {
        parts.push(img.placeholder + " " + img.altText);
      });
    }

    if (extras.instruction) parts.push("\n---\n" + extras.instruction);
    return parts.join("\n\n");
  }

  /**
   * Pack chunks into windows that each stay under maxTokens (including system overhead).
   */
  function windowsFor(chunks, maxTokens, systemOverhead) {
    maxTokens = maxTokens || 6000;
    systemOverhead = systemOverhead || 400;
    var budget = maxTokens - systemOverhead;
    var windows = [];
    var cur = [];
    var curT = 0;
    (chunks || []).forEach(function (c) {
      var t = c.tokenEstimate || estimateTokens(c.text);
      if (curT + t > budget && cur.length) {
        windows.push(cur);
        cur = [];
        curT = 0;
      }
      cur.push(c);
      curT += t;
    });
    if (cur.length) windows.push(cur);
    return windows;
  }

  global.ConductinoAI = global.ConductinoAI || {};
  global.ConductinoAI.chunkText = chunkText;
  global.ConductinoAI.estimateTokens = estimateTokens;
  global.ConductinoAI.assemblePayload = assemblePayload;
  global.ConductinoAI.windowsFor = windowsFor;
})(typeof window !== "undefined" ? window : globalThis);
