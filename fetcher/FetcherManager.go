package fetcher

import (
	"fmt"
)

var fetcherMap map[string]Fetcher

func init() {
	fetcherMap = make(map[string]Fetcher)
	fmt.Println("初始化 fetcherManager")
}

func RegistFetcher(name string, fetcher Fetcher) {
	if _, ok := fetcherMap[name]; ok {
		fmt.Printf("%s fetcher 已经存在\r\n", name)
	}

	fetcherMap[name] = fetcher
	fmt.Printf("注册fetcher:%s\r\n", name)
}

func GetFetcher(name string) (fetcher Fetcher) {
	return fetcherMap[name]
}
