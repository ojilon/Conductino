/**
 * Thin helpers around Wails runtime + App bindings.
 * window.go.main.App.* is injected by Wails after startup.
 */
(function (global) {
  "use strict";

  function app() {
    try {
      return global.go && global.go.main && global.go.main.App;
    } catch (_) {
      return null;
    }
  }

  var Bridge = {
    ready: function () {
      return !!app();
    },
    greet: function (name) {
      var a = app();
      if (!a || !a.Greet) return Promise.resolve("");
      return a.Greet(name || "");
    },
    appInfo: function () {
      var a = app();
      if (!a || !a.AppInfo) return Promise.resolve(null);
      return a.AppInfo();
    },
    /** Step 3: OpenURL / Navigate / OpenFile will land here. */
    openURL: function (url) {
      var a = app();
      if (a && a.OpenURL) return a.OpenURL(url);
      // Interim: system browser via anchor
      if (url) {
        var el = document.createElement("a");
        el.href = url;
        el.target = "_blank";
        el.rel = "noopener";
        el.click();
      }
      return Promise.resolve();
    },
    openFile: function () {
      var a = app();
      if (a && a.OpenFile) return a.OpenFile();
      return Promise.resolve("");
    },
    extractDocument: function (path) {
      var a = app();
      if (a && a.ExtractDocument) return a.ExtractDocument(path);
      return Promise.resolve("");
    },
  };

  global.ConductinoBridge = Bridge;
})(window);
