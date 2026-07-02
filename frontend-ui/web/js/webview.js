import { DOM } from "./dom.js";

/*
submit handler
-for proxy url construction
*/
export function navigate(url) {
    DOM.iframe.src = "http://127.0.0.1:8080/api/proxy?url=" + encodeURIComponent(url);
}
/*
allows the things like the code parts controlling the history,
tab manager, and also search results section to call the navigate()
withouth duplicating the proxy url construction.
*/

export function reload() {
    DOM.iframe.src = DOM.iframe.src;
}

export function goBack() {
    try {
        DOM.iframe.contentWindow.history.back();
    }catch (err) {
        console.error(err);
    }
}

export function goForward() {
    try {
        DOM.iframe.contentWindow.history.forward();
    }catch (err) {
        console.error(err);
    }
}

export function initializeWebView() {
    DOM.backButton.onclick = goBack;
    DOM.forwardButton.onclick = goForward;
    DOM.reloadButton.onclick = reload;
}