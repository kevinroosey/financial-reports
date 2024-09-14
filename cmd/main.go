package main

import (
	"log"
	"net/http"
	// import the package "filings" from the "filings" directory
	"github.com/kevinroosey/financial-reports/pkg/filings"
)


func main() {
	http.HandleFunc("/filings", filings.fetchFilings)
	log.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
