package navigation_helper

import (
	"compress/flate"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/cookiejar"
	"time"
)

type Browser struct {
	Client *http.Client
}

func NewBrowser() *Browser {
	jar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}

	client := &http.Client{
		Jar:     jar,
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			//useful for debugging redirect chains.
			//returning nil means "follow the redirect".
			return nil
		},
	}

	return &Browser{
		Client: client,
	}
}

func (b *Browser) Do(req *http.Request) (*http.Response, error) {
	resp, err := b.Client.Do(req)
	if err != nil {
		return nil, err
	}

	//Intercept and decompress based on the server's respoonse header
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		gzReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			resp.Body.Close()
			return nil, err
		}

		//wrap the response body in a custom read closer to prevent memory leaks
		resp.Body = &decompressedReadCloser{reader: gzReader, closer: resp.Body}

		//strip  the header so downstream code (and the browser) knows it's uncompressed
		resp.Header.Del("Content-Encoding")

	case "deflate":
		flateReader := flate.NewReader(resp.Body)
		resp.Body = &decompressedReadCloser{reader: flateReader, closer: resp.Body}
		resp.Header.Del("Content-Encoding")
	}

	return resp, nil
}

// Helper structure to close both the decompressor and the underlying network stream well
type decompressedReadCloser struct {
	reader io.Reader
	closer io.ReadCloser
}

// Close implements [io.ReadCloser].
func (d *decompressedReadCloser) Close() error {
	if c, ok := d.reader.(io.Closer); ok {
		c.Close()
	}
	return  d.closer.Close()
}

func (d *decompressedReadCloser) Read(p []byte) (n int, err error) {
	return d.reader.Read(p)
}


/*notes
Centralized Protection: 
By intercepting the stream inside Browser.Do, 
your ProxyHandler code doesn't have to care about
 compression layers at all. It will always receive clean,
  plain-text bytes.

Double Closing Guard: The custom decompressedReadCloser struct
 guarantees that when your ProxyHandler calls defer resp.Body.Close(), 
 it shuts down both the internal decompression buffers and the underlying 
 network socket cleanly—preventing file descriptor leaks.

 Safe Header Trimming: Explicitly removing resp.Header.Del("Content-Encoding")
  stops your proxy from confusing the frontend browser iframe.
*/