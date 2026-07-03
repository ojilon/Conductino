package handlers

import (
	"io"
	"log"
	"net/http"
	"net/url"
)

//--------------api/proxy -----------------------
func (c *BackendClient) ProxyHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "GET required", http.StatusMethodNotAllowed)
        return
    }

    // 1. Read ?url=
    targetURL := r.URL.Query().Get("url")
    if targetURL == "" {
        http.Error(w, "missing url parameter", http.StatusBadRequest)
        return
    }

    //validate URL
    if _, err := url.ParseRequestURI(targetURL); err != nil {
        http.Error(w, "invalid URL", http.StatusBadRequest)
        return
    }

    /*proposed future refactor
    func NewHTTPClient() *http.Client {

        return &http.Client{

            CheckRedirect: func(req *http.Request, via []*http.Request) error {

                log.Println("Redirect ->", req.URL.String())

                return nil
            },
        }
    }

    then client becomes
    client := NewHTTPClient()

    to lATER add things like 
    cookies, proxy support, custom User-Agent, TLS settings without
    touching the function ProxyHAndler
    */

    //create the request
    req, err := http.NewRequest(
        http.MethodGet,
        targetURL,
        nil,
    )   
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadGateway)
        return
    }

    // Set required custom headers
    req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0 Safari/537.36")
    req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
    req.Header.Set("Accept-Language", "en-US,en;q=0.9")
    //req.Header.Set("Accept-Encoding", "gzip, deflate, br")

    //execute request to get response
    resp, err := c.Browser.Do(req)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadGateway)
        return
    }
    defer resp.Body.Close()

    //response debugging added
    log.Println("Final URL     :", resp.Request.URL.String())
    log.Println("Status        :", resp.Status)
    log.Println("Content-Type  :", resp.Header.Get("Content-Type"))

    // 3. Remove headers that stop embedding in Webview
    resp.Header.Del("X-Frame-Options")
    // TODO:
    // Instead of deleting CSP completely,
    // rewrite it where possible.
    // Removing it permanently reduces security.
    resp.Header.Del("Content-Security-Policy")


    // Stream body directly
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    //handle rewriting links to the resources of downloaded response
    contentType := resp.Header.Get("Content-Type")
    action := DecideResponse(contentType)
    switch action {
    case ActionDisplayHTML:
        baseURL := resp.Request.URL
        body, err = RewriteHTML(body, baseURL)
        if err != nil {
            log.Println("HTML rewrite error: ",err)
        }

        //copy headers, skip the ones that break modified payloads
        for k, values := range resp.Header {
            if k == "Content-Length" || k == "Content-Encoding" || k == "Content-Type" {
                continue
            }
            for _, v := range values {
                w.Header().Add(k, v)
            }
        }

        //force the correct content type for rewritten HTMl
        w.Header().Set("Content-Type", "text/html; charset=utf-8")
        //write status before the body transmission
        w.WriteHeader(http.StatusOK)

        _, err = w.Write(body)
        if err != nil {
            log.Println("Write error:", err)
            return
        }
    case ActionStream:
        //safe copy for non-breaking headers
        for k, values := range resp.Header {
            if k == "Content-Length" {
                continue
            }
            for _, v := range values {
                w.Header().Add(k, v)
            }
        }

        w.Header().Set("Content-Type", contentType)
        w.WriteHeader(http.StatusOK)

        _, err = w.Write(body)
        if err != nil {
            log.Println("Write error (Stream):", err)
            return
        }
    case ActionDownload:
        for k, values := range resp.Header {
            if k == "Content-Length" || k == "Content-Disposition" {
                continue
            }
            for _, v := range values {
                w.Header().Add(k, v)
            }
        }

        w.Header().Set("Content-Type", contentType)
        w.Header().Set("Content-Disposition", "attachment")
        w.WriteHeader(http.StatusOK)

        _,err = w.Write(body)
        if err != nil {
            log.Println("Write error (Download):", err)
            return
        }
    default:
        http.Error(w, "Unsupported proxy action" , http.StatusBadRequest)
        return
    }
}