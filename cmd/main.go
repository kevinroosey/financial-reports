package main

import (
	"log"
	"net/http"
	// Update this import path
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

func main() {
	http.HandleFunc("/filings", filings.fetchFilings)
	log.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
