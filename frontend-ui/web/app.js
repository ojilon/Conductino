// app.js — Runs INSIDE WebView2. Pure browser JS.
//
// Responsibilities:
//   1. Capture text selections inside the WebView pane.
//   2. Show a floating highlight toolbar.
//   3. POST a "Note Highlight Event" JSON packet to the Go IPC router.
//   4. GET search results and render them in the Study Sidebar.
//
// SECURITY (Local Context Isolation):
//   This script can ONLY reach the backend via fetch('/api/...') against
//   the same origin (127.0.0.1:8080). It has NO direct access to SQLite,
//   the file system, or the Zig process. The Go router is the gatekeeper.

import { DOM } from "./js/dom";

(function () {
    "use strict";

    let currentSelection = null; //{text, context, rect, coords}

    //1.URL bar
    /*
    checking whether the input looks like a web url, eg has a domain suffix
    like .com or starts with http, if not, redirect it to a search ening like Google.
    */
    function looksLikeURL(input) {
        // add protocol if missing so URL() can parse it
        const withProto = /^https?:\/\//i.test(input) ? input : "https://" + input;
        try {
            const u = new URL(withProto);
            // URL() accepts things like "https://word" (no TLD) so add a
            // small extra check: hostname must have a dot OR be localhost
            return u.hostname.includes(".") || u.hostname === "localhost";
        } catch (_) {
            return false;
        }
    }

    DOM.urlForm.addEventListener("submit", function (e) {
        e.preventDefault();
        let input = DOM.url.value.trim();
        if(!input) return;

        let displayURL;
        if (looksLikeURL(input)) {
            const full = /^https?:\/\//i.test(input) ? input : "https://" + input;
            displayURL = full;
            DOM.iframe.src = "http://127.0.0.1:8080/api/proxy?url=" + encodeURIComponent(full);
        } else {
            const searchURL = "https://www.google.com/search?q=" + encodeURIComponent(input);
            displayURL = searchURL;
            DOM.iframe.src = "http://127.0.0.1:8080/api/proxy?url=" + encodeURIComponent(searchURL);
        }

        DOM.url.value = displayURL;
    });


    DOM.btnBack.onclick = function () { DOM.iframe.contentWindow.history.back(); };
    DOM.btnFwd.onclick = function () {history.forward(); };
    DOM.btnRel.onclick = function () {DOM.iframe.src = DOM.iframe.src; };

    /*2. Selection capture inside iframe

    Because of same origin restrictions, this only works fully when the 
    iframe loads a same-origin document (e.g. oor own help pages or a 
    page fetched and rewritten by /api/archive). For the external pages
    we fall back to listening to the host document's selectionchange.
    */

    function pickUpSelection(doc) {
        const sel = doc.getSelection();
        if(!sel || sel.isCollapsed) { hideToolbar(); return;}

        const text = sel.toString().trim();
        if (text.length < 3) { hideToolbar(); return; }

        const range =sel.getRangeAt(0);
        const rect  = range.getBoundingClientRect();

        currentSelection = {
            text: text,
            context: getContext(range),
            coords: {
                start_x: Math.round(rect.left),
                start_y: Math.round(rect.top),
                end_x: Math.round(rect.right),
                end_y: Math.round(rect.bottom)
            }
        };

        showToolbar(rect);
    }

    function getContext(range) {
        //grab ~120 chars of surrounding text for FTS5 context snippets.
        try {
            const container = range.commonAncestorContainer;
            const text = (container.textContent || "");
            const idx = text.indexOf(range.length.toString());
            if (idx < 0) return "";
            const start = Math.max(0, idx-60);
            const end =  Math.min(text.length, idx + range.toString().length + 60);
            return text.slice(start, end).replace(/\s+/g, " ").trim();
        }catch (_) {return ""; }
    }

    function showToolbar(rect) {
        //position the toolbar above the selection (inside the iframe).
        const paneRect = document.getElementById("webiew-pane").getBoundingClientRect();
        DOM.toolbar.style.left  = (paneRect.left + rect.left + rect.width / 2) + "px";
        DOM.toolbar.style.top = (paneRect.top + rect.top - 44) + "px";
        DOM.toolbar.classList.remove("hidden");
    }
    function hideToolbar() {
        DOM.toolbar.classList.add("hidden");
        currentSelection = null;
    }

    //Try to attach to the iframe doc when it loads(same origin only).
    DOM.iframe.addEventListener("load", function () {
        //update the url bar when iframe navigates itself
        try {
            const loc = DOM.iframe.contentDocument.location.href;
            if (!loc && loc !== "about:blank") {
                DOM.url.value = loc;
            }
        } catch (_) { /*corss-orogin - ignored */}
    });
    //Also listen on the host document (works for archived / local pages).
    document.addEventListener("selectionchange", function () {pickUpSelection(document); });

    //3. Highlight toolbar buttons -> POST /api/save_notes --------------------------
    Array.prototype.forEach.call(DOM.toolbar.querySelectorAll("button[data-color]"),
        function (btn) {
          btn.addEventListener("click", function () {
            saveNote(btn.getAttribute("data-color"));
          });
        });
    document.getElementById("btn-add-note").addEventListener("click", function () {
        const note = window.prompt("Add a note for this highlight:");
        saveNote("#00add8", note || "");
      });

    function saveNote(color, extraNote) {
        if(!currentSelection) return;
        const packet = {
            page_url:   DOM.iframe.src,
            page_title: DOM.iframe.contentDocument ? DOM.iframe.contentDocument.title : "",
            selection:  currentSelection.text,
            context:    (extraNote ? extraNote + " · " : "") + currentSelection.context,
            color:      color,
            coords:     currentSelection.coords,
            created_at: Date.now()
        };

        // ------- IPC call: web -> web -> Go -> Zig -> SQLite --------------
        fetch("/api/save_note", {
            method: "POST",
            headers: {"Content-Type": "application/json"},
            body: JSON.stringify(packet)
        })
        .then(function (r) {
            if (!r.ok) throw new Error("HTTP " + r.status);
            return r.json();
        })
        .then(function (saved) {
            perpendHighlight(saved.result || packet);
            hideToolbar();
        })
        .catch(function (err) {
            console.error("save_note failed:", err);
            DOM.statusD.classList.add("down");
        });
    }

    //4. Search -> GET api/search -------------------
    DOM.searchF.addEventListener("submit", function (e){
        e.preventDefault();
        const q = DOM.search.value.trim();
        if (!q) return;
        fetch("/api/search?query=" + encodeURIComponent(q))
          .then(function (r) {return r.json(); })
          .then(function (data) {
            DOM.list.innerHTML = "";
            (data.results || []). forEach(perpendHighlight);
          });
    });

    //5. Archive button -> POST /api/archive 
    DOM.btnArc.addEventListener("click", function () {
        fetch("/api/archive", {
            method: "POST",
            headers: {"Content-Type": "application/json"},
            body:   JSON.stringify({url: DOM.iframe.src })
        }).then(function () {console.log("archived"); });
    });

    //Render helper
    function perpendHighlight(h) {
        const li = document.createElement("li");
        li.style.borderLeftColor = h.color || "#e94560";
        li.innerHTML = 
           '<div class="selection">"' + escapeHTML(h.selection) + '"</div>' +
           '<di class="ctx">' + escapeHTML(h.context || "") + '</div>' +
           '<div class="meta">' +
             '<span>' + escapeHTML(h.page_title || h.page_url) + '</span>' + 
             '<span>' + new Date(h.created_at || Date.now()).toLocaleTimeString()+ '<span>' +
            '</div>';
        DOM.list.insertBefore(li, DOM.list.firstChild);
    }
    function escapeHTML(s) {
        return String(s).replace(/[&<>"']/g, function (c) {
            return ({"&":"&amp;","<":"&lt;",">":"&gt;","\"":"&quot;","'":"&#39;"})[c];
        });
    }

    // Heartbeat
    setInterval(function () {
        fetch("/api/search?query=__ping__")
        .then(function (r) { DOM.statusD.classList.toggle("down", !r.ok); })
        .catch(function () {DOM.statusD.classList.add("down"); });
    }, 10000);

})();
