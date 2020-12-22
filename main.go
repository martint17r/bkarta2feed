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
		Title: "Laufende Fusionskontrollverfahren",
		Link:  &feeds.Link{Href: rawUri},
		Description: `Hier finden Sie eine Liste der laufenden Fusionskontrollverfahren.

Das Prüfverfahren beginnt nach dem Eingang der vollständigen Anmeldeunterlagen beim Bundeskartellamt. Die Behörde hat zunächst einen Monat Zeit, um den Zusammenschluss zu prüfen (sog. "erste Phase"). Erweist sich das Fusionsvorhaben als unproblematisch, gibt die Beschlussabteilung den Zusammenschluss vor Ablauf der Monatsfrist formlos frei.

Hält die Beschlussabteilung dagegen eine weitere Prüfung für erforderlich, wird ein förmliches Hauptprüfverfahren eingeleitet (sog. "zweite Phase") und die Frist für die Prüfung des Vorhabens verlängert. Das Bundeskartellamt muss bei Durchführung eines Hauptprüfverfahrens innerhalb von vier Monaten ab Eingang der vollständigen Anmeldung entscheiden. Die Liste der Hauptprüfverfahren finden Sie hier.

Bei den mit * gekennzeichneten Verfahren wurde der Zusammenschluss bereits vollzogen. Hier handelt es sich um eine nachträgliche Prüfung im fristungebundenen Verfahren nach § 41 Abs. 3 GWB.

Die Liste erhebt keinen Anspruch auf Vollständigkeit.  `,
		Author:  &feeds.Author{Name: "Martin Treusch von Buttlar", Email: "fusion2feed@m.t17r.de"},
		Created: now,
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
