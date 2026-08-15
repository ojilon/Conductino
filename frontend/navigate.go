package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

var (
	homeMu  sync.RWMutex
	homeURL = "http://wails.localhost"
)

// SetHomeURL is called once from the asset UI so we can return after browsing.
func (a *App) SetHomeURL(url string) {
	url = strings.TrimSpace(url)
	if url == "" {
		return
	}
	// origin only
	if i := strings.Index(url, "?"); i >= 0 {
		url = url[:i]
	}
	if i := strings.Index(url, "#"); i >= 0 {
		url = url[:i]
	}
	url = strings.TrimRight(url, "/")
	homeMu.Lock()
	homeURL = url
	homeMu.Unlock()
}

func (a *App) getHomeURL() string {
	homeMu.RLock()
	defer homeMu.RUnlock()
	return homeURL
}

// Navigate loads a URL in the main WebView2 (real browser navigation — not an iframe).
func (a *App) Navigate(url string) error {
	if a.ctx == nil {
		return fmt.Errorf("app not started")
	}
	url = strings.TrimSpace(url)
	if url == "" {
		return fmt.Errorf("empty url")
	}
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") &&
		!strings.HasPrefix(url, "about:") {
		url = "https://" + url
	}

	b, _ := json.Marshal(url)
	js := fmt.Sprintf("window.location.href = %s;", string(b))
	runtime.WindowExecJS(a.ctx, js)

	// Inject floating Conductino bar after the page has a chance to load.
	home := a.getHomeURL()
	go a.injectToolBarLater(home, 1200)
	go a.injectToolBarLater(home, 3000)
	go a.injectToolBarLater(home, 6000)
	return nil
}

// GoHome returns to the Conductino asset UI.
func (a *App) GoHome() error {
	if a.ctx == nil {
		return fmt.Errorf("app not started")
	}
	home := a.getHomeURL()
	b, _ := json.Marshal(home + "/")
	runtime.WindowExecJS(a.ctx, fmt.Sprintf("window.location.href = %s;", string(b)))
	return nil
}

func (a *App) injectToolBarLater(home string, delayMs int) {
	time.Sleep(time.Duration(delayMs) * time.Millisecond)
	if a.ctx == nil {
		return
	}
	js := buildInjectToolbarJS(home)
	runtime.WindowExecJS(a.ctx, js)
}

func buildInjectToolbarJS(home string) string {
	homeJSON, _ := json.Marshal(home + "/")
	// Runs in the *page* context — selection works; Cloudflare/ResearchGate render normally.
	return fmt.Sprintf(`(function(){
  try {
    if (window.__conductinoBarInstalled) return;
    var href = String(location.href || "");
    if (href.indexOf("wails.localhost") >= 0) return;
    if (href.indexOf("localhost") >= 0 && href.indexOf("34115") >= 0) return;
    window.__conductinoBarInstalled = true;
    var home = %s;
    var bar = document.createElement("div");
    bar.id = "conductino-floating-bar";
    bar.setAttribute("data-conductino", "1");
    bar.style.cssText = "all:initial;position:fixed;top:0;left:0;right:0;z-index:2147483647;"+
      "display:flex;align-items:center;gap:8px;padding:6px 10px;"+
      "background:#141a30;color:#e8ecfb;font:13px system-ui,Segoe UI,sans-serif;"+
      "box-shadow:0 2px 12px rgba(0,0,0,.35);border-bottom:1px solid #26305a;";
    function btn(label, primary){
      var b=document.createElement("button");
      b.textContent=label;
      b.style.cssText="all:initial;cursor:pointer;padding:6px 12px;border-radius:8px;"+
        "font:12px system-ui,Segoe UI,sans-serif;color:#fff;"+
        (primary?"background:#7c7bff;":"background:#1e2745;border:1px solid #26305a;color:#e8ecfb;");
      return b;
    }
    var homeBtn = btn("Conductino Home", false);
    homeBtn.onclick = function(){ location.href = home; };
    var copyBtn = btn("Copy selection", false);
    copyBtn.onclick = function(){
      var t = "";
      try { t = String(window.getSelection() || ""); } catch(e) {}
      if(!t){ alert("Select text on the page first"); return; }
      if(navigator.clipboard && navigator.clipboard.writeText){
        navigator.clipboard.writeText(t).then(function(){ copyBtn.textContent="Copied"; setTimeout(function(){copyBtn.textContent="Copy selection";},1200); });
      } else {
        prompt("Copy this text:", t);
      }
    };
    var studyBtn = btn("Selection → Study", true);
    studyBtn.onclick = function(){
      var t = "";
      try { t = String(window.getSelection() || ""); } catch(e) {}
      if(!t){ alert("Select text, or use Copy selection then open Study"); return; }
      if(navigator.clipboard && navigator.clipboard.writeText){
        navigator.clipboard.writeText(t).then(function(){
          location.href = home + "#conductino-study-clipboard";
        });
      } else {
        location.href = home + "#conductino-study-clipboard";
      }
    };
    var sumBtn = btn("Summarize selection", true);
    sumBtn.onclick = function(){
      var t = "";
      try { t = String(window.getSelection() || ""); } catch(e) {}
      if(!t){ alert("Select text on the page first"); return; }
      if(navigator.clipboard && navigator.clipboard.writeText){
        navigator.clipboard.writeText(t).then(function(){
          location.href = home + "#conductino-summarize-clipboard";
        });
      } else {
        location.href = home + "#conductino-summarize-clipboard";
      }
    };
    var hideBtn = btn("Hide", false);
    hideBtn.onclick = function(){ bar.remove(); window.__conductinoBarInstalled=false; };
    bar.appendChild(homeBtn);
    bar.appendChild(copyBtn);
    bar.appendChild(studyBtn);
    bar.appendChild(sumBtn);
    bar.appendChild(hideBtn);
    var mount = document.documentElement || document.body;
    if (mount) mount.appendChild(bar);
    document.documentElement.style.scrollPaddingTop = "44px";
  } catch (e) {}
})();`, string(homeJSON))
}
