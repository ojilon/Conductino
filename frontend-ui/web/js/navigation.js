import { DOM } from "./dom.js";

export async function initializeNavigation() {
    DOM.urlForm.addEventListener("submit", async function (e) {
        e.preventDefault();
        let input = DOM.url.value.trim();
        if(!input) return;

        try {
            const response = await fetch("/api/navigate", {
                method: "POST",
                headers: {
                    "Content-Type": "application/json"
                },
                body: JSON.stringify({input: input})
            });

            //catch HTTP error status codes eg 400
            if (!response.ok) {
                DOM.iframe.src = "/api/proxy?url=" + encodeURIComponent("error-fallback");
                return;
            }

            const decision = await response.json();

            //check if Go backend sent a successful structurebut with internal error
            if (decision.error) {
                DOM.iframe.src = "/api/proxy?url=" + encodeURIComponent("error:"+decision.error);
                return;
            }

            DOM.url.value = decision.url;
            DOM.iframe.src = "/api/proxy?url=" + encodeURIComponent(decision.url);
        }catch (error) {
            reportErrorToBackend(error.message);
        }
    });
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