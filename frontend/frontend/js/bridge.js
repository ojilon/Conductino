/**
 * Wails JS ↔ Go bridge helpers.
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

  function call(method, arg) {
    var a = app();
    if (!a || typeof a[method] !== "function") {
      return Promise.reject(new Error("binding missing: " + method));
    }
    if (arguments.length > 1) return a[method](arg);
    return a[method]();
  }

  var Bridge = {
    ready: function () {
      return !!app();
    },
    greet: function (name) {
      return call("Greet", name || "");
    },
    appInfo: function () {
      return call("AppInfo");
    },
    openURL: function (url) {
      return call("OpenURL", url).catch(function () {
        // Fallback if binding not ready
        var el = document.createElement("a");
        el.href = url;
        el.target = "_blank";
        el.rel = "noopener";
        el.click();
      });
    },
    openFile: function () {
      return call("OpenFile").then(function (p) {
        return typeof p === "string" ? p : "";
      }).catch(function () {
        return "";
      });
    },
    extractDocument: function (path) {
      return call("ExtractDocument", path).then(function (t) {
        return typeof t === "string" ? t : "";
      }).catch(function (e) {
        console.warn("[bridge] extract", e);
        return "";
      });
    },
    importDocument: function (path) {
      return call("ImportDocument", path).then(function (p) {
        return typeof p === "string" ? p : "";
      }).catch(function () {
        return "";
      });
    },
    windowMinimise: function () {
      return call("WindowMinimise");
    },
    windowToggleMaximise: function () {
      return call("WindowToggleMaximise");
    },
    windowClose: function () {
      return call("WindowClose");
    },
  };

  global.ConductinoBridge = Bridge;
})(window);
