/** Library panel — folders, bookmarks, summary files. */
(function () {
  "use strict";

  var selectedFolder = "";

  function $(id) {
    return document.getElementById(id);
  }

  function status(msg) {
    var s = $("library-status");
    if (s) s.textContent = msg || "";
  }

  function bridge() {
    return window.ConductinoBridge;
  }

  async function refreshFolders() {
    var b = bridge();
    var list = $("library-folder-list");
    if (!b || !list) return;
    var folders = [];
    try {
      folders = await b.listLibraryFolders();
    } catch (e) {
      status("Folders: " + e);
      return;
    }
    list.innerHTML = "";
    if (!folders.length) {
      list.innerHTML = "<p class=\"muted small\" style=\"padding:8px\">No folders yet. Create one below.</p>";
      return;
    }
    folders.forEach(function (f) {
      var btn = document.createElement("button");
      btn.type = "button";
      btn.textContent = f;
      if (f === selectedFolder) btn.className = "active";
      btn.addEventListener("click", function () {
        selectedFolder = f;
        refreshFolders();
        refreshBookmarks();
        refreshSummaries();
      });
      list.appendChild(btn);
    });
  }

  async function refreshBookmarks() {
    var b = bridge();
    var list = $("library-bookmark-list");
    if (!b || !list) return;
    list.innerHTML = "";
    if (!selectedFolder) {
      list.innerHTML = "<p class=\"muted small\" style=\"padding:8px\">Select a folder.</p>";
      return;
    }
    var items = [];
    try {
      items = await b.listBookmarks(selectedFolder);
    } catch (e) {
      status("Bookmarks: " + e);
      return;
    }
    if (!items.length) {
      list.innerHTML = "<p class=\"muted small\" style=\"padding:8px\">No bookmarks in this folder.</p>";
      return;
    }
    items.forEach(function (bm) {
      var btn = document.createElement("button");
      btn.type = "button";
      btn.innerHTML =
        "<strong>" +
        escapeHtml(bm.title || bm.url) +
        "</strong><span class=\"meta\">" +
        escapeHtml(bm.url) +
        "</span>";
      btn.addEventListener("click", function () {
        if (b.openURL) b.openURL(bm.url);
      });
      btn.addEventListener("contextmenu", async function (e) {
        e.preventDefault();
        if (confirm("Remove bookmark?\n" + bm.title)) {
          await b.removeBookmark(bm.id);
          refreshBookmarks();
        }
      });
      list.appendChild(btn);
    });
  }

  async function refreshSummaries() {
    var b = bridge();
    var list = $("library-summary-list");
    if (!b || !list) return;
    list.innerHTML = "";
    if (!selectedFolder) {
      list.innerHTML = "<p class=\"muted small\" style=\"padding:8px\">Select a folder.</p>";
      return;
    }
    var items = [];
    try {
      items = await b.listSummaries(selectedFolder);
    } catch (e) {
      status("Summaries: " + e);
      return;
    }
    if (!items.length) {
      list.innerHTML = "<p class=\"muted small\" style=\"padding:8px\">No summary file yet — use Ensure summary.</p>";
      return;
    }
    items.forEach(function (s) {
      var btn = document.createElement("button");
      btn.type = "button";
      btn.innerHTML =
        "<strong>" +
        escapeHtml(s.title || s.relPath) +
        "</strong><span class=\"meta\">" +
        escapeHtml(s.relPath) +
        "</span>";
      btn.addEventListener("click", async function () {
        try {
          var text = await b.readSummaryFile(s.relPath);
          if (window.ConductinoStudy && window.ConductinoStudy.loadKnowledgeText) {
            window.ConductinoStudy.loadKnowledgeText(text, s.relPath);
          }
          if (window.ConductinoChrome && window.ConductinoChrome.showPanel) {
            window.ConductinoChrome.showPanel("study");
          }
          status("Opened summary in Study knowledge pane");
        } catch (err) {
          status("Read failed: " + err);
        }
      });
      list.appendChild(btn);
    });
  }

  function escapeHtml(s) {
    return String(s || "")
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;")
      .replace(/"/g, "&quot;");
  }

  function currentOmniboxUrl() {
    var o = $("omnibox");
    return o ? (o.value || "").trim() : "";
  }

  function init() {
    var create = $("btn-lib-create-folder");
    var addBm = $("btn-lib-add-bookmark");
    var ensure = $("btn-lib-ensure-summary");
    var openRoot = $("btn-lib-open-root");
    var merge = $("btn-lib-merge");
    var refresh = $("btn-lib-refresh");

    if (create) {
      create.addEventListener("click", async function () {
        var rel = window.prompt("Folder path (e.g. plantphysiology/growth)");
        if (!rel) return;
        try {
          selectedFolder = await bridge().createLibraryFolder(rel);
          status("Created " + selectedFolder);
          await refreshFolders();
          await refreshBookmarks();
          await refreshSummaries();
        } catch (e) {
          status("Create failed: " + e);
        }
      });
    }

    if (addBm) {
      addBm.addEventListener("click", async function () {
        if (!selectedFolder) {
          status("Select or create a folder first");
          return;
        }
        var url = currentOmniboxUrl() || window.prompt("Paper URL");
        if (!url) return;
        var title = window.prompt("Title (optional)", url) || url;
        try {
          await bridge().addBookmark(selectedFolder, url, title);
          status("Bookmarked into " + selectedFolder);
          await refreshBookmarks();
        } catch (e) {
          status("Bookmark failed: " + e);
        }
      });
    }

    if (ensure) {
      ensure.addEventListener("click", async function () {
        if (!selectedFolder) {
          status("Select a folder first");
          return;
        }
        try {
          var ref = await bridge().ensureSummary(selectedFolder);
          status("Summary ready: " + ref.relPath);
          await refreshSummaries();
        } catch (e) {
          status("Ensure failed: " + e);
        }
      });
    }

    if (openRoot) {
      openRoot.addEventListener("click", async function () {
        try {
          var root = await bridge().libraryRoot();
          status("Library on disk: " + root);
          if (bridge().openURL) bridge().openURL("file:///" + root.replace(/\\/g, "/"));
        } catch (e) {
          status("" + e);
        }
      });
    }

    if (merge) {
      merge.addEventListener("click", async function () {
        var a = window.prompt("Relative path of summary A (from list meta)");
        var b = window.prompt("Relative path of summary B");
        var dest = window.prompt("Dest folder (e.g. plantphysiology/merged)", selectedFolder || "");
        var title = window.prompt("New file title", "merged-summary");
        if (!a || !b || !dest) return;
        try {
          var ref = await bridge().mergeSummaryFiles(a, b, dest, title);
          status("Merged → " + ref.relPath);
          selectedFolder = dest;
          await refreshFolders();
          await refreshSummaries();
        } catch (e) {
          status("Merge failed: " + e);
        }
      });
    }

    if (refresh) {
      refresh.addEventListener("click", async function () {
        await refreshFolders();
        await refreshBookmarks();
        await refreshSummaries();
        status("Refreshed");
      });
    }

    refreshFolders();
  }

  window.ConductinoLibrary = {
    refresh: function () {
      return refreshFolders().then(refreshBookmarks).then(refreshSummaries);
    },
    selectedFolder: function () {
      return selectedFolder;
    },
    setFolder: function (f) {
      selectedFolder = f || "";
    },
  };

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init);
  } else {
    init();
  }
})();
