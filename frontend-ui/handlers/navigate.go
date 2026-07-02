package handlers

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
)

type NavigationKind int

const (
	NavWebsite NavigationKind = iota
	NavSearch
	NavInternal
	NavFile
)

//for browser returning structured decisions instead of true/false
type NavigationDecision struct {
	Kind NavigationKind
	Input string
	URL string
	Query string
}

func (k NavigationKind) String() string {
	switch k {
	case NavWebsite:
		return  "website"
	case NavSearch:
		return "search"
	case NavFile:
		return "file"
	default:
		return "unknown"
	}
} 

type NavigationRequest struct {
	Input string `json:"input"`
}

type NavigationResponse struct {
	Kind string `json:"kind"`
	URL string `json:"url"`
	Query string `json:"query,omitempty"`
}

const DefualtSearchEngine = "https://www.google.com/search?q="

func DetectNavigation(input string) NavigationDecision {
	input = strings.TrimSpace(input)
	if input == "" {
		return NavigationDecision{}
	}

    /*
    Handle things like

    browser://history

    browser://downloads

    browser://settings

    browser://bookmarks
    */
	if strings.HasPrefix(input, "browser://"){
		return NavigationDecision{
			Kind: NavInternal,
			Input: input,
			URL: input,
		}

	}

	candidate := input
	if !strings.HasPrefix(candidate, "http://") && !strings.HasPrefix(candidate, "https://") {
		candidate = "https://" + candidate
	}

	if u, err := url.Parse(candidate); err == nil {
		host := u.Hostname()
		if strings.Contains(host, ".") || host == "localhost" {
			return NavigationDecision{
				Kind: NavWebsite,
				Input: input,
				URL: candidate,
			}
		}
	}

    //everything else is a search
	return NavigationDecision{
		Kind: NavSearch,
		Input: input,
		Query: input,
		URL: DefualtSearchEngine + url.QueryEscape(input),
	}
}


/*
illustration
This handler will receive:

{
    "input":"photosynthesis"
}

and return

{
    "kind":"search",
    "url":"https://www.google.com/search?q=photosynthesis"
}
*/
func (c *BackendClient) DetectNavigationHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed,)
		return
	}

	var req NavigationRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest,)
		return
	}

	decision := DetectNavigation(req.Input)
	resp := NavigationResponse{
		Kind: decision.Kind.String(),
		URL: decision.URL,
		Query: decision.Query,
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	json.NewEncoder(w).Encode(resp)

}