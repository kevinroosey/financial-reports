package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type FinancialData struct {
	TotalNetSales           int `json:"totalNetSales"`
	TotalCostOfSales        int `json:"totalCostOfSales"`
	TotalOperatingExpenses  int `json:"totalOperatingExpenses"`
	OperatingExpenses       int `json:"operatingExpenses"`
	BasicEarningsPerShare   int `json:"basicEarningsPerShare"`
	DilutedEarningsPerShare int `json:"dilutedEarningsPerShare"`
}

type Filing struct {
	Form          string        `json:"form"`
	FilingDate    string        `json:"filingDate"`
	AccessionNo   string        `json:"accessionNo"`
	ReportDate    string        `json:"reportDate"`
	PrimaryDoc    string        `json:"primaryDoc"`
	FinancialData FinancialData `json:"financialData"`
}

func fetchFilings(w http.ResponseWriter, r *http.Request) {
	cik := r.URL.Query().Get("cik")
	if cik == "" {
		http.Error(w, "CIK is required", http.StatusBadRequest)
		return
	}

	// Make request to the EDGAR API
	url := fmt.Sprintf("https://data.sec.gov/submissions/CIK%s.json", cik)
	log.Printf("Fetching URL: %s\n", url)

	// Create a new request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("Failed to create request: %v\n", err)
		http.Error(w, "Failed to fetch filings", http.StatusInternalServerError)
		return
	}

	// Set required headers
	req.Header.Set("User-Agent", "YourAppName/1.0 (your-email@example.com)")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to fetch URL: %v\n", err)
		http.Error(w, "Failed to fetch filings", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response body: %v\n", err)
		http.Error(w, "Failed to fetch filings", http.StatusInternalServerError)
		return
	}

	// Decode the JSON response
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		log.Printf("Failed to parse JSON: %v\n", err)
		http.Error(w, "Failed to fetch filings", http.StatusInternalServerError)
		return
	}

	// Extract the recent filings
	recentFilings, ok := data["filings"].(map[string]interface{})["recent"].(map[string]interface{})
	if !ok {
		log.Printf("Failed to extract recent filings")
		http.Error(w, "No recent filings found", http.StatusInternalServerError)
		return
	}
	log.Printf("Recent filings keys: %v", reflect.ValueOf(recentFilings).MapKeys())

	// Extract fields, checking for nil and type safety
	formTypes, ok := recentFilings["form"].([]interface{})
	if !ok {
		log.Printf("Failed to extract form types")
		http.Error(w, "No form types found", http.StatusInternalServerError)
		return
	}

	filingDates, _ := recentFilings["filingDate"].([]interface{})
	accessionNos, _ := recentFilings["accessionNumber"].([]interface{})
	reportDates, _ := recentFilings["reportDate"].([]interface{})
	primaryDocs, _ := recentFilings["primaryDocument"].([]interface{})

	// Filter the filings for 10-Q and 10-K
	filteredFilings := make([]map[string]interface{}, 0)
	for i, formType := range formTypes {
		// Ensure formType is "10-Q" or "10-K"
		if formType == "10-K" {
			// Check if the required fields are within bounds before accessing them
			if i >= len(filingDates) || i >= len(accessionNos) || i >= len(primaryDocs) {
				log.Printf("Skipping entry due to missing fields at index %d", i)
				continue
			}

			// Safely assign data if the field exists and has enough entries
			filing := map[string]interface{}{
				"form":        formType,
				"filingDate":  filingDates[i],
				"accessionNo": accessionNos[i],
				"reportDate":  reportDates[i],
				"primaryDoc":  primaryDocs[i],
			}

			// Scrape financial data for the filing
			financialData, err := scrapeFinancialData(cik, accessionNos[i].(string), primaryDocs[i].(string))
			if err != nil {
				log.Printf("Error scraping financial data: %v\n", err)
				http.Error(w, "Error scraping financial data", http.StatusInternalServerError)
				return
			}
			filing["financialData"] = financialData

			filteredFilings = append(filteredFilings, filing)
		}
	}

	// Send the filtered filings as the response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(filteredFilings); err != nil {
		log.Printf("Failed to encode filtered filings: %v\n", err)
		http.Error(w, "Failed to encode filings", http.StatusInternalServerError)
	}
}

func scrapeFinancialData(cik string, accessionNo string, primaryDoc string) ([]string, error) {
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
	financialData := []string{}

	doc.Find("table").Each(func(i int, table *goquery.Selection) {
		table.Find("tr").Each(func(j int, row *goquery.Selection) {
			// Collect the table row text
			rowText := row.Text()

			// Now, check for your specific text
			if strings.Contains(strings.ToLower(rowText), "total net sales") { // Use case-insensitive matching
				financialData = append(financialData, rowText)

			}
		})
	})

	log.Printf("Financial data: %v\n", financialData)

	return financialData, nil
}

func main() {
	http.HandleFunc("/filings", fetchFilings)
	log.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
