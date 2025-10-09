package service

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

func extractGradesLinks(htmlContent []byte) (links []string, err error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewBuffer(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML content: %v", err)
	}

	doc.Find("table#overview-grade tbody tr").Each(func(i int, tr *goquery.Selection) {
		if tr.HasClass("emptyrow") {
			return
		}
		linkSel := tr.Find("td.c0 a").First()

		href, _ := linkSel.Attr("href")
		links = append(links, href)
	})

	return
}

func extractItems(htmlContent []byte) (courseName string, rows [][]string, err error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewBuffer(htmlContent))
	if err != nil {
		return "", nil, fmt.Errorf("failed to parse HTML content: %v", err)
	}

	courseName = doc.Find("div.page-header-headings h1").First().Text()
	if courseName == "" {
		return "", nil, fmt.Errorf("failed to extract course name")
	}

	doc.Find("table.user-grade tbody tr").Each(func(i int, tr *goquery.Selection) {
		isAggregation := tr.Find("span[title='Aggregation']").First()
		if isAggregation.Length() > 0 {
			return
		}

		var row []string
		thName := tr.Find("th").First().Find("div.rowtitle").Children().First().Text()
		// slog.Debug("Extracted thname", "thname", thName)

		if thName == "" {
			return
		}
		row = append(row, trim(thName))

		tr.Find("td").Each(func(i int, s *goquery.Selection) {
			row = append(row, (firstTextNode(s)))
		})
		if len(rows) == 7 {
			rows = append(rows, row)
		}
		// slog.Debug("Extracted row", "row", rows)
	})

	return
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
