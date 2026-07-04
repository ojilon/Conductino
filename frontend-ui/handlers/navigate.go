package handlers

import (
	"encoding/json"
	"log"
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
	case NavInternal:
		return "internal"
	case NavFile:
		return "file"
	default:
		return "unknown"
	}
} 

type NavigationRequest struct {
	Input string `json:"input"`
	Engine string `json:"engine"`//allow user to specify engine
}

type NavigationResponse struct {
	Kind string `json:"kind"`
	URL string `json:"url"`
	Query string `json:"query,omitempty"`
	Engines map[string]string `json:"engines,omitempty"`
}

//map of supported search engines
var SearchEngines = map[string]string{
	"google": "https://www.google.com/search?=",
	"duckduckgo": "https://duckduckgo.com/?q=",
	"bing": "https://www.bing.com/search?q=",
}

func DetectNavigation(input string, selectEngine string) NavigationDecision {
	input = strings.TrimSpace(input)
	if input == "" {
		return NavigationDecision{}
	}

	if strings.HasPrefix(input, "browser://"){
		return NavigationDecision{
			Kind: NavInternal,
			Input: input,
			URL: input,
		}

	}

	//check if its explicit website navigation
	candidate := input
	if !strings.HasPrefix(candidate, "http://") && !strings.HasPrefix(candidate, "https://") {
		candidate = "https://" + candidate
	}

	if u, err := url.Parse(candidate); err == nil {
		host := u.Hostname()
		if (strings.Contains(host, ".") && !strings.Contains(input, " ")) || host == "localhost" {
			return NavigationDecision{
				Kind: NavWebsite,
				Input: input,
				URL: candidate,
			}
		}
	}

	//Defualt fallbakc search engine if invalid or unspecified
	baseURL, exists := SearchEngines[strings.ToLower(selectEngine)]
	if !exists {
		baseURL = SearchEngines["duckduckgo"]
	}

    //everything else is a search
	return NavigationDecision{
		Kind: NavSearch,
		Input: input,
		Query: input,
		URL: baseURL + url.QueryEscape(input),
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
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		log.Println("Error: POST method required, got: ", r.Method)
		return
	}

	var req NavigationRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Println("Bad request: ", err)
		return
	}

	decision := DetectNavigation(req.Input, req.Engine)
	resp := NavigationResponse{
		Kind: decision.Kind.String(),
		URL: decision.URL,
		Query: decision.Query,
		Engines: map[string]string{
			"Google": SearchEngines["google"] + url.QueryEscape(decision.Query),
			"DuckDuckGo": SearchEngines["duckduckgo"] + url.QueryEscape(decision.Query),
			"Bing": SearchEngines["bing"] + url.QueryEscape(decision.Query),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) //tell js request succeeded

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Println("Error encoding json response: ", err)
		//header already sent, so no use of http.Error here
	}

}