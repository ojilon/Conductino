package handlers

import (
	"bytes"
	"log"
	"net/url"

	"golang.org/x/net/html"
)

//RewriteHTML parses an HTML document, rewrites URLs,
//then retruns the modified HTMl.
func RewriteHTML(body []byte, baseURL *url.URL) ([]byte, error){

	//parse HTML into a node tree
	doc, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	//walk every node
	walk(doc, baseURL)

	//convert tree back into html
	var out bytes.Buffer
	if err := html.Render(&out, doc); err != nil {
		return  nil, err
	}

	return out.Bytes(), nil
}

/*
More items to look out for
<base href="..."> tags, which change the base URL for all relative links.
CSS resources, including url(...) inside stylesheets and inline <style> blocks.
srcset on responsive images.
JavaScript-generated URLs from fetch(), XMLHttpRequest, or dynamic DOM manipulation.
Cookies, redirects, and other HTTP behaviors.
Additional HTML attributes like poster, data-*, and meta refresh URLs.
*/

func walk(node *html.Node, baseURL *url.URL) {

	//Only care about the html elements
	if node.Type == html.ElementNode {
		switch node.Data {
		case "img", "script", "iframe", "video", "audio", "source":
			rewriteAttribute(node, "src", baseURL)
		case "link", "a":
			rewriteAttribute(node, "href", baseURL)
		case "form":
			rewriteAttribute(node, "action", baseURL)
		}
	}

	//Visit every child
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		walk(child, baseURL)
	}
}

func rewriteAttribute(node *html.Node, attrName string, baseURL *url.URL) {
	
	for i := range node.Attr {
		attr := &node.Attr[i]
		if attr.Key != attrName {
			continue
		}

		//Skip empty values
		if attr.Val == ""{
			return
		}

		//parse URl found in html
		ref, err := url.Parse(attr.Val)
		if err != nil {
			return
		}

		//resolve relative URLs
		absolute := baseURL.ResolveReference(ref)

		//rewrite through the proxy
		attr.Val = "/api/proxy?url=" + url.QueryEscape(absolute.String())

		log.Println(attrName, "->", absolute.String())

		return
	}
}