package handlers

import (
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
		Jar: jar,
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			//useful for debugging redirect chains.
			//returning nil means "follow the redirect".
			return  nil
		},
	}

	return &Browser{
		Client: client,
	}
}

func (b *Browser) Do(req *http.Request) (*http.Response, error) {
	return b.Client.Do(req)
}
/*
Allows complexity to easily be adopted in log requests, throttle them,
inject headers, collect timing statistics: all these will require modifying
the Browser.Do() alone
*/