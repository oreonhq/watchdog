package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/mmcdole/gofeed"
)

var tmpl = template.Must(template.New("index.html").Funcs(template.FuncMap{
	"safeHTML": func(s string) template.HTML { return template.HTML(s) },
}).ParseFiles("index.html"))

type Item struct {
	Title       string
	Link        string
	Description template.HTML
	Published   string
}

type FeedData struct {
	Title string
	Items []Item
}

var currentFeed *gofeed.Feed

func main() {
	http.Handle("/style.css", http.FileServer(http.Dir(".")))

	go func() {
		for {
			fetchFeedAndStore()
			time.Sleep(10 * time.Minute)
		}
	}()

	http.HandleFunc("/", handler)
	fmt.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func fetchFeedAndStore() {
	feedURL := "https://cvefeed.io/rssfeed/latest.xml"
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(feedURL)
	if err != nil {
		log.Printf("Error fetching feed: %v", err)
		return
	}
	currentFeed = feed
	log.Printf("Successfully updated feed: %s", feed.Title)
}

func handler(w http.ResponseWriter, r *http.Request) {
	if currentFeed == nil {
		http.Error(w, "Feed not available yet", http.StatusServiceUnavailable)
		return
	}

	items := make([]Item, 0, len(currentFeed.Items))
	for _, it := range currentFeed.Items {
		desc := ""
		if it.Description != "" {
			desc = it.Description
		} else if it.Content != "" {
			desc = it.Content
		}

		published := ""
		if it.Published != "" {
			published = it.Published
		} else if it.PublishedParsed != nil {
			published = it.PublishedParsed.Format(time.RFC1123)
		}

		items = append(items, Item{
			Title:       it.Title,
			Link:        it.Link,
			Description: template.HTML(desc),
			Published:   published,
		})
	}

	data := FeedData{
		Title: currentFeed.Title,
		Items: items,
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
