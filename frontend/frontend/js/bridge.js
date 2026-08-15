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

  function call(method) {
    var a = app();
    if (!a || typeof a[method] !== "function") {
      return Promise.reject(new Error("binding missing: " + method));
    }
    var args = Array.prototype.slice.call(arguments, 1);
    return a[method].apply(a, args);
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
        var el = document.createElement("a");
        el.href = url;
        el.target = "_blank";
        el.rel = "noopener";
        el.click();
      });
    },
    fetchPageText: function (url) {
      return call("FetchPageText", url).then(function (t) {
        return typeof t === "string" ? t : "";
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

    libraryRoot: function () {
      return call("LibraryRoot");
    },
    listLibraryFolders: function () {
      return call("ListLibraryFolders").then(function (x) {
        return x || [];
      });
    },
    createLibraryFolder: function (rel) {
      return call("CreateLibraryFolder", rel);
    },
    listBookmarks: function (folder) {
      return call("ListBookmarks", folder || "").then(function (x) {
        return x || [];
      });
    },
    addBookmark: function (folder, url, title) {
      return call("AddBookmark", folder, url, title || "");
    },
    removeBookmark: function (id) {
      return call("RemoveBookmark", id);
    },
    ensureSummary: function (folder) {
      return call("EnsureSummary", folder);
    },
    appendSummary: function (folder, sectionTitle, body) {
      return call("AppendSummary", folder, sectionTitle || "", body || "");
    },
    readSummaryFile: function (relPath) {
      return call("ReadSummaryFile", relPath);
    },
    writeSummaryFile: function (relPath, content) {
      return call("WriteSummaryFile", relPath, content);
    },
    listSummaries: function (folder) {
      return call("ListSummaries", folder || "").then(function (x) {
        return x || [];
      });
    },
    mergeSummaryFiles: function (a, b, destFolder, title) {
      return call("MergeSummaryFiles", a, b, destFolder, title || "");
    },
    pickLibraryFolder: function () {
      return call("PickLibraryFolder").then(function (p) {
        return typeof p === "string" ? p : "";
      });
    },
  };

  global.ConductinoBridge = Bridge;
})(window);
