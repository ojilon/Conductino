export const BrowserState = {
    PAGE: "page",
    SEARCH: "search",
    LOADING: "loading",
    ERROR: "error",
    DOWNLOAD: "download",
    CHALLENGE: "challenge"
};

let currentState = BrowserState.PAGE;

export function getState() {
    return currentState;
}

export function setState(state) {
    currentState = state;
    console.log("Browser State:", state)
}

/*
skeleton usage
when Go tells the frontend:

403 Cloudflare

you'll simply call

setState(BrowserState.CHALLENGE);

When downloading a page:

setState(BrowserState.LOADING);

When showing search results:

setState(BrowserState.SEARCH);
*/