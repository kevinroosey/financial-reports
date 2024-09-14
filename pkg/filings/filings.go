package filings

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"

	"github.com/kevinroosey/financial-reports/pkg/annualreports"
)

func FetchFilings(w http.ResponseWriter, r *http.Request) {
	ticker := r.URL.Query().Get("ticker")
	if ticker == "" {
		http.Error(w, "Ticker is required", http.StatusBadRequest)
		return
	}

	tickerToCIK, err := LoadTickerToCIK("pkg/data/ticker-to-cik.csv")
	if err != nil {
		log.Fatalf("Error loading CSV file: %v\n", err)
	}

	// Example: Find the CIK for a given ticker
	cik, err := GetCIKByTicker(ticker, tickerToCIK)
	if err != nil {
		log.Fatalf("Error finding CIK: %v\n", err)
	}

	// Make request to the EDGAR API
	url := fmt.Sprintf("https://data.sec.gov/submissions/CIK%s.json", cik)

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

	// Filter the filings for 10-K
	filteredFilings := make([]map[string]interface{}, 0)
	for i, formType := range formTypes {
		// Ensure formType is "10-K"
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
			financialData, err := annualreports.ScrapeFinancialData(cik, accessionNos[i].(string), primaryDocs[i].(string))
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

func GetCIKByTicker(ticker string, tickerToCIK map[string]string) (string, error) {
	// Lookup the ticker in the map
	if cik, exists := tickerToCIK[ticker]; exists {
		return cik, nil
	}
	return "", fmt.Errorf("CIK not found for ticker: %s", ticker)
}

func LoadTickerToCIK(csvFile string) (map[string]string, error) {
	// Open the CSV file
	file, err := os.Open(csvFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %v", err)
	}
	defer file.Close()

	// Create a new CSV reader
	reader := csv.NewReader(file)

	// Read all the records
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV file: %v", err)
	}

	// Create a map to store ticker-to-CIK mappings
	tickerToCIK := make(map[string]string)

	// Loop through the records and populate the map (skipping the header)
	for i, record := range records {
		if i == 0 {
			continue // Skip header
		}
		// Ensure the record has at least two fields (CIK and Ticker)
		if len(record) >= 2 {
			cik := record[0]
			ticker := record[1]
			tickerToCIK[ticker] = cik
		}
	}

	return tickerToCIK, nil
}
