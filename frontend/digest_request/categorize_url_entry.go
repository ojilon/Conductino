package digest_request


type NavigationRequest struct {
	Input string `json:"input"`
	Engine string `json:"engine"` //user can specify the engine
}

type NavigationResponse struct {
	Kind string `json:"kind"`
	URL string `json:"url"`
	Query string `json:"query,omitempty"`
	Engines map[string]string `json:"engines,omitempty"`
}

//map of supported engines
var SearchEngines = map[string]string{
	"google": "https://www.google.com/search?=",
	"duckduckgo": "https://duckduckgo.com/html/?q=",
	"bing": "https://www.bing.com/search?q=",
}

