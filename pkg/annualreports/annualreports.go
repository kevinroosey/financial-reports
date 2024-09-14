package annualreports

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type FinancialData struct {
	TotalNetSales           string `json:"totalNetSales"`
	TotalCostOfSales        string `json:"totalCostOfSales"`
	TotalOperatingExpenses  string `json:"totalOperatingExpenses"`
	BasicEarningsPerShare   string `json:"basicEarningsPerShare"`
	DilutedEarningsPerShare string `json:"dilutedEarningsPerShare"`
}

type Filing struct {
	Form          string        `json:"form"`
	FilingDate    string        `json:"filingDate"`
	AccessionNo   string        `json:"accessionNo"`
	ReportDate    string        `json:"reportDate"`
	PrimaryDoc    string        `json:"primaryDoc"`
	FinancialData FinancialData `json:"financialData"`
}

func ScrapeFinancialData(cik string, accessionNo string, primaryDoc string) ([]FinancialData, error) {
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

	// Initialize a slice to store financial data for multiple reports
	var financialDataReports []FinancialData
	found := false
	var data FinancialData
	// Loop through each table and search for rows that contain financial data
	doc.Find("table").Each(func(i int, table *goquery.Selection) {
		// Initialize a new FinancialData object for this report

		// Loop through each row (tr) in the table
		table.Find("tr").Each(func(j int, row *goquery.Selection) {
			// Search for each field by matching the label in one of the cells
			row.Find("td").Each(func(k int, cell *goquery.Selection) {
				cellText := strings.ToLower(strings.TrimSpace(cell.Text()))

				// Look for specific fields based on the text in the cell
				switch {
				case strings.Contains(cellText, "total net sales") || strings.Contains(cellText, "net sales") || strings.Contains(cellText, "net revenue") || strings.Contains(cellText, "total revenues") || strings.Contains(cellText, "revenues"):
					value := extractStringFromNextCell(row, k, 2)
					data.TotalNetSales = value
					found = true
				case strings.Contains(cellText, "total cost of sales"):
					value := extractStringFromNextCell(row, k, 1)
					data.TotalCostOfSales = value
					found = true
				case strings.Contains(cellText, "total operating expenses"):
					value := extractStringFromNextCell(row, k, 1)
					data.TotalOperatingExpenses = value
					found = true

				case strings.Contains(cellText, "basic"):
					value := extractStringFromNextCell(row, k, 2)
					data.BasicEarningsPerShare = value
					found = true
				case strings.Contains(cellText, "diluted"):
					value := extractStringFromNextCell(row, k, 2)
					data.DilutedEarningsPerShare = value
					found = true
				}
			})
		})

		// Only add the financial data if we found relevant fields in this table

	})
	if found {
		financialDataReports = append(financialDataReports, data)
	}

	// Return the slice of financial data objects
	return financialDataReports, nil
}

// Helper function to extract a string from the next cell in the row
func extractStringFromNextCell(row *goquery.Selection, index int, skip int) string {
	var value string
	row.Find("td").Each(func(k int, cell *goquery.Selection) {
		if k == index+skip {
			// Get the text from the next cell
			value = strings.TrimSpace(cell.Text())
		}
	})
	return value
}
