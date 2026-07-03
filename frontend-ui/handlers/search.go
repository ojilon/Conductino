package handlers

import (
	"net/http"
	"strings"
)

// ---------/api/search -------------------------------------------
//GEt ?query=memory - passes through to the Zig FTS5 endpoint.
func (c *BackendClient) SearchHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "GET required", http.StatusMethodNotAllowed)
        return
    }
    q := r.URL.Query().Get("query")
    if strings.TrimSpace(q) == "" {
        http.Error(w , "query param required", http.StatusBadRequest)
        return
    }
    //query is already in r.URL.RawQuesry; forward will pass it along.
    c.forward(w, r, "/search")
}
