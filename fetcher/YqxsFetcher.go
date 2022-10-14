package fetcher

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/wy/crawler/config"
	"github.com/wy/crawler/utils"

	"github.com/PuerkitoBio/goquery"
)

type YqxsFetcher struct {
}

func init() {
	var fetcher Fetcher = YqxsFetcher{}
	RegistFetcher(fetcher.GetName(), fetcher)
}

func (fetcher YqxsFetcher) GetName() string {
	return "yqxs"
}

func (fetcher YqxsFetcher) FetchAllBook(indexUrl string, otherUrls []string) {
	bookIndexUrls := FetchAllBookIndexUrl(indexUrl, otherUrls)

	baseDir := config.GetValue("base.dir")
	bookIndexDir := baseDir + "/url/"
	bookIndexPath := bookIndexDir + "index.txt"
	fmt.Println(bookIndexPath)

	if !utils.IsFileExist(bookIndexPath) {
		os.MkdirAll(bookIndexDir, 777)
		// 保存当前小说的链接
		utils.WriteToFile(bookIndexPath, bookIndexUrls, false)
	} else {
		oldBookIndexUrls := utils.ReadLines(bookIndexPath)

		bookIndexUrls = utils.Distinct(bookIndexUrls, oldBookIndexUrls)
		utils.WriteToFile(bookIndexPath, bookIndexUrls, false)
	}

	taskQueue := make(chan string, len(bookIndexUrls))
	for _, bookIndexUrl := range bookIndexUrls {
		taskQueue <- bookIndexUrl
	}

	workerNum := runtime.NumCPU()
	wg := sync.WaitGroup{}
	wg.Add(workerNum)

	for i := 0; i < workerNum; i++ {
		go func() {
			for {
				select {
				case bookIndexUrl, ok := <-taskQueue:
					if !ok {
						wg.Done()
						break
					}
					fmt.Printf("爬虫当前爬取链接: %s\r\n", bookIndexUrl)
					FetchNewOrUpdateBook(bookIndexUrl)
				}
			}
		}()
	}

	wg.Wait()

	fmt.Println("爬取结束")
}

func FetchAllBookIndexUrl(indexUrl string, otherUrls []string) (bookIndexUrls []string) {
	pageIndexFilter := make(map[string]int8)
	bookIndexFilter := make(map[string]int8)
	pageIndexFilter[indexUrl] = 1

	pageUrls := []string{}
	pageUrls = append(pageUrls, indexUrl)
	for i := 0; i < len(otherUrls); i++ {
		pageUrls = append(pageUrls, otherUrls[i])
	}

	url, _ := url.Parse(indexUrl)
	host := url.Scheme + "://" + url.Host

	// 首页的小说链接
	for i := 0; i < len(pageUrls); i++ {
		pageUrl := pageUrls[i]

		// 获取当页的
		bookIndexFilter2, pageIndexFilter2 := FetchIndexPageUrl(pageUrl)
		for url, _ := range bookIndexFilter2 {
			bookIndexFilter[url] = 1
		}

		for url, _ := range pageIndexFilter2 {
			if _, isKeyExist := pageIndexFilter[url]; isKeyExist {
				continue
			}
			//fmt.Printf("%s ==> %s\r\n", pageUrl, url)

			pageUrl2 := host + url

			pageIndexFilter[url] = 1
			pageUrls = append(pageUrls, pageUrl2)
		}
	}

	bookIndexUrls = []string{}
	for url, _ := range bookIndexFilter {
		pageUrl3 := host + url
		bookIndexUrls = append(bookIndexUrls, pageUrl3)
		// fmt.Println(pageUrl3)
	}

	return
}

