package httpx

import (
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"path"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func Origin(req *http.Request) string {
	scheme := "http"
	if req.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + req.Host
}

func FullURL(req *http.Request) string {
	url := Origin(req) + req.URL.Path
	if req.URL.RawQuery != "" {
		url += "?" + req.URL.RawQuery
	}
	return url
}

// GenerateReplacedIndexHTML loads index.html in fSys file and prefixes src and href node attributes with assetPath.
// If serverData is provided a script which defines window.SERVER_DATA will be added to head.
func GenerateReplacedIndexHTML(fSys fs.FS, assetPath string, serverData string) (string, error) {
	f, err := fSys.Open("index.html")
	if err != nil {
		return "", err
	}
	doc, err := html.Parse(f)
	if err != nil {
		fmt.Println("Error:", err)
		return "", nil
	}
	var head *html.Node
	var body *html.Node
	current := doc.FirstChild
	for current != nil {
		if current.Type == html.ElementNode && current.Data == "html" {
			current = current.FirstChild
			continue
		}
		if current.Type == html.ElementNode && current.Data == "head" {
			head = current
		}
		if current.Type == html.ElementNode && current.Data == "body" {
			body = current
		}
		current = current.NextSibling
	}
	if head == nil {
		return "", errors.New("could not find head node")
	}
	prefixSrcHrefAttribs(head, assetPath)
	if body != nil {
		prefixSrcHrefAttribs(body, assetPath)
	}
	if serverData != "" {
		firstChild := head.FirstChild
		scriptContentNode := &html.Node{
			Type: html.TextNode,
			Data: "\n        window.SERVER_DATA = " + serverData + "\n    ",
		}
		newLineNode := &html.Node{
			Type: html.TextNode,
			Data: "\n    ",
		}
		scriptNode := &html.Node{
			Parent:      head,
			FirstChild:  scriptContentNode,
			LastChild:   scriptContentNode,
			PrevSibling: newLineNode,
			NextSibling: firstChild,
			Type:        html.ElementNode,
			DataAtom:    atom.Script,
			Data:        "script",
			Attr:        []html.Attribute{{Key: "type", Val: "application/javascript"}},
		}
		newLineNode.NextSibling = scriptNode
		firstChild.PrevSibling = scriptNode
		scriptContentNode.Parent = scriptNode
		head.FirstChild = newLineNode
	}
	w := NewInMemResponseWriter()
	if err := html.Render(w, doc); err != nil {
		return "", err
	}
	return string(w.Body), nil
}

func prefixSrcHrefAttribs(node *html.Node, prefix string) {
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		for i := range c.Attr {
			if c.Attr[i].Key == "src" || c.Attr[i].Key == "href" && !isExternalPath(c.Attr[i].Val) {
				c.Attr[i].Val = path.Join(prefix, c.Attr[i].Val)
			}
		}
	}
}

func isExternalPath(path string) bool {
	return strings.HasPrefix(path, "https://") || strings.HasPrefix(path, `http://`)
}
