/**
 * Real-browser navigation helpers (main WebView2 — not iframe).
 */
(function () {
  "use strict";

  function bridge() {
    return window.ConductinoBridge;
  }

  function navigate(url) {
    url = (url || "").trim();
    if (!url) return;
    var b = bridge();
    if (b && b.navigate) {
      return b.navigate(url);
    }
    window.location.href = url;
  }

  /** After returning from a web page via #hash, pull clipboard into Study. */
  async function handleReturnHash() {
    var hash = (location.hash || "").replace(/^#/, "");
    if (hash !== "conductino-study-clipboard" && hash !== "conductino-summarize-clipboard") {
      return;
    }
    history.replaceState(null, "", location.pathname + location.search);

    var text = "";
    try {
      if (navigator.clipboard && navigator.clipboard.readText) {
        text = await navigator.clipboard.readText();
      }
    } catch (e) {
      console.warn("[browser] clipboard", e);
    }
    if (!text || !text.trim()) {
      alert("Clipboard empty. On the web page: select text → use the floating bar again.");
      return;
    }

    if (window.ConductinoStudy && window.ConductinoStudy.addDoc) {
      window.ConductinoStudy.addDoc({
        id: "doc-" + Date.now().toString(36),
        title: "Web selection",
        text: text.trim(),
        sourceType: "web-selection",
        pathOrUrl: "",
      });
    }
    if (window.ConductinoChrome && window.ConductinoChrome.showPanel) {
      window.ConductinoChrome.showPanel("study");
    }
    if (hash === "conductino-summarize-clipboard" && window.ConductinoStudy && window.ConductinoStudy.summarizeNow) {
      window.ConductinoStudy.summarizeNow(false);
    }
  }

  function registerHome() {
    var b = bridge();
    if (b && b.setHomeURL) {
      b.setHomeURL(window.location.origin);
    }
  }

  function init() {
    registerHome();
    // Retry once bindings appear
    setTimeout(registerHome, 500);
    setTimeout(registerHome, 1500);
    handleReturnHash();
    window.addEventListener("hashchange", handleReturnHash);
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init);
  } else {
    init();
  }

  window.ConductinoBrowser = {
    navigate: navigate,
    show: function () {},
    currentUrl: function () {
      return "";
    },
  };
})();
