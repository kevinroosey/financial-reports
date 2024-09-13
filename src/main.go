package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

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

	// Log the raw JSON response
	log.Printf("Raw JSON response: %s\n", string(body))

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

	// Extract fields, checking for nil and type safety
	formTypes, ok := recentFilings["form"].([]interface{})
	if !ok {
		log.Printf("Failed to extract form types")
		http.Error(w, "No form types found", http.StatusInternalServerError)
		return
	}

	filingDates, _ := recentFilings["filingDate"].([]interface{})
	accessionNos, _ := recentFilings["accessionNo"].([]interface{})
	reportDates, _ := recentFilings["reportDate"].([]interface{})
	primaryDocs, _ := recentFilings["primaryDoc"].([]interface{})

	// Filter the filings for 10-Q and 10-K
	filteredFilings := make([]map[string]interface{}, 0)
	for i, formType := range formTypes {
		if formType == "10-Q" || formType == "10-K" {
			filing := map[string]interface{}{
				"form": formType,
			}

			// Safely assign data if the field exists and has enough entries
			if i < len(filingDates) {
				filing["filingDate"] = filingDates[i]
			}
			if i < len(accessionNos) {
				filing["accessionNo"] = accessionNos[i]
			}
			if i < len(reportDates) {
				filing["reportDate"] = reportDates[i]
			}
			if i < len(primaryDocs) {
				filing["primaryDoc"] = primaryDocs[i]
			}

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

func main() {
	http.HandleFunc("/filings", fetchFilings)
	log.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
