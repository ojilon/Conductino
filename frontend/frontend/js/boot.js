/**
 * Step 1 — verify Wails JS ↔ Go bindings.
 * Full chrome + study UI arrive in later steps.
 */
(function () {
  "use strict";

  var statusEl = document.getElementById("status");
  var btn = document.getElementById("btn-ping");

  function setStatus(msg, ok) {
    if (!statusEl) return;
    statusEl.textContent = msg;
    statusEl.style.borderColor = ok ? "#2f6f4e" : "#2a3340";
  }

  async function callGreet() {
    try {
      // Wails injects window.go.main.App after runtime loads.
      if (!window.go || !window.go.main || !window.go.main.App) {
        setStatus("Go bindings not ready yet — wait a moment and retry.", false);
        return;
      }
      var text = await window.go.main.App.Greet("student");
      setStatus(text, true);
    } catch (e) {
      setStatus("Bridge error: " + (e && e.message ? e.message : String(e)), false);
    }
  }

  async function loadInfo() {
    try {
      if (!window.go || !window.go.main || !window.go.main.App) return;
      var info = await window.go.main.App.AppInfo();
      if (info && info.version) {
        setStatus("Engine " + info.engine + " · v" + info.version + " — click Ping Go", true);
      }
    } catch (_) {}
  }

  if (btn) btn.addEventListener("click", callGreet);

  // Runtime may load slightly after DOM.
  function tryInfo(n) {
    loadInfo();
    if (n < 10) setTimeout(function () { tryInfo(n + 1); }, 200);
  }
  tryInfo(0);
})();