func FetchIndexPageUrl(pageIndexUrl string) (bookIndexFilter, pageIndexFilter map[string]int8) {
	content := Fetch(pageIndexUrl)
	defer content.Close()

	b, _ := ioutil.ReadAll(content)
	html := string(b)

	bookIndexFilter = make(map[string]int8)
	pageIndexFilter = make(map[string]int8)

	pageIndexUrlRegex, _ := regexp.Compile("/yq/\\d+/\\d+\\.html")
	pageIndexUrls := pageIndexUrlRegex.FindAllString(html, -1)
	for _, pageIndex := range pageIndexUrls {
		// fmt.Println("page: " + pageIndex)
		pageIndexFilter[pageIndex] = 1
	}

	bookIndexUrlRegex, _ := regexp.Compile("/html/\\d+/\\d+/index\\.html")
	bookIndexUrls := bookIndexUrlRegex.FindAllString(html, -1)
	for _, index := range bookIndexUrls {
		bookIndexFilter[index] = 1

	}

	return
}

func FetchNewOrUpdateBook(bookUrl string) {
	// 爬取目录页，获取章节链接和章节名
	hrefList, titleMap, bookName := FetchCatalogue(bookUrl)

	baseDir := config.GetValue("base.dir")
	bookUrlDir := baseDir + "/url/"
	bookUrlPath := bookUrlDir + bookName + ".txt"
	// 全量更新
	if !utils.IsFileExist(bookUrlPath) {
		os.MkdirAll(bookUrlDir, 777)

		// 抓取小说全文
		FetchNewBook(bookUrl)
		// 记录章节链接
		utils.WriteToFile(bookUrlPath, hrefList, false)

		return
	}

	// 增量更新
	appendHrefList := []string{}
	hrefFilter := make(map[string]int)

	oldHrefList := utils.ReadLines(bookUrlPath)
	for _, oldHref := range oldHrefList {
		hrefFilter[oldHref] = 1
	}
	for _, href := range hrefList {
		if _, isKeyExist := hrefFilter[href]; !isKeyExist {
			appendHrefList = append(appendHrefList, href)
		}
	}
	if len(appendHrefList) == 0 {
		fmt.Printf("小说《%s》暂时未更新", bookName)
		return
	}

	baseUrl, _ := url.Parse(bookUrl)
	host := baseUrl.Scheme + "://" + baseUrl.Host
	bookPath := baseDir + "/" + bookName + ".txt"

	sleepTime := 0
	if config.ContainsKey("sleep_time") {
		sleepTime, _ = strconv.Atoi(config.GetValue("sleep_time"))
	}

	chapterLogOpen := config.GetIntValue("log.switch.chapter")

	lines := []string{}
	for i, href := range appendHrefList {
		chapterUrl := host + href
		if chapterLogOpen == 1 {
			fmt.Printf("[append]%d  %s   %s\r\n", i, chapterUrl, titleMap[href])
		}

		content := FetchBookChapterContent(chapterUrl)
		// 写入章节名称
		lines = append(lines, titleMap[href])
		lines = append(lines, "\r\n")
		// 写入章节内容
		lines = append(lines, content)
		lines = append(lines, "\r\n\r\n")
		//休眠一段时间
		if sleepTime > 0 {
			time.Sleep(time.Duration(sleepTime) * time.Millisecond)
		}
	}

	// 记录新加的内容到文件
	utils.WriteToFile(bookPath, lines, true)
	// 记录新加的章节链接
	utils.WriteToFile(bookUrlPath, appendHrefList, true)
}

func FetchNewBook(bookUrl string) {
	// 爬取目录页，获取章节链接和章节名
	hrefList, titleMap, bookName2 := FetchCatalogue(bookUrl)

	baseUrl, _ := url.Parse(bookUrl)
	host := baseUrl.Scheme + "://" + baseUrl.Host

	// 存储路径
	baseDir := config.GetValue("base.dir")
	bookPath := baseDir + "/" + bookName2 + ".txt"
	fmt.Println("创建文件：" + bookPath)

	file, err := os.OpenFile(bookPath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		fmt.Printf("打开文件错误=%v \n", err)
		return
	}
	defer file.Close()

	sleepTime := 0
	if config.ContainsKey("sleep_time") {
		sleepTime, _ = strconv.Atoi(config.GetValue("sleep_time"))
	}

	chapterLogOpen := config.GetIntValue("log.switch.chapter")

	writer := bufio.NewWriter(file)
	for i, href := range hrefList {
		chapterUrl := host + href
		if chapterLogOpen == 1 {
			fmt.Printf("[fetch-]%d  %s   %s\r\n", i, chapterUrl, titleMap[href])
		}

		content := FetchBookChapterContent(chapterUrl)
		// fmt.Println(content)
		// 写入章节名称
		writer.WriteString(titleMap[href])
		writer.WriteString("\r\n")
		// 写入章节内容
		writer.WriteString(content)
		writer.WriteString("\r\n\r\n")

		// 休眠一段时间
		if sleepTime > 0 {
			time.Sleep(time.Duration(sleepTime) * time.Millisecond)
		}
	}

	// 刷新缓冲流
	writer.Flush()
}

