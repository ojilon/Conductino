/**
 * Chrome bootstrap — tabs, omnibox, right sidebar, settings, library.
 */
(function () {
  "use strict";

  var ENGINES = {
    duckduckgo: { name: "DuckDuckGo", url: "https://duckduckgo.com/?q=%s" },
    google: { name: "Google", url: "https://www.google.com/search?q=%s" },
    bing: { name: "Bing", url: "https://www.bing.com/search?q=%s" },
    startpage: { name: "Startpage", url: "https://www.startpage.com/sp/search?query=%s" },
  };

  var AI_STORAGE_KEY = "conductino.ai";

  var tabs = [];
  var nextId = 1;
  var activeId = null;
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

  function activeTab() {
    for (var i = 0; i < tabs.length; i++) if (tabs[i].id === activeId) return tabs[i];
    return tabs[0] || null;
  }

  function renderTabs() {
    if (!el.tabs) return;
    el.tabs.innerHTML = "";
    tabs.forEach(function (tab) {
      var btn = document.createElement("button");
      btn.type = "button";
      btn.className = "tab" + (tab.id === activeId ? " active" : "");
      btn.setAttribute("role", "tab");
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
        closeTab(tab.id);
      });

      btn.appendChild(title);
      btn.appendChild(close);
      btn.addEventListener("click", function () {
        activateTab(tab.id);
      });
      el.tabs.appendChild(btn);
    });

    var t = activeTab();
    if (t && el.omnibox) {
      el.omnibox.value = t.url || "";
      updateOmniboxIcon(t.url);
    }
    if (el.back) el.back.disabled = true;
    if (el.fwd) el.fwd.disabled = true;
  }

  function ensureSeed() {
    if (tabs.length) return;
    tabs.push({ id: nextId++, title: "New Tab", url: "", panel: "welcome" });
    activeId = tabs[0].id;
  }

  function activateTab(id) {
    activeId = id;
    var t = activeTab();
    renderTabs();
    if (t && t.panel) showPanel(t.panel);
    else showPanel("welcome");
  }

  function closeTab(id) {
    tabs = tabs.filter(function (t) {
      return t.id !== id;
    });
    if (!tabs.length) {
      ensureSeed();
    } else if (activeId === id) {
      activeId = tabs[0].id;
    }
    renderTabs();
    var t = activeTab();
    showPanel(t && t.panel ? t.panel : "welcome");
  }

  function newTab() {
    var t = { id: nextId++, title: "New Tab", url: "", panel: "welcome" };
    tabs.push(t);
    activeId = t.id;
    renderTabs();
    showPanel("welcome");
  }

  function updateOmniboxIcon(url) {
    if (!el.omniboxIcon) return;
    if (!url) el.omniboxIcon.textContent = "🔍";
    else if (/^https:\/\//i.test(url)) el.omniboxIcon.textContent = "🔒";
    else if (/^http:\/\//i.test(url)) el.omniboxIcon.textContent = "⚠";
    else el.omniboxIcon.textContent = "📄";
  }

  var PANEL_IDS = ["welcome", "settings-panel", "stub-panel", "study-panel", "library-panel"];

  function showPanel(name) {
    PANEL_IDS.forEach(function (id) {
      var node = document.getElementById(id);
      if (!node) return;
      var on =
        id === name ||
        (name === "settings" && id === "settings-panel") ||
        (name === "stub" && id === "stub-panel") ||
        (name === "study" && id === "study-panel") ||
        (name === "library" && id === "library-panel");
      node.classList.toggle("active", on);
      if (on) node.removeAttribute("hidden");
      else node.setAttribute("hidden", "");
    });
    var t = activeTab();
    if (t) {
      if (name === "settings") t.panel = "settings-panel";
      else if (name === "stub") t.panel = "stub-panel";
      else if (name === "study") t.panel = "study-panel";
      else if (name === "library") t.panel = "library-panel";
      else t.panel = name;
    }
    if (name === "settings" || name === "settings-panel") fillAiFormFromStorage();
    if ((name === "library" || name === "library-panel") && window.ConductinoLibrary) {
      window.ConductinoLibrary.refresh();
    }
  }

  function showStub(title, body) {
    if (el.stubTitle) el.stubTitle.textContent = title;
    if (el.stubBody) el.stubBody.textContent = body;
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
    var t = activeTab();
    if (t) {
      t.url = url;
      t.title = url.replace(/^https?:\/\//, "").split("/")[0] || "Tab";
      renderTabs();
    }
    var b = window.ConductinoBridge;
    if (b && b.openURL) b.openURL(url);
    else showStub("Open URL", url);
  }

  function setSidebarOpen(open) {
    if (!el.sidebar) return;
    if (open) {
      el.sidebar.removeAttribute("hidden");
      el.sidebar.classList.add("open");
      if (el.btnSidebar) el.btnSidebar.setAttribute("aria-pressed", "true");
    } else {
      el.sidebar.setAttribute("hidden", "");
      el.sidebar.classList.remove("open");
      if (el.btnSidebar) el.btnSidebar.setAttribute("aria-pressed", "false");
    }
  }

  function loadAiConfigs() {
    try {
      var raw = localStorage.getItem(AI_STORAGE_KEY);
      if (!raw) return [];
      var parsed = JSON.parse(raw);
      return Array.isArray(parsed) ? parsed : parsed ? [parsed] : [];
    } catch (e) {
      return [];
    }
  }

  function saveAiConfigs(list) {
    localStorage.removeItem(AI_STORAGE_KEY);
    localStorage.setItem(AI_STORAGE_KEY, JSON.stringify(list || []));
    if (window.ConductinoAI && window.ConductinoAI.setProviders) {
      window.ConductinoAI.setProviders(list || []);
    }
  }

  function maskKey(key) {
    if (!key) return "(none)";
    if (key.length <= 8) return "****";
    return key.slice(0, 4) + "…" + key.slice(-4);
  }

  function updateAiStatusLine(cfg) {
    var line = document.getElementById("ai-config-status");
    if (!line) return;
    if (!cfg || !cfg.apiKey) {
      line.textContent = "No API key saved.";
      return;
    }
    line.textContent =
      "Saved: " + (cfg.name || cfg.id) + " · model " + (cfg.model || "?") + " · key " + maskKey(cfg.apiKey);
  }

  function googleEndpointForModel(model) {
    var m = model || "gemini-2.5-flash";
    return (
      "https://generativelanguage.googleapis.com/v1beta/models/" +
      encodeURIComponent(m) +
      ":generateContent"
    );
  }

  function buildProviderFromForm() {
    var preset = (document.getElementById("setting-ai-preset") || {}).value || "google";
    var key = ((document.getElementById("setting-ai-key") || {}).value || "").trim();
    var endpoint = ((document.getElementById("setting-ai-endpoint") || {}).value || "").trim();
    var model = ((document.getElementById("setting-ai-model") || {}).value || "").trim();
    var base = {
      id: preset,
      name: preset,
      apiKey: key,
      endpoint: endpoint,
      model: model,
      style: preset === "google" ? "google" : "openai",
    };
    if (preset === "google") {
      base.name = "Google AI Studio";
      base.model = model || "gemini-2.5-flash";
      base.endpoint = endpoint || googleEndpointForModel(base.model);
      base.style = "google";
    } else if (preset === "openrouter") {
      base.name = "OpenRouter";
      base.endpoint = endpoint || "https://openrouter.ai/api/v1/chat/completions";
      base.model = model || "google/gemini-2.5-flash";
      base.style = "openai";
    } else if (preset === "groq") {
      base.name = "Groq";
      base.endpoint = endpoint || "https://api.groq.com/openai/v1/chat/completions";
      base.model = model || "llama-3.3-70b-versatile";
      base.style = "openai";
    } else {
      base.name = "Custom";
      base.style = "openai";
    }
    return base;
  }

  function fillAiFormFromStorage() {
    var list = loadAiConfigs();
    var cfg = list[0] || null;
    var presetEl = document.getElementById("setting-ai-preset");
    var keyEl = document.getElementById("setting-ai-key");
    var endpointEl = document.getElementById("setting-ai-endpoint");
    var modelEl = document.getElementById("setting-ai-model");
    if (cfg) {
      if (presetEl && cfg.id) presetEl.value = cfg.id;
      if (keyEl) keyEl.value = cfg.apiKey || "";
      if (endpointEl) endpointEl.value = cfg.endpoint || "";
      if (modelEl) modelEl.value = cfg.model || "";
    }
    updateAiStatusLine(cfg);
  }

  function bindAiSettings() {
    var btn = document.getElementById("btn-ai-save");
    var btnClear = document.getElementById("btn-ai-clear");
    if (btn) {
      btn.addEventListener("click", function () {
        var base = buildProviderFromForm();
        if (!base.apiKey) {
          alert("API key required");
          return;
        }
        if (!base.endpoint) {
          alert("Endpoint required (or pick a preset)");
          return;
        }
        saveAiConfigs([base]);
        updateAiStatusLine(base);
        alert("Saved " + base.name + "\nModel: " + base.model + "\nKey: " + maskKey(base.apiKey));
      });
    }
    if (btnClear) {
      btnClear.addEventListener("click", function () {
        if (!confirm("Clear all saved AI keys?")) return;
        saveAiConfigs([]);
        ["setting-ai-key", "setting-ai-endpoint", "setting-ai-model"].forEach(function (id) {
          var n = document.getElementById(id);
          if (n) n.value = "";
        });
        updateAiStatusLine(null);
      });
    }
    fillAiFormFromStorage();
  }

  async function quickBookmark() {
    var b = window.ConductinoBridge;
    if (!b) return;
    var url = (el.omnibox && el.omnibox.value) || "";
    url = url.trim();
    if (!url) {
      url = window.prompt("URL to bookmark") || "";
    }
    if (!url) return;
    var folder = window.prompt("Library folder (e.g. plantphysiology/growth)");
    if (!folder) return;
    try {
      await b.createLibraryFolder(folder);
      await b.addBookmark(folder, url, url);
      alert("Bookmarked into " + folder);
      if (window.ConductinoLibrary) window.ConductinoLibrary.refresh();
    } catch (e) {
      alert("Bookmark failed: " + e);
    }
  }

  if (el.newTab) el.newTab.addEventListener("click", newTab);
  if (el.omniboxForm) {
    el.omniboxForm.addEventListener("submit", function (e) {
      e.preventDefault();
      navigateTo(el.omnibox.value);
    });
  }
  if (el.btnSidebar) {
    el.btnSidebar.addEventListener("click", function () {
      setSidebarOpen(el.sidebar.hasAttribute("hidden"));
    });
  }
  if (el.btnSidebarClose) {
    el.btnSidebarClose.addEventListener("click", function () {
      setSidebarOpen(false);
    });
  }

  document.querySelectorAll(".sidebar-item").forEach(function (btn) {
    btn.addEventListener("click", function () {
      var action = btn.getAttribute("data-action");
      if (action === "settings") showPanel("settings");
      else if (action === "study") showPanel("study");
      else if (action === "library") showPanel("library");
      else if (action === "downloads") showStub("Downloads", "Use Library folders → downloads/ (coming next).");
      else if (action === "bookmarks") showPanel("library");
      setSidebarOpen(false);
    });
  });

  var btnStudy = document.getElementById("btn-study");
  if (btnStudy) btnStudy.addEventListener("click", function () { showPanel("study"); });
  var btnOpenStudy = document.getElementById("btn-open-study");
  if (btnOpenStudy) btnOpenStudy.addEventListener("click", function () { showPanel("study"); });
  var btnOpenLib = document.getElementById("btn-open-library");
  if (btnOpenLib) btnOpenLib.addEventListener("click", function () { showPanel("library"); });
  var btnBm = document.getElementById("btn-bookmark");
  if (btnBm) btnBm.addEventListener("click", quickBookmark);

  if (el.settingTheme) {
    el.settingTheme.addEventListener("change", function () {
      applyTheme(el.settingTheme.value);
    });
  }
  if (el.settingEngine) {
    el.settingEngine.addEventListener("change", function () {
      settings.engine = el.settingEngine.value;
      saveSettings();
    });
  }
  if (el.btnSettingsDone) {
    el.btnSettingsDone.addEventListener("click", function () {
      showPanel("welcome");
    });
  }
  if (el.btnStubBack) {
    el.btnStubBack.addEventListener("click", function () {
      showPanel("welcome");
    });
  }

  bindAiSettings();
  applyTheme(settings.theme || "aurora-dark");
  if (el.settingEngine) el.settingEngine.value = settings.engine || "duckduckgo";

  window.ConductinoChrome = {
    showPanel: showPanel,
    showStub: showStub,
  };

  ensureSeed();
  renderTabs();
  showPanel("welcome");
})();
