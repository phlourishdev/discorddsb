package htmlparser

import (
	"golang.org/x/net/html"
	"io"
	"log"
	"net/http"
	"strings"
)

// classEntry represents a single class entry with its name and the periods entries.
type classEntry struct {
	Name          string
	PeriodEntries [][]string
}

// allClassEntries holds all class entries parsed from the HTML.
var allClassEntries []classEntry

// getHTMLContents fetches and returns the HTML content from the specified URL.
func getHTMLContents(URL string) (string, error) {
	resp, err := http.Get(URL)
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// getText extracts and returns the concatenated text content of an HTML node.
func getText(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	var text string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		text += getText(c)
	}
	return strings.TrimSpace(text)
}

// getAttribute returns the value of an attribute from an HTML node.
func getAttribute(n *html.Node, attrKey string) string {
	for _, a := range n.Attr {
		if a.Key == attrKey {
			return a.Val
		}
	}
	return ""
}

// countTableColumns returns the amount of columns in a table of an HTML node.
func countTableColumns(n *html.Node) int {
	columnCount := 0
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && (c.Data == "td" || c.Data == "th") {
			columnCount++
		}
	}
	return columnCount
}

// processHeaderCell handles the header cells of the table, parses class name and creates slice for period entries.
func processHeaderCell(cell *html.Node, columnCount int) {
	entry := classEntry{
		Name:          getText(cell),
		PeriodEntries: make([][]string, columnCount),
	}
	allClassEntries = append(allClassEntries, entry)
}

// processDataCell handles the period entries of a class, appends them to the last index of allClassEntries.
func processDataCell(cell *html.Node, periodNum *int) {
	if len(allClassEntries) == 0 {
		return
	}

	lastEntry := &allClassEntries[len(allClassEntries)-1]
	content := getText(cell)

	// Check if PeriodEntries exists for the current periodNum
	if len(lastEntry.PeriodEntries) <= *periodNum {
		// Initialize the missing slice elements if needed
		for i := len(lastEntry.PeriodEntries); i <= *periodNum; i++ {
			lastEntry.PeriodEntries[i] = make([]string, 0)
		}
	}

	lastEntry.PeriodEntries[*periodNum] = append(lastEntry.PeriodEntries[*periodNum], content)
	*periodNum++
}

// processRow processes a single HTML table row (tr element).
func processRow(trNode *html.Node, columnCount int) {
	periodNum := 0
	for cell := trNode.FirstChild; cell != nil; cell = cell.NextSibling {
		if cell.Type == html.ElementNode && cell.Data == "td" {
			colspan := getAttribute(cell, "colspan")
			if colspan == "8" {
				// Handle header rows spanning across all columns.
				processHeaderCell(cell, columnCount)
			} else {
				// Handle standard data cells.
				processDataCell(cell, &periodNum)
			}
		}
	}
}

// getDateTitle parses the title of the table, which includes the date and
// sometimes other information like week letter (according to the parsed html).
func getDateTitle(n *html.Node) string {
	var dateTitle string

	class := getAttribute(n, "class")
	if class == "mon_title" {
		dateTitle = getText(n.FirstChild)
	}
	return dateTitle
}

// parseClassEntries takes the HTML content and extracts class entries from it.
func parseClassEntries(htmlContent string) (string, error) {
	var dateTitle string
	tableCount := 0
	var columnCount int

	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return "", err
	}

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "table" {
			tableCount++
		}
		if n.Type == html.ElementNode && n.Data == "div" {
			dateTitle = getDateTitle(n)
		}
		if tableCount > 1 { // Skipping the first table
			newColumnCount := countTableColumns(n)
			if newColumnCount > columnCount {
				columnCount = newColumnCount
			}
			if n.Type == html.ElementNode && n.Data == "tr" {
				processRow(n, columnCount)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	return dateTitle, nil
}

// ParseHTML is the main function to parse HTML content from the specified URL and process class entries.
func ParseHTML(URL string) ([]classEntry, string) {
	// clear allClassEntries to avoid redundancy when calling the function multiple times
	allClassEntries = nil

	body, err := getHTMLContents(URL)
	if err != nil {
		log.Fatalf("Could not get HTML content: %v", err)
	}
	// log.Printf("Successfully downloaded HTML content from URL %s", URL)

	dateTitle, err := parseClassEntries(body)
	if err != nil {
		log.Fatalf("Could not parse class entries: %v", err)
	}
	log.Printf("Parsed class entries: %d", len(allClassEntries))

	return allClassEntries, dateTitle
}
