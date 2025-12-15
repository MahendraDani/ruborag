// Parse html file, strip html tags
package parser

import (
	"bufio"
	"os"
	"strings"

	"golang.org/x/net/html"
)

// recursively walk the HTML node tree
// and collect text nodes while ignoring scripts/styles.
func extractText(n *html.Node, sb *strings.Builder) {
	if n.Type == html.TextNode {
		text := strings.TrimSpace(n.Data)
		if text != "" {
			sb.WriteString(text)
			sb.WriteString(" ")
		}
	}

	// Skip script and style tags entirely
	if n.Type == html.ElementNode && (n.Data == "script" || n.Data == "style") {
		return
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		extractText(c, sb)
	}
}

func RemoveHTMLTagsFromFile(inputPath string) (string, error) {
	file, err := os.Open(inputPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	doc, err := html.Parse(reader)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	extractText(doc, &sb)

	text := strings.Join(strings.Fields(sb.String()), " ")
	return text, nil
}

func RemoveHTMLTagsFromFileUnbuffered(inputPath string) (string, error) {
	file, err := os.Open(inputPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	doc, err := html.Parse(file)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	extractText(doc, &sb)

	// normalize whitespace
	text := strings.Join(strings.Fields(sb.String()), " ")
	return text, nil
}
