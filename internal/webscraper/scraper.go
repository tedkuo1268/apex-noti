package webscraper

type Scraper interface {
	GetData(url string)
}

func Scrape(s Scraper, url string) {
	s.GetData(url)
}
