package main

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/net/html"
)

// extractPkgList returns the names of all pkgs listed on the godoc http
// server's /pkg page.
func extractPkgList() ([]string, error) {
	if len(pkgPageBody) == 0 {
		return nil, errors.New("apparently no data from godoc http server /pkg")
	}

	// We could probably use a selector and do the extraction in two lines, but I've
	// never used the html pkg before, so let's try it, even if it's super brittle.

	// Read and parse the godoc http server's /pkg page.
	r := bytes.NewReader(pkgPageBody)
	doc, err := html.Parse(r)
	if err != nil {
		return nil, fmt.Errorf("failed to parse html from go doc http server /pkg")
	}

	var stdlibNode *html.Node
	var thirdPartyNode *html.Node

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "div" {
			for _, attr := range n.Attr {
				if attr.Key == "id" {
					if attr.Val == "stdlib" {
						stdlibNode = n
						return
					}
					if attr.Val == "thirdparty" {
						thirdPartyNode = n
						return
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	if stdlibNode == nil {
		// Theoretically the thirdparty section could not exist? So just verify that stdlib exists
		return nil, fmt.Errorf("failed to load stlib section from /pkg")
	}

	pkgs := extractPkgsFromNode(stdlibNode)
	pkgs = append(pkgs, extractPkgsFromNode(thirdPartyNode)...)

	return pkgs, nil
}

func extractPkgsFromNode(node *html.Node) (pkgs []string) {
	// node looks like <div id="stdlib" class="toggleVisible">
	node = nodeFirstChildElementMatch(node, elemMatch{data: "div", attrKey: "class", attrVal: "expanded"})
	if node == nil {
		return
	}
	// node is now <div class="expanded">

	node = nodeFirstChildElementMatch(node, elemMatch{data: "div", attrKey: "class", attrVal: "pkg-dir"})
	if node == nil {
		return
	}
	// node is now <div class="pkg-dir">

	node = nodeFirstChildElementMatch(node, elemMatch{data: "table"})
	if node == nil {
		return
	}
	// node is now <table>

	node = nodeFirstChildElementMatch(node, elemMatch{data: "tbody"})
	if node == nil {
		return
	}
	// node is now <tbody>

	node = nodeFirstChildElementMatch(node, elemMatch{data: "tr"})
	// consume the first <tr>, because it's a header row
	if node == nil {
		return
	}

	for node := node.NextSibling; node != nil; node = node.NextSibling {
		if node.Type != html.ElementNode || node.Data != "tr" {
			continue
		}
		// node is now a real <tr> data elem

		n := nodeFirstChildElementMatch(node, elemMatch{data: "td", attrKey: "class", attrVal: "pkg-name"})
		if n == nil {
			continue
		}
		// n is now <td class="pkg-name"> elem
		n = nodeFirstChildElementMatch(n, elemMatch{data: "a"})
		if node == nil {
			continue
		}

		// n is now of the form <a href="archive/">archive</a>
		for _, attr := range n.Attr {
			if attr.Key == "href" {
				pkg := attr.Val
				if strings.HasSuffix(pkg, "/") {
					pkg = pkg[:len(pkg)-1]
				}
				pkgs = append(pkgs, pkg)
			}
		}
	}

	return pkgs

}

// elemMatch is used to match a node
type elemMatch struct {
	data    string
	attrKey string
	attrVal string
}

func nodeFirstChildElementMatch(node *html.Node, m elemMatch) *html.Node {
	if node == nil {
		return nil
	}

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == m.data {
			if m.attrKey == "" {
				return c
			}

			for _, attr := range c.Attr {
				if attr.Key == m.attrKey && attr.Val == m.attrVal {
					return c
				}
			}
		}
	}
	return nil
}
