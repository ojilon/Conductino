/**
 * In-app browser panel — navigate without leaving the shell.
 * Cross-origin pages cannot expose selection; use clipboard or Fetch page text.
 */
(function () {
  "use strict";

  var lastUrl = "";

  function $(id) {
    return document.getElementById(id);
  }

  function frame() {
    return $("browser-frame");
  }

  function status(msg) {
    var s = $("browser-status");
    if (s) s.textContent = msg || "";
  }

  function showBrowserPanel() {
    if (window.ConductinoChrome && window.ConductinoChrome.showPanel) {
      window.ConductinoChrome.showPanel("browser");
    }
  }

  function navigate(url) {
    url = (url || "").trim();
    if (!url) return;
    if (!/^https?:\/\//i.test(url) && !/^about:/i.test(url)) {
      if (/^[a-z0-9.-]+\.[a-z]{2,}/i.test(url) || /^localhost/i.test(url)) {
        url = "https://" + url;
      }
    }
    lastUrl = url;
    showBrowserPanel();
    var f = frame();
    if (!f) return;
    status("Loading… (some sites block embed — use Fetch text or system browser)");
    try {
      f.src = url;
    } catch (e) {
      status("Navigate error: " + e);
    }
    var omni = $("omnibox");
    if (omni) omni.value = url;
  }

  function currentUrl() {
    var f = frame();
    try {
      if (f && f.contentWindow && f.contentWindow.location && f.contentWindow.location.href) {
        var href = f.contentWindow.location.href;
        if (href && href !== "about:blank") return href;
      }
    } catch (_) {}
    return lastUrl || (($("omnibox") || {}).value || "");
  }

  /** Same-origin only — returns "" on cross-origin. */
  function tryFrameSelection() {
    var f = frame();
    if (!f) return "";
    try {
      var sel = f.contentWindow && f.contentWindow.getSelection && f.contentWindow.getSelection();
      if (sel && !sel.isCollapsed) return String(sel);
    } catch (_) {}
    return "";
  }

  async function selectionOrClipboard() {
    var t = tryFrameSelection();
    if (t && t.trim()) return t.trim();
    try {
      if (navigator.clipboard && navigator.clipboard.readText) {
        t = await navigator.clipboard.readText();
        if (t && t.trim()) return t.trim();
      }
    } catch (_) {}
    return "";
  }

  async function sendSelectionToStudy(summarize) {
    var text = await selectionOrClipboard();
    if (!text) {
      status(
        "No selection readable (cross-origin). Select text → Ctrl+C, then click again — or use Fetch page text."
      );
      return;
    }
    if (window.ConductinoStudy && window.ConductinoStudy.addDoc) {
      window.ConductinoStudy.addDoc({
        id: "doc-" + Date.now().toString(36),
        title: "Selection · " + (currentUrl() || "page").slice(0, 40),
        text: text,
        sourceType: "web-selection",
        pathOrUrl: currentUrl(),
      });
    }
    if (window.ConductinoChrome) window.ConductinoChrome.showPanel("study");
    if (summarize && window.ConductinoStudy && window.ConductinoStudy.summarizeNow) {
      window.ConductinoStudy.summarizeNow(false);
    }
    status(summarize ? "Sent to Study + summarize" : "Sent selection to Study");
  }

  async function fetchPageIntoStudy() {
    var url = currentUrl();
    if (!url || !/^https?:/i.test(url)) {
      status("Navigate to a page first");
      return;
    }
    var b = window.ConductinoBridge;
    if (!b || !b.fetchPageText) {
      status("Fetch binding missing");
      return;
    }
    status("Fetching page text…");
    try {
      var text = await b.fetchPageText(url);
      if (!text) {
        status("Empty extract");
        return;
      }
      if (window.ConductinoStudy && window.ConductinoStudy.addDoc) {
        window.ConductinoStudy.addDoc({
          id: "doc-" + Date.now().toString(36),
          title: url.replace(/^https?:\/\//, "").split("/")[0],
          text: text,
          sourceType: "web-fetch",
          pathOrUrl: url,
        });
      }
      if (window.ConductinoChrome) window.ConductinoChrome.showPanel("study");
      status("Page text loaded in Study (" + text.length + " chars)");
    } catch (e) {
      status("Fetch failed: " + (e && e.message ? e.message : e));
    }
  }

  function openExternal() {
    var url = currentUrl();
    if (!url) return;
    var b = window.ConductinoBridge;
    if (b && b.openURL) b.openURL(url);
  }

  function init() {
    var f = frame();
    if (f) {
      f.addEventListener("load", function () {
        status("Loaded (if blank, site blocked iframe — use Fetch page text)");
        try {
          var href = f.contentWindow.location.href;
          if (href && href !== "about:blank") {
            lastUrl = href;
            var omni = $("omnibox");
            if (omni) omni.value = href;
          }
        } catch (_) {
          /* cross-origin — keep lastUrl */
        }
      });
    }

    var toStudy = $("btn-browser-to-study");
    var sum = $("btn-browser-summarize");
    var fetchBtn = $("btn-browser-fetch");
    var ext = $("btn-browser-external");
    if (toStudy) toStudy.addEventListener("click", function () {
      sendSelectionToStudy(false);
    });
    if (sum) sum.addEventListener("click", function () {
      sendSelectionToStudy(true);
    });
    if (fetchBtn) fetchBtn.addEventListener("click", fetchPageIntoStudy);
    if (ext) ext.addEventListener("click", openExternal);
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init);
  } else {
    init();
  }

  window.ConductinoBrowser = {
    navigate: navigate,
    currentUrl: currentUrl,
    show: showBrowserPanel,
  };
})();
