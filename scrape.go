package main

import (
	"io"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func scrapePkgPage(r io.Reader) ([]string, error) {

	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
	}

	var pkgs []string

	selector := ".pkg-dir td.pkg-name a[href]"

	doc.Find(selector).Each(func(_ int, s *goquery.Selection) {

		v, ok := s.Attr("href")
		if ok {
			if strings.HasSuffix(v, "/") {
				v = v[0 : len(v)-1]
			}
			pkgs = append(pkgs, v)
		}

	})

	return pkgs, nil

}
