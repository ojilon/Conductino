/**
 * Dual-webview browser helpers — chrome stays; content is a native WebView2 child.
 */
(function () {
  "use strict";

  function bridge() {
    return window.ConductinoBridge;
  }

  function measureChromeHeight() {
    var tab = document.getElementById("tabstrip");
    var tool = document.getElementById("toolbar");
    var h = 0;
    if (tab) h += tab.getBoundingClientRect().height;
    if (tool) h += tool.getBoundingClientRect().height;
    return Math.max(40, Math.round(h));
  }

  async function ensure() {
    var b = bridge();
    if (!b) return;
    try {
      if (b.setHomeURL) await b.setHomeURL(window.location.origin);
      if (b.contentEnsure) await b.contentEnsure();
      if (b.contentSetChromeHeight) await b.contentSetChromeHeight(measureChromeHeight());
    } catch (e) {
      console.warn("[browser] ensure", e);
    }
  }

  function navigate(url) {
    url = (url || "").trim();
    if (!url) return;
    var b = bridge();
    if (b && b.navigate) return b.navigate(url);
    window.location.href = url;
  }

  async function setBrowsingMode(on) {
    var b = bridge();
    if (!b || !b.contentSetVisible) return;
    try {
      await b.contentSetVisible(!!on);
      if (on && b.contentSetChromeHeight) {
        await b.contentSetChromeHeight(measureChromeHeight());
      }
    } catch (e) {
      console.warn("[browser] visible", e);
    }
  }

  async function copySelectionToStudy(summarize) {
    var b = bridge();
    if (b && b.contentCopySelection) {
      try {
        await b.contentCopySelection();
      } catch (_) {}
    }
    // Give clipboard a moment, then read in the chrome webview
    setTimeout(async function () {
      var text = "";
      try {
        if (navigator.clipboard && navigator.clipboard.readText) {
          text = await navigator.clipboard.readText();
        }
      } catch (_) {}
      if (!text || !text.trim()) {
        alert("Select text in the page first, then try again.");
        return;
      }
      await setBrowsingMode(false);
      if (window.ConductinoStudy && window.ConductinoStudy.addDoc) {
        window.ConductinoStudy.addDoc({
          id: "doc-" + Date.now().toString(36),
          title: "Web selection",
          text: text.trim(),
          sourceType: "web-selection",
          pathOrUrl: "",
        });
      }
      if (window.ConductinoChrome) window.ConductinoChrome.showPanel("study");
      if (summarize && window.ConductinoStudy && window.ConductinoStudy.summarizeNow) {
        window.ConductinoStudy.summarizeNow(false);
      }
    }, 200);
  }

  function init() {
    ensure();
    setTimeout(ensure, 800);
    setTimeout(ensure, 2000);

    window.addEventListener("resize", function () {
      var b = bridge();
      if (b && b.contentResize) b.contentResize();
      if (b && b.contentSetChromeHeight) b.contentSetChromeHeight(measureChromeHeight());
    });

    // Selection tools on chrome (work while content pane is visible)
    var btnSum = document.getElementById("btn-chrome-summarize");
    var btnStudy = document.getElementById("btn-chrome-to-study");
    if (btnSum) {
      btnSum.addEventListener("click", function () {
        copySelectionToStudy(true);
      });
    }
    if (btnStudy) {
      btnStudy.addEventListener("click", function () {
        copySelectionToStudy(false);
      });
    }
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init);
  } else {
    init();
  }

  window.ConductinoBrowser = {
    navigate: navigate,
    setBrowsingMode: setBrowsingMode,
    ensure: ensure,
    show: function () {
      setBrowsingMode(true);
    },
  };
})();
