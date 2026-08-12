/**
 * Conductino chrome bootstrap
 * Tabs · navigation · sidebar · settings · window controls
 *
 * Remote navigation is requested via Go bindings (native webview).
 * Local state panels (welcome, settings, stubs) live in #content-host.
 * See docs/GUI.md.
 */
(function () {
  "use strict";

  // —— Search engines (mirror Android search_engines idea) ——
  var ENGINES = {
    duckduckgo: { name: "DuckDuckGo", url: "https://duckduckgo.com/?q=%s" },
    google: { name: "Google", url: "https://www.google.com/search?q=%s" },
    bing: { name: "Bing", url: "https://www.bing.com/search?q=%s" },
    startpage: { name: "Startpage", url: "https://www.startpage.com/sp/search?query=%s" },
  };

  // —— State ——
  var tabs = [];
  var activeTabId = null;
  var nextTabId = 1;
  var settings = loadSettings();

  // —— DOM ——
  var el = {
    tabs: document.getElementById("tabs"),
    newTab: document.getElementById("btn-new-tab"),
    back: document.getElementById("btn-back"),
    fwd: document.getElementById("btn-fwd"),
    reload: document.getElementById("btn-reload"),
    omnibox: document.getElementById("omnibox"),
    omniboxForm: document.getElementById("omnibox-form"),
    omniboxIcon: document.getElementById("omnibox-icon"),
    sidebar: document.getElementById("sidebar"),
    btnSidebar: document.getElementById("btn-sidebar"),
    btnSidebarClose: document.getElementById("btn-sidebar-close"),
    welcome: document.getElementById("welcome"),
    settingsPanel: document.getElementById("settings-panel"),
    stubPanel: document.getElementById("stub-panel"),
    stubTitle: document.getElementById("stub-title"),
    stubBody: document.getElementById("stub-body"),
    settingTheme: document.getElementById("setting-theme"),
    settingEngine: document.getElementById("setting-engine"),
    btnSettingsDone: document.getElementById("btn-settings-done"),
    btnStubBack: document.getElementById("btn-stub-back"),
    btnMin: document.getElementById("btn-min"),
    btnMax: document.getElementById("btn-max"),
    btnClose: document.getElementById("btn-close"),
  };

  // —— Settings persistence (local for now) ——
  function loadSettings() {
    try {
      var raw = localStorage.getItem("conductino.settings");
      if (raw) return JSON.parse(raw);
    } catch (e) {}
    return { theme: "aurora-dark", engine: "duckduckgo" };
  }

  function saveSettings() {
    try {
      localStorage.setItem("conductino.settings", JSON.stringify(settings));
    } catch (e) {}
  }

  function applyTheme(name) {
    document.documentElement.setAttribute("data-theme", name);
    settings.theme = name;
    if (el.settingTheme) el.settingTheme.value = name;
    saveSettings();
  }

  // —— Tabs ——
  function createTab(opts) {
    opts = opts || {};
    var tab = {
      id: nextTabId++,
      title: opts.title || "New Tab",
      url: opts.url || "",
      canBack: false,
      canFwd: false,
    };
    tabs.push(tab);
    renderTabs();
    activateTab(tab.id);
    return tab;
  }

  function activateTab(id) {
    activeTabId = id;
    var tab = tabs.find(function (t) {
      return t.id === id;
    });
    if (!tab) return;
    renderTabs();
    el.omnibox.value = tab.url || "";
    updateNavButtons(tab);
    updateOmniboxIcon(tab.url);
    if (!tab.url) {
      showPanel("welcome");
    }
  }

  function closeTab(id) {
    var idx = tabs.findIndex(function (t) {
      return t.id === id;
    });
    if (idx < 0) return;
    tabs.splice(idx, 1);
    if (tabs.length === 0) {
      createTab();
      return;
    }
    if (activeTabId === id) {
      var next = tabs[Math.max(0, idx - 1)];
      activateTab(next.id);
    } else {
      renderTabs();
    }
  }

  function activeTab() {
    return tabs.find(function (t) {
      return t.id === activeTabId;
    });
  }

  function renderTabs() {
    el.tabs.innerHTML = "";
    tabs.forEach(function (tab) {
      var btn = document.createElement("button");
      btn.type = "button";
      btn.className = "tab" + (tab.id === activeTabId ? " active" : "");
      btn.setAttribute("role", "tab");
      btn.setAttribute("aria-selected", tab.id === activeTabId ? "true" : "false");
      btn.dataset.tabId = String(tab.id);

      var title = document.createElement("span");
      title.className = "tab-title";
      title.textContent = tab.title;

      var close = document.createElement("span");
      close.className = "tab-close";
      close.title = "Close tab";
      close.textContent = "×";
      close.addEventListener("click", function (e) {
        e.stopPropagation();
        closeTab(tab.id);
      });

      btn.appendChild(title);
      btn.appendChild(close);
      btn.addEventListener("click", function () {
        activateTab(tab.id);
      });
      el.tabs.appendChild(btn);
    });
  }

  function updateNavButtons(tab) {
    el.back.disabled = !tab || !tab.canBack;
    el.fwd.disabled = !tab || !tab.canFwd;
  }

  function updateOmniboxIcon(url) {
    if (!url) {
      el.omniboxIcon.textContent = "🔍";
      return;
    }
    if (/^https:\/\//i.test(url)) {
      el.omniboxIcon.textContent = "🔒";
    } else if (/^http:\/\//i.test(url)) {
      el.omniboxIcon.textContent = "⚠";
    } else {
      el.omniboxIcon.textContent = "📄";
    }
  }

  // —— Panels (local states) ——
  function showPanel(name) {
    ["welcome", "settings-panel", "stub-panel"].forEach(function (id) {
      var node = document.getElementById(id);
      if (!node) return;
      var on = id === name || (name === "settings" && id === "settings-panel") || (name === "stub" && id === "stub-panel");
      node.classList.toggle("active", on);
      if (on) node.removeAttribute("hidden");
      else node.setAttribute("hidden", "");
    });
  }

  function showStub(title, body) {
    el.stubTitle.textContent = title;
    el.stubBody.textContent = body;
    showPanel("stub");
  }

  // —— Navigation ——
  function looksLikeUrl(input) {
    var s = input.trim();
    if (!s) return false;
    if (/\s/.test(s)) return false;
    if (/^(https?:\/\/|about:|file:)/i.test(s)) return true;
    if (/^[a-z0-9.-]+\.[a-z]{2,}([\/:].*)?$/i.test(s)) return true;
    if (/^localhost(:\d+)?([\/:].*)?$/i.test(s)) return true;
    return false;
  }

  function normalizeUrl(input) {
    var s = input.trim();
    if (/^(https?:\/\/|about:|file:)/i.test(s)) return s;
    return "https://" + s;
  }

  function buildSearchUrl(query) {
    var eng = ENGINES[settings.engine] || ENGINES.duckduckgo;
    return eng.url.replace("%s", encodeURIComponent(query));
  }

  function navigateTo(input) {
    var raw = (input || "").trim();
    if (!raw) return;

    var url = looksLikeUrl(raw) ? normalizeUrl(raw) : buildSearchUrl(raw);
    var tab = activeTab();
    if (tab) {
      tab.url = url;
      tab.title = url.replace(/^https?:\/\//, "").split("/")[0] || "Tab";
      renderTabs();
      el.omnibox.value = url;
      updateOmniboxIcon(url);
    }

    // Native navigation via Go binding when available.
    if (typeof window.hostNavigate === "function") {
      window.hostNavigate(url);
    } else {
      console.info("[conductino] hostNavigate not bound yet — would navigate to:", url);
      showStub("Native navigation", "Go binding hostNavigate will load:\n" + url);
    }
  }

  // —— Sidebar ——
  function setSidebarOpen(open) {
    if (open) {
      el.sidebar.removeAttribute("hidden");
      el.btnSidebar.setAttribute("aria-pressed", "true");
    } else {
      el.sidebar.setAttribute("hidden", "");
      el.btnSidebar.setAttribute("aria-pressed", "false");
    }
  }

  function toggleSidebar() {
    setSidebarOpen(el.sidebar.hasAttribute("hidden"));
  }

  // —— Host bridge helpers (safe if unbound) ——
  function hostCall(name) {
    var fn = window[name];
    if (typeof fn === "function") {
      try {
        return fn();
      } catch (e) {
        console.warn(name, e);
      }
    } else {
      console.info("[conductino] binding missing:", name);
    }
  }

  // —— Wire events ——
  el.newTab.addEventListener("click", function () {
    createTab();
  });

  el.back.addEventListener("click", function () {
    hostCall("hostGoBack");
  });
  el.fwd.addEventListener("click", function () {
    hostCall("hostGoForward");
  });
  el.reload.addEventListener("click", function () {
    hostCall("hostReload");
  });

  el.omniboxForm.addEventListener("submit", function (e) {
    e.preventDefault();
    navigateTo(el.omnibox.value);
  });

  el.btnSidebar.addEventListener("click", toggleSidebar);
  el.btnSidebarClose.addEventListener("click", function () {
    setSidebarOpen(false);
  });

  document.querySelectorAll(".sidebar-item").forEach(function (btn) {
    btn.addEventListener("click", function () {
      var action = btn.getAttribute("data-action");
      if (action === "settings") {
        showPanel("settings");
        setSidebarOpen(false);
      } else if (action === "downloads") {
        showStub("Downloads", "Download manager will live here. See docs/GUI.md for how to implement this sidebar item.");
        setSidebarOpen(false);
      } else if (action === "bookmarks") {
        showStub("Bookmarks", "Bookmarks UI will live here. See docs/GUI.md for how to implement this sidebar item.");
        setSidebarOpen(false);
      }
    });
  });

  el.settingTheme.addEventListener("change", function () {
    applyTheme(el.settingTheme.value);
  });
  el.settingEngine.addEventListener("change", function () {
    settings.engine = el.settingEngine.value;
    saveSettings();
  });
  el.btnSettingsDone.addEventListener("click", function () {
    showPanel("welcome");
  });
  el.btnStubBack.addEventListener("click", function () {
    showPanel("welcome");
  });

  el.btnMin.addEventListener("click", function () {
    hostCall("hostMinimize");
  });
  el.btnMax.addEventListener("click", function () {
    hostCall("hostMaximize");
  });
  el.btnClose.addEventListener("click", function () {
    hostCall("hostClose");
  });

  // —— Boot ——
  applyTheme(settings.theme || "aurora-dark");
  el.settingEngine.value = settings.engine || "duckduckgo";
  createTab({ title: "New Tab" });

  // Expose a tiny API for Go → JS updates later (title, canBack, etc.)
  window.ConductinoChrome = {
    setTabMeta: function (meta) {
      var tab = activeTab();
      if (!tab || !meta) return;
      if (meta.title) tab.title = meta.title;
      if (typeof meta.url === "string") tab.url = meta.url;
      if (typeof meta.canBack === "boolean") tab.canBack = meta.canBack;
      if (typeof meta.canFwd === "boolean") tab.canFwd = meta.canFwd;
      renderTabs();
      updateNavButtons(tab);
      if (typeof meta.url === "string") {
        el.omnibox.value = meta.url;
        updateOmniboxIcon(meta.url);
      }
    },
    showWelcome: function () {
      showPanel("welcome");
    },
  };
})();
