/**
 * Conductino JS ↔ Go bridge helpers
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
    sidebarOpen: function (open) {
      return call("hostSidebarOpen", !!open);
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
