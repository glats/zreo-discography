package main

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gocolly/colly"
)

func main() {
	var wg sync.WaitGroup
	domains := colly.AllowedDomains("zreomirror.com", "www.zreomirror.com")
	agent := colly.UserAgent("Mozilla/5.0 (Windows NT 6.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2228.0 Safari/537.36")
	mainCollection := colly.NewCollector(
		domains,
		agent,
	)
	childrenCollection := colly.NewCollector(
		domains,
		agent,
	)
	mainCollection.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if 0 != strings.Compare("download", strings.ToLower(e.Text)) {
			return
		}
		fmt.Printf("Main: link found: %q -> %s\n", e.Text, link)
		childrenCollection.Visit(e.Request.AbsoluteURL(link))
	})
	childrenCollection.OnHTML("body", func(e *colly.HTMLElement) {
		title := e.DOM.Find("div[class=album-title]").ChildrenFiltered("h1").Text()
		link, _ := e.DOM.Find("div[class=col-md-9]").ChildrenFiltered("div[class=btn-group]").Children().Eq(1).Attr("href")
		fmt.Printf("title: %s , link: %s\n", title, link)
		u, _ := url.Parse(link)
		zipfile := strings.Split(u.Path, "/")[len(strings.Split(u.Path, "/"))-1]
		wg.Add(1)
		go func(zipfile string, link string, title string, wg *sync.WaitGroup) {
			defer wg.Done()
			err := Download(filepath.Join("downloads", zipfile), link)
			fmt.Printf("error: %s\n", err)
			_, ziperr := Unzip(filepath.Join("downloads", zipfile), filepath.Join("disks", title))
			fmt.Printf("ziperr: %s\n", ziperr)
		}(zipfile, link, title, &wg)

	})
	mainCollection.Visit("http://www.zreomirror.com/")
	wg.Wait()
}
