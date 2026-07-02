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
                const errorData = await response.json().catch(() => ({}));
                throw new Error(errorData.error || `Server responded with status ${response.status}`);
            }

            const decision = await response.json();

            //check if Go backend sent a successful structurebut with internal error
            if (decision.error) {
                throw new Error(decision.error);
            }

            DOM.url.value = decision.url;
            DOM.iframe.src = "/api/proxy?url=" + encodeURIComponent(decision.url);
        }catch (error) {
            //handle the error visually in the frontend
            console.error("Navigation error:", error);
            showErrorToUser(error.message);
        }
    });
}

//helper function to display the error to the user
function showErrorToUser(message) {
    alert("Navigation Failed: " + message);
}