package fetcher

type Fetcher interface {
	GetName() string
	FetchAllBook(indexUrl string, otherUrls []string)
}
