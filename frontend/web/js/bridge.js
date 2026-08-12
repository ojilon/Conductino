/**
 * Conductino JS ↔ Go bridge helpers
 *
 * Go bindings (webview.Bind) appear as global functions:
 *   hostNavigate, hostGoBack, hostGoForward, hostReload,
 *   hostTabNew, hostTabClose, hostTabActivate, hostTabList,
 *   hostShowChrome, hostMinimize, hostMaximize, hostClose, hostPing
 *
 * This module normalizes calls and tab-snapshot application.
 */
(function (global) {
  "use strict";

  function call(name, arg) {
    var fn = global[name];
    if (typeof fn !== "function") {
      console.info("[bridge] missing binding:", name);
      return undefined;
    }
    try {
      if (arguments.length > 1) return fn(arg);
      return fn();
    } catch (e) {
      console.warn("[bridge]", name, e);
      return undefined;
    }
  }

  var Bridge = {
    ping: function () {
      return call("hostPing");
    },
    navigate: function (url) {
      return call("hostNavigate", url);
    },
    back: function () {
      return call("hostGoBack");
    },
    forward: function () {
      return call("hostGoForward");
    },
    reload: function () {
      return call("hostReload");
    },
    showChrome: function () {
      return call("hostShowChrome");
    },
    tabNew: function () {
      return call("hostTabNew");
    },
    tabClose: function (id) {
      return call("hostTabClose", id);
    },
    tabActivate: function (id) {
      return call("hostTabActivate", id);
    },
    tabList: function () {
      var raw = call("hostTabList");
      if (typeof raw !== "string") return [];
      try {
        return JSON.parse(raw);
      } catch (e) {
        return [];
      }
    },
    minimize: function () {
      return call("hostMinimize");
    },
    maximize: function () {
      return call("hostMaximize");
    },
    close: function () {
      return call("hostClose");
    },
  };

  global.ConductinoBridge = Bridge;
})(window);
