# financial-reports

Tiny weekend project that allows querying annual reports from thousands of publicly traded companies without having to scrape data from the SEC. /


Some improvements that need to be made:

- [ ] Improve taxnonomy between companies, i.e, one company may call revenue "Net revenues" whereas others might call it "total revenues"
- [ ] change financial data from being strings in the milions to full integer numbers

## Example usage

```
curl "https://financial-reports-production.up.railway.app/filings?ticker=AAPL"
```

### Output
```
[
  {
    "accessionNo": "0000320193-23-000106",
    "filingDate": "2023-11-03",
    "financialData": [
      {
        "totalNetSales": "383,285",
        "totalCostOfSales": "214,137",
        "totalOperatingExpenses": "54,847",
        "basicEarningsPerShare": "6.16",
        "dilutedEarningsPerShare": "6.13"
      }
    ],
    "form": "10-K",
    "primaryDoc": "aapl-20230930.htm",
    "reportDate": "2023-09-30"
  },
  {
    "accessionNo": "0000320193-22-000108",
    "filingDate": "2022-10-28",
    "financialData": [
      {
        "totalNetSales": "394,328",
        "totalCostOfSales": "223,546",
        "totalOperatingExpenses": "51,345",
        "basicEarningsPerShare": "6.15",
        "dilutedEarningsPerShare": "6.11"
      }
    ],
    "form": "10-K",
    "primaryDoc": "aapl-20220924.htm",
    "reportDate": "2022-09-24"
  }
]
```

