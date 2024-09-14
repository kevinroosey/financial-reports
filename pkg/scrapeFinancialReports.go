package scrapeFinancialData

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func scrapeFinancialData(cik string, accessionNo string, primaryDoc string) ([]FinancialData, error) {
	// Construct the URL to the filing document
	accessionNoNoDashes := strings.ReplaceAll(accessionNo, "-", "")
	url := fmt.Sprintf("https://www.sec.gov/Archives/edgar/data/%s/%s/%s", cik, accessionNoNoDashes, primaryDoc)

	// Create the request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("Failed to create request: %v\n", err)
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Set a custom User-Agent header
	req.Header.Set("User-Agent", "KevinWebScraping/1.0 (kevinroosey@gmail.com)")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to fetch filing document: %v\n", err)
		return nil, fmt.Errorf("failed to fetch filing document: %v", err)
	}
	defer resp.Body.Close()

	// Read the body of the document
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read document body: %v\n", err)
		return nil, fmt.Errorf("failed to read document body: %v", err)
	}

	// Parse the document with goquery
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		log.Printf("Failed to parse document: %v\n", err)
		return nil, fmt.Errorf("failed to parse document: %v", err)
	}

	// Example: Scrape financial data by looking for table rows with financial data
	financialData := []FinancialData{}

	doc.Find("table").Each(func(i int, table *goquery.Selection) {
		table.Find("tr").Each(func(j int, row *goquery.Selection) {
			// Fill in all Filing fields

		})
	})

	return financialData, nil
}
