from sec_cik_mapper import StockMapper
from pathlib import Path
'''
    Unfortunately, SEC edgar data requires CIK instead of ticker symbols so we need to map the ticker symbols to CIKs so API route can use ticker symbols instead of ciks
'''
def main():
    stock_mapper = StockMapper()
    stock_mapper.ticker_to_cik
    
    csv_path = Path("../../src/data/ticker-to-cik.csv")
    stock_mapper.save_metadata_to_csv(csv_path)
    
if __name__ == "__main__":
    main()
