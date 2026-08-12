package digest_request

import (
	"log"
	"net/url"
	"strings"
)

type NavigationKind int

const (
	NavWebsite NavigationKind = iota
	NavPlainText
	NavLocal
)

func (k NavigationKind) String() string {
	switch k {
	case NavWebsite:
		return "Website"
	case NavPlainText:
		return "PlainText"
	case NavLocal:
		return "SearchLocal"
	default:
		return "SearchLocal"
	}
}

//for browser returningn structured decisions instead of true/false
type NavigationDecision struct {
	Kind NavigationKind
	Input string
	URL string
	Query string
}

func FindoutNavigationDecided(input string, selectEngine string) NavigationDecision {
	input = strings.TrimSpace(input)
	if input == "" {
		return NavigationDecision{}
	}

	/*proposed improvements
	-> making sure if website, correct the typos
	eg - ww.cell
	   -w.cell.com
	   -cell.com
	   -www.cell.com
	should alteast cause the option to reach for a website, not plain text
	*/

	//check for "https://"
	specimen := input
	if !strings.Contains(specimen, "://") {
		specimen = "https://" + specimen
	}

	if u, err := url.Parse(specimen); err == nil {
		host := u.Hostname()
		if(strings.Contains(host, ".") && !strings.Contains(input, " ")) || host == "localhost" {

			return NavigationDecision{
				Kind: NavWebsite,
				Input: input,
				URL: specimen,
			}
		}
	}

	//Default search engine if invalid or unspecified
	baseURL, exists := SearchEngines[strings.ToLower(selectEngine)]
	if !exists {
		baseURL = SearchEngines["duckduckgo"]
	}

	//everything else is taken as plain text
	NavigationDecision_ := NavigationDecision{
		Kind: NavPlainText,
		Input: input,
		Query: input,
		URL: baseURL + url.QueryEscape(input),
	}

	log.Println("Plain text search: ", NavigationDecision_.URL)
	return  NavigationDecision_
}