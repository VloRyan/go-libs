package httpx

import (
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"path"

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
	var headNode *html.Node
	for c := doc.FirstChild; c != nil; {
		if c.Type == html.ElementNode && c.Data == "html" {
			c = c.FirstChild
			continue
		}
		if c.Type == html.ElementNode && c.Data == "head" {
			headNode = c
			break
		}
		c = c.NextSibling
	}
	if headNode == nil {
		return "", errors.New("could not find head node")
	}
	for c := headNode.FirstChild; c != nil; c = c.NextSibling {
		for i := range c.Attr {
			if c.Attr[i].Key == "src" || c.Attr[i].Key == "href" {
				c.Attr[i].Val = path.Join(assetPath, c.Attr[i].Val)
			}
		}
	}
	if serverData != "" {
		firstChild := headNode.FirstChild
		scriptContentNode := &html.Node{
			Type: html.TextNode,
			Data: "\n        window.SERVER_DATA = " + serverData + "\n    ",
		}
		newLineNode := &html.Node{
			Type: html.TextNode,
			Data: "\n    ",
		}
		scriptNode := &html.Node{
			Parent:      headNode,
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
		headNode.FirstChild = newLineNode
		//	indexHtml = strings.Replace(indexHtml, "__SERVER_DATA__", string(b), 1)
	}
	w := NewInMemResponseWriter()
	if err := html.Render(w, doc); err != nil {
		return "", err
	}
	return string(w.Body), nil
}
