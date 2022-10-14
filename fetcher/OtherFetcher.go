package fetcher

import "fmt"

type OtherFetcher struct {
}

func init() {
	var fetcher Fetcher = OtherFetcher{}
	RegistFetcher(fetcher.GetName(), fetcher)
}

func (fetcher OtherFetcher) GetName() string {
	return "other"
}

func (fetcher OtherFetcher) FetchAllBook(indexUrl string, otherUrls []string) {
	fmt.Println("do nothing")
}
