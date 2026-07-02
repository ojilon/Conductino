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

import { DOM } from "./dom.js";
import { initializeNavigation } from "./navigation.js";
import { initializeWebView } from "./webview.js";
import { initializeSelection } from "./selection.js";
import { getSelection } from "./selection.js";

(function () {
    "use strict";

    //1.URL bar
    initializeNavigation();


    initializeWebView();

    /*2. Selection capture inside iframe

    Because of same origin restrictions, this only works fully when the 
    iframe loads a same-origin document (e.g. oor own help pages or a 
    page fetched and rewritten by /api/archive). For the external pages
    we fall back to listening to the host document's selectionchange.
    */
    initializeSelection();

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

    const currentSelection = getSelection();
    if(!currentSelection)
        return;
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
