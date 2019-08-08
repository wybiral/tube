package app

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/wybiral/feeds"
)

// buildFeed creates RSS feed attribute for App based on Library contents.
func buildFeed(a *App) {
	cfg := a.Config.Feed
	now := time.Now()
	f := &feeds.Feed{
		Title:       cfg.Title,
		Link:        &feeds.Link{Href: cfg.Link},
		Description: cfg.Description,
		Author: &feeds.Author{
			Name:  cfg.Author.Name,
			Email: cfg.Author.Email,
		},
		Created:   now,
		Copyright: cfg.Copyright,
	}
	var externalURL string
	if len(cfg.ExternalURL) > 0 {
		externalURL = cfg.ExternalURL
	} else if a.Tor != nil {
		onion, err := a.Tor.OnionKey.Onion()
		if err != nil {
			return
		}
		externalURL = fmt.Sprintf("http://%s.onion", onion.ServiceID)
	} else {
		hostname, err := os.Hostname()
		if err != nil {
			host := a.Config.Server.Host
			port := a.Config.Server.Port
			externalURL = fmt.Sprintf("http://%s:%d", host, port)
		} else {
			externalURL = fmt.Sprintf("http://%s", hostname)
		}
	}
	for _, v := range a.Library.Playlist() {
		u, err := url.Parse(externalURL)
		if err != nil {
			return
		}
		u.Path = path.Join(u.Path, "v", v.ID)
		id := u.String()
		f.Items = append(f.Items, &feeds.Item{
			Id:          id,
			Title:       v.Title,
			Link:        &feeds.Link{Href: id},
			Description: v.Description,
			Enclosure: &feeds.Enclosure{
				Url:    id + ".mp4",
				Length: strconv.FormatInt(v.Size, 10),
				Type:   "video/mp4",
			},
			Author: &feeds.Author{
				Name:  cfg.Author.Name,
				Email: cfg.Author.Email,
			},
			Created: v.Timestamp,
		})
	}
	feed, err := f.ToRss()
	if err != nil {
		return
	}
	a.Feed = []byte(feed)
}
