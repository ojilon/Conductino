import { DOM } from "./dom.js";
import { navigate } from "./webview.js";

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


export function initializeNavigation() {
    DOM.urlForm.addEventListener("submit", function (e) {
        e.preventDefault();
        let input = DOM.url.value.trim();
        if(!input) return;

        let displayURL;
        if (looksLikeURL(input)) {
            const full = /^https?:\/\//i.test(input) ? input : "https://" + input;
            displayURL = full;
            
            //submit handler
            navigate(full);

        } else {
            const searchURL = "https://www.google.com/search?q=" + encodeURIComponent(input);
            displayURL = searchURL;
            
            //submit handler
            navigate(searchURL);

        }

        DOM.url.value = displayURL;
    });
}