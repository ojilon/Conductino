import { DOM } from "./dom.js";
import { BrowserState, setState } from "./state.js";

let currentEngine = "duckduckgo";

export async function initializeNavigation() {
    DOM.urlForm.addEventListener("submit", async function (e) {
        e.preventDefault();
        let input = DOM.url.value.trim();
        if(!input) return;

        setState(BrowserState.LOADING);

        try {
            const response = await fetch("/api/navigate", {
                method: "POST",
                headers: {"Content-Type": "application/json"},
                body: JSON.stringify({input: input,  engine: currentEngine})
            });

            //catch HTTP error status codes eg 400
            if (!response.ok) {
                setState(BrowserState.ERROR);
                return;
            }

            const decision = await response.json();

            if (decision.kind === "search") {
                setState(BrowserState.SEARCH);

                /*isntead of instanlty loading a heavy broken page inside the iframe,
                build an internal landing page layout insidethe iframe,
                or load the direct query with choice.
                */
                DOM.url.value = decision.query;

                //Option A: Direct load the safer proxy url
                DOM.iframe.src = "/api/proxy?url=" + encodeURIComponent(decision.url);

                //or inject custom html elements or render 'templates/search.html' here
            }else {
                setState(BrowserState.PAGE);
                DOM.url.value = decision.url;

                //perform proxy fetching
                const proxyURL = "/api/proxy?url=" + encodeURIComponent(decision.url);
                const pageResp = await fetch(proxyURL);

                //if backend flagged a cloudflare challenge
                if (pageResp.headers.get("X-Browser-State") === "CHALLENGE") {
                    setState(BrowserState.CHALLENGE);
                }

                DOM.iframe.src = proxyURL;
            }

        }catch (error) {
            setState(BrowserState.ERROR);
            reportErrorToBackend(error.message);
        }
    });
}

//helper to switch engines
export function setEnginePreference(engineName) {
    currentEngine = engineName;
}

async function reportErrorToBackend(message) {
    try {
        await fetch("/api/log-error", {
            method: "POST",
            headers: {"Content-Type": "application/json"},
            body: JSON.stringify({ error: message})
        });
    }catch (e) {
        DOM.iframe.src = "/api/proxy?url=" + encodeURIComponent("error:"+e);
    }
}