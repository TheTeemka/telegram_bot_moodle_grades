package main

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

func ExtractGradesLinks(htmlContent []byte) ([]string, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewBuffer(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML content: %v", err)
	}

	var links []string
	doc.Find("table#overview-grade tbody tr").Each(func(i int, tr *goquery.Selection) {
		if tr.HasClass("emptyrow") {
			return
		}
		linkSel := tr.Find("td.c0 a").First()

		href, _ := linkSel.Attr("href")
		links = append(links, href)
	})

	return links, nil
}

func ExtractItems(htmlContent []byte) (string, [][]string, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewBuffer(htmlContent))
	if err != nil {
		return "", nil, err
	}

	name := doc.Find("div.page-header-headings h1").First().Text()
	if name == "" {
		panic("course name is empty")
	}
	var items [][]string
	doc.Find("table.user-grade tbody tr").Each(func(i int, tr *goquery.Selection) {
		isAggregation := tr.Find("span[title='Aggregation']").First()
		if isAggregation.Length() > 0 {
			return
		}

		var rows []string
		thName := tr.Find("th").First().Find("div.rowtitle").Children().First().Text()
		// slog.Debug("Extracted thname", "thname", thName)

		if thName == "" {
			return
		}
		rows = append(rows, trim(thName))

		tr.Find("td").Each(func(i int, s *goquery.Selection) {
			rows = append(rows, (firstTextNode(s)))
		})
		if len(rows) == 7 {
			items = append(items, rows)
		}
		// slog.Debug("Extracted row", "row", rows)
	})

	return name, items, nil
}
func trim(s string) string {
	s = strings.ReplaceAll(s, "\t", "")
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.TrimSpace(s)
	return s
}

func firstTextNode(s *goquery.Selection) string {
	// recursive helper that searches depth-first for the first non-empty text node
	var find func(n *html.Node) string
	find = func(n *html.Node) string {
		if n == nil {
			return ""
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.TextNode {
				txt := strings.TrimSpace(c.Data)
				if txt != "" {
					return txt
				}
			}
			if res := find(c); res != "" {
				return res
			}
		}
		return ""
	}

	for _, n := range s.Nodes {
		if txt := find(n); txt != "" {
			// slog.Debug("firstTextNode: found text node", "text", txt)
			return trim(txt)
		}
	}

	// slog.Debug("firstTextNode: no text node found", "text", s.Text())
	return trim(s.Text())
}
