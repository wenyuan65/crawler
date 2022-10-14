package main

import (
	"strings"

	"github.com/wy/crawler/config"
	"github.com/wy/crawler/fetcher"
)

func main() {
	templates := config.GetValue("templates")
	templateList := strings.Split(templates, ",")

	for _, templateName := range templateList {
		// name := config.GetValue("handler." + handler + ".name")
		url := config.GetValue("template." + templateName + ".url")
		spcials := config.GetValue("template." + templateName + ".spcials")
		spcialUrls := strings.Split(spcials, ",")

		f := fetcher.GetFetcher(templateName)
		// fmt.Printf("[%s] fetch: %s, spcials:%d\r\n", f.GetName(), url, len(spcialUrls))
		f.FetchAllBook(url, spcialUrls)
	}

}
