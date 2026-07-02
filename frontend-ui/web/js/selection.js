import { DOM } from "./dom.js";


let currentSelection = null; //{text, context, rect, coords}

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

export function getSelection() {
    return currentSelection;
}

export function initializeSelection() {
    DOM.iframe.addEventListener("load", function (){
        try {
            const loc = DOM.iframe.contentDocument.location.href;
            if(loc && loc !== "about:blank")
                DOM.url.value = loc;
        }catch (_) {}
    });

    document.addEventListener("selectionchange", function() {
        pickUpSelection(document);
    })
}