package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/feeds"
)

func main() {
	for _, rawUri := range os.Args[1:] {
		b, err := readerFromURI(rawUri)
		if err != nil {
			log.Fatal(err)
		}
		defer b.Close()
		f, err := extractFeed(rawUri, b)
		if err != nil {
			log.Fatal(err)
		}
		payload, err := f.ToAtom()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Print(payload)
	}
}

func readerFromURI(rawUri string) (io.ReadCloser, error) {
	uri, err := url.ParseRequestURI(rawUri)
	if err != nil {
		return nil, err
	}

	switch uri.Scheme {
	case "http":
		fallthrough
	case "https":
		r, err := http.Get(uri.String())
		if r.StatusCode >= 299 {
			return nil, fmt.Errorf("can't handle status %q (%q) for %s", r.StatusCode, r.Status, uri)
		}
		return r.Body, err
	case "file":
		var fname string
		if uri.Host != "" && uri.Host != "localhost" {
			fname = uri.Host
		} else {
			fname = uri.Path
		}
		return os.Open(fname)
	}
	return nil, fmt.Errorf("can not handle scheme: %q", uri.Scheme)
}

func extractFeed(rawUri string, body io.Reader) (*feeds.Feed, error) {
	now := time.Now()
	feed := &feeds.Feed{
		Title:       "jmoiron.net blog",
		Link:        &feeds.Link{Href: rawUri},
		Description: "discussion about tech, footie, photos",
		Author:      &feeds.Author{Name: "Martin Treusch von Buttlar", Email: "fusion2feed@m.t17r.de"},
		Created:     now,
	}
	feed.Items = make([]*feeds.Item, 0, 20)

	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return nil, err
	}
	doc.Find("table.csvTable tr").Each(func(i int, row *goquery.Selection) {
		if err != nil {
			return
		}
		/*
			if i < 50 || len(feed.Items) > 3 {
				return
			}
		*/
		data := make([]string, 0, 6)
		row.Find("td").Each(func(_ int, cell *goquery.Selection) {
			raw := cell.Text()
			raw = strings.ReplaceAll(raw, "\u00ad", "")
			data = append(data, raw)
		})
		if len(data) == 0 {
			return
		}
		var created time.Time
		created, err = time.ParseInLocation("02.01.2006", data[0], time.Local)
		if err != nil {
			return
		}
		//fmt.Printf("%#v\n", data)
		item := &feeds.Item{
			Title:       data[2],
			Link:        &feeds.Link{Href: fmt.Sprintf("%s#%s", rawUri, data[1])},
			Id:          data[1],
			Description: fmt.Sprintf("%s<br>%s<br>%s", data[2], data[3], data[4]),
			Created:     created,
		}

		feed.Items = append(feed.Items, item)
	})
	return feed, err

}
