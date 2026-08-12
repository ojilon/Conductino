/**
 * Conductino chrome bootstrap
 * Tabs · navigation · sidebar · settings
 * Window min/max/close come from the OS title bar.
 */
(function () {
  "use strict";

  var ENGINES = {
    duckduckgo: { name: "DuckDuckGo", url: "https://duckduckgo.com/?q=%s" },
    google: { name: "Google", url: "https://www.google.com/search?q=%s" },
    bing: { name: "Bing", url: "https://www.bing.com/search?q=%s" },
    startpage: { name: "Startpage", url: "https://www.startpage.com/sp/search?query=%s" },
  };

  var tabSnap = [];
  var settings = loadSettings();

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
    stubTitle: document.getElementById("stub-title"),
    stubBody: document.getElementById("stub-body"),
    settingTheme: document.getElementById("setting-theme"),
    settingEngine: document.getElementById("setting-engine"),
    btnSettingsDone: document.getElementById("btn-settings-done"),
    btnStubBack: document.getElementById("btn-stub-back"),
  };

  function B() {
    return window.ConductinoBridge || null;
  }

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

  function activeFromSnap() {
    for (var i = 0; i < tabSnap.length; i++) {
      if (tabSnap[i].active) return tabSnap[i];
    }
    return tabSnap[0] || null;
  }

  function renderTabs() {
    el.tabs.innerHTML = "";
    tabSnap.forEach(function (tab) {
      var btn = document.createElement("button");
      btn.type = "button";
      btn.className = "tab" + (tab.active ? " active" : "");
      btn.setAttribute("role", "tab");
      btn.setAttribute("aria-selected", tab.active ? "true" : "false");
      btn.dataset.tabId = String(tab.id);

      var title = document.createElement("span");
      title.className = "tab-title";
      title.textContent = tab.title || "New Tab";

      var close = document.createElement("span");
      close.className = "tab-close";
      close.title = "Close tab";
      close.textContent = "×";
      close.addEventListener("click", function (e) {
        e.stopPropagation();
        var b = B();
        if (b) b.tabClose(tab.id);
        else closeTabLocal(tab.id);
      });

      btn.appendChild(title);
      btn.appendChild(close);
      btn.addEventListener("click", function () {
        var b = B();
        if (b) b.tabActivate(tab.id);
        else activateLocal(tab.id);
      });
      el.tabs.appendChild(btn);
    });

    var active = activeFromSnap();
    if (active) {
      el.omnibox.value = active.url || "";
      el.back.disabled = !active.canBack;
      el.fwd.disabled = !active.canFwd;
      updateOmniboxIcon(active.url);
      if (!active.url) showPanel("welcome");
    }
  }

  function applyTabSnapshot(list) {
    if (!Array.isArray(list)) return;
    tabSnap = list;
    renderTabs();
  }

  var localTabs = [];
  var localActive = null;
  var localNext = 1;

  function ensureLocalSeed() {
    if (localTabs.length) return;
    localTabs = [{ id: 1, title: "New Tab", url: "", canBack: false, canFwd: false, active: true }];
    localActive = 1;
    localNext = 2;
    tabSnap = localTabs.slice();
  }

  function activateLocal(id) {
    localTabs.forEach(function (t) {
      t.active = t.id === id;
    });
    localActive = id;
    tabSnap = localTabs.map(function (t) {
      return Object.assign({}, t);
    });
    renderTabs();
  }

  function closeTabLocal(id) {
    localTabs = localTabs.filter(function (t) {
      return t.id !== id;
    });
    if (!localTabs.length) {
      localTabs = [{ id: localNext++, title: "New Tab", url: "", canBack: false, canFwd: false, active: true }];
      localActive = localTabs[0].id;
    } else if (localActive === id) {
      localTabs[0].active = true;
      localActive = localTabs[0].id;
    }
    tabSnap = localTabs.map(function (t) {
      return Object.assign({}, t);
    });
    renderTabs();
  }

  function updateOmniboxIcon(url) {
    if (!url) {
      el.omniboxIcon.textContent = "🔍";
      return;
    }
    if (/^https:\/\//i.test(url)) el.omniboxIcon.textContent = "🔒";
    else if (/^http:\/\//i.test(url)) el.omniboxIcon.textContent = "⚠";
    else el.omniboxIcon.textContent = "📄";
  }

  function showPanel(name) {
    ["welcome", "settings-panel", "stub-panel"].forEach(function (id) {
      var node = document.getElementById(id);
      if (!node) return;
      var on =
        id === name ||
        (name === "settings" && id === "settings-panel") ||
        (name === "stub" && id === "stub-panel");
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

  function looksLikeUrl(input) {
    var s = input.trim();
    if (!s || /\s/.test(s)) return false;
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
    var b = B();
    if (b) {
      b.navigate(url);
      return;
    }
    var t = activeFromSnap();
    if (t) {
      t.url = url;
      t.title = url.replace(/^https?:\/\//, "").split("/")[0] || "Tab";
      renderTabs();
    }
    showStub("Native navigation", "Would call hostNavigate:\n" + url);
  }

  function setSidebarOpen(open) {
    if (open) {
      el.sidebar.removeAttribute("hidden");
      el.btnSidebar.setAttribute("aria-pressed", "true");
    } else {
      el.sidebar.setAttribute("hidden", "");
      el.btnSidebar.setAttribute("aria-pressed", "false");
    }
  }

  el.newTab.addEventListener("click", function () {
    var b = B();
    if (b) b.tabNew();
    else {
      ensureLocalSeed();
      localTabs.forEach(function (t) {
        t.active = false;
      });
      var nt = { id: localNext++, title: "New Tab", url: "", canBack: false, canFwd: false, active: true };
      localTabs.push(nt);
      localActive = nt.id;
      tabSnap = localTabs.map(function (t) {
        return Object.assign({}, t);
      });
      renderTabs();
      showPanel("welcome");
    }
  });

  el.back.addEventListener("click", function () {
    var b = B();
    if (b) b.back();
  });
  el.fwd.addEventListener("click", function () {
    var b = B();
    if (b) b.forward();
  });
  el.reload.addEventListener("click", function () {
    var b = B();
    if (b) b.reload();
  });

  el.omniboxForm.addEventListener("submit", function (e) {
    e.preventDefault();
    navigateTo(el.omnibox.value);
  });

  el.btnSidebar.addEventListener("click", function () {
    setSidebarOpen(el.sidebar.hasAttribute("hidden"));
  });
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
        showStub("Downloads", "Download manager will live here. See docs/GUI.md.");
        setSidebarOpen(false);
      } else if (action === "bookmarks") {
        showStub("Bookmarks", "Bookmarks UI will live here. See docs/GUI.md.");
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

  applyTheme(settings.theme || "aurora-dark");
  el.settingEngine.value = settings.engine || "duckduckgo";

  window.ConductinoChrome = {
    applyTabSnapshot: applyTabSnapshot,
    setTabMeta: function (meta) {
      var active = activeFromSnap();
      if (!active || !meta) return;
      if (meta.title) active.title = meta.title;
      if (typeof meta.url === "string") active.url = meta.url;
      if (typeof meta.canBack === "boolean") active.canBack = meta.canBack;
      if (typeof meta.canFwd === "boolean") active.canFwd = meta.canFwd;
      renderTabs();
    },
    showWelcome: function () {
      showPanel("welcome");
    },
  };

  function syncFromHost() {
    var b = B();
    if (!b) {
      ensureLocalSeed();
      renderTabs();
      return;
    }
    var list = b.tabList();
    if (list && list.length) applyTabSnapshot(list);
    else {
      ensureLocalSeed();
      renderTabs();
    }
  }

  syncFromHost();
  setTimeout(syncFromHost, 50);
  setTimeout(syncFromHost, 200);
})();
