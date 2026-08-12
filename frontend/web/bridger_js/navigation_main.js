import { DOM } from "./dom.js";

let currentEngine = "duckduckgo";

export async function ManifestNavigation() {
    DOM.urlForm.addEventListener("submit", async function (e) {
        e.preventDefault();

        let input = DOM.url.value.trim();
        if (!input) return;

        try {
            const response = await fetch ()
        }
    } );
}