func FetchBookChapterContent(bookChapterUrl string) (content string) {
	// http请求
	contentReader := Fetch(bookChapterUrl)
	defer contentReader.Close()
	// 解析html网页
	doc, err := goquery.NewDocumentFromReader(contentReader)
	if err != nil {
		fmt.Println("parser url content error")
		log.Fatal(err)
		return
	}

	html := doc.Find("html")
	// head := html.Find("head")
	body := html.Find("body")
	txt, _ := body.Find("div.content").Find("div.showtxt").Html()

	txt2 := utils.ConvertToString(txt, "GBK", "UTF-8")

	txt2 = strings.Replace(txt2, "<script>app2();</script><br/><script>read2();</script>", "", 1)
	txt2 = strings.ReplaceAll(txt2, "<br/><br/>", "\r\n")
	end := strings.Index(txt2, "<script>app2();</script>")
	if end > 0 {
		txt2 = txt2[0:end]
	}
	content = strings.ReplaceAll(txt2, "<br/>", "\r\n")

	return
}

func FetchCatalogue(bookUrl string) (hrefList []string, titleMap map[string]string, bookName2 string) {
	// http请求
	contentReader := Fetch(bookUrl)
	defer contentReader.Close()

	// 解析html网页
	doc, err := goquery.NewDocumentFromReader(contentReader)
	if err != nil {
		fmt.Println("parser url content error")
		log.Fatal(err)
		return
	}

	html := doc.Find("html")
	head := html.Find("head")
	bookCharset, _ := head.Find("meta[http-equiv='Content-Type']").Attr("content")
	charset := strings.Split(bookCharset, "=")[1]

	bookName, _ := head.Find("meta[property='og:novel:book_name']").Attr("content")
	bookName2 = utils.ConvertToString(bookName, charset, "UTF-8")

	body := html.Find("body")
	listMain := body.Find("div.listmain")
	// chapterList := listMain.Find("dl").Find("dd")
	chapterList := listMain.Find("dl").Children()

	titleMap = make(map[string]string)
	hrefList = []string{}

	chapterList.Each(func(i int, s *goquery.Selection) {
		a := s.Find("a")

		title := a.Text()
		title = utils.ConvertToString(title, charset, "UTF-8")
		if len(title) == 0 {
			for k := range titleMap {
				delete(titleMap, k)
			}
			hrefList = []string{}
			return
		}

		href, _ := a.Attr("href")
		titleMap[href] = title
		hrefList = append(hrefList, href)
	})

	return
}

func Fetch(url string) (content io.ReadCloser) {
	retryTimes := config.GetIntValue("http.retryTimes")
	failSleepTime := config.GetIntValue("http.failSleepTime")

	for retryTimes > 0 {
		content = doFetch(url)
		if content != nil {
			return
		} else {
			time.Sleep(time.Duration(failSleepTime) * time.Millisecond)
			retryTimes--
		}
	}

	return
}

func doFetch(url string) (content io.ReadCloser) {
	// defer func() {
	// 	if i := recover(); i != nil {
	// 		fmt.Printf("recover from fetch, %v\r\n", i)
	// 	}
	// }()

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("request url %s error\r\n", url)
		return
	}

	content = resp.Body
	return

}
