// Package app manages main application server.
package app

import (
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/feeds"
	"github.com/gorilla/mux"
	"github.com/wybiral/tube/pkg/media"
)

// App represents main application.
type App struct {
	Config    *Config
	Library   *media.Library
	Watcher   *fsnotify.Watcher
	Templates *template.Template
	Listener  net.Listener
	Router    *mux.Router
}

// NewApp returns a new instance of App from Config.
func NewApp(cfg *Config) (*App, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	a := &App{
		Config: cfg,
	}
	// Setup Library
	a.Library = media.NewLibrary()
	// Setup Watcher
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	a.Watcher = w
	// Setup Listener
	ln, err := newListener(cfg.Server)
	if err != nil {
		return nil, err
	}
	a.Listener = ln
	// Setup Templates
	a.Templates = template.Must(template.ParseGlob("templates/*"))
	// Setup Router
	r := mux.NewRouter().StrictSlash(true)
	r.HandleFunc("/", a.indexHandler).Methods("GET")
	r.HandleFunc("/v/{id}.mp4", a.videoHandler).Methods("GET")
	r.HandleFunc("/t/{id}", a.thumbHandler).Methods("GET")
	r.HandleFunc("/v/{id}", a.pageHandler).Methods("GET")
	r.HandleFunc("/feed.xml", a.rssHandler).Methods("GET")
	// Static file handler
	fsHandler := http.StripPrefix(
		"/static/",
		http.FileServer(http.Dir("./static/")),
	)
	r.PathPrefix("/static/").Handler(fsHandler).Methods("GET")
	a.Router = r
	return a, nil
}

// Run imports the library and starts server.
func (a *App) Run() error {
	path := a.Config.LibraryPath
	err := a.Library.Import(path)
	if err != nil {
		return err
	}
	a.Watcher.Add(path)
	go a.watch()
	return http.Serve(a.Listener, a.Router)
}

// Watch the library path and update Library with changes.
func (a *App) watch() {
	for {
		e, ok := <-a.Watcher.Events
		if !ok {
			return
		}
		if e.Op&fsnotify.Create > 0 {
			// add new files to library
			a.Library.Add(e.Name)
		} else if e.Op&(fsnotify.Write|fsnotify.Chmod) > 0 {
			// writes and chmods should remove old file before adding again
			a.Library.Remove(e.Name)
			a.Library.Add(e.Name)
		} else if e.Op&(fsnotify.Remove|fsnotify.Rename) > 0 {
			// remove and rename just remove file
			// fsnotify will signal a Create event with the new file name
			a.Library.Remove(e.Name)
		}
	}
}

// HTTP handler for /
func (a *App) indexHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("/")
	pl := a.Library.Playlist()
	if len(pl) > 0 {
		http.Redirect(w, r, "/v/"+pl[0].ID, 302)
	} else {
		a.Templates.ExecuteTemplate(w, "index.html", &struct {
			Playing  *media.Video
			Playlist media.Playlist
		}{
			Playing:  &media.Video{ID: ""},
			Playlist: a.Library.Playlist(),
		})
	}
}

// HTTP handler for /v/id
func (a *App) pageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	log.Printf("/v/%s", id)
	playing, ok := a.Library.Videos[id]
	if !ok {
		a.Templates.ExecuteTemplate(w, "index.html", &struct {
			Playing  *media.Video
			Playlist media.Playlist
		}{
			Playing:  &media.Video{ID: ""},
			Playlist: a.Library.Playlist(),
		})
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	a.Templates.ExecuteTemplate(w, "index.html", &struct {
		Playing  *media.Video
		Playlist media.Playlist
	}{
		Playing:  playing,
		Playlist: a.Library.Playlist(),
	})
}

// HTTP handler for /v/id.mp4
func (a *App) videoHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	log.Printf("/v/%s", id)
	m, ok := a.Library.Videos[id]
	if !ok {
		return
	}
	title := m.Title
	disposition := "attachment; filename=\"" + title + ".mp4\""
	w.Header().Set("Content-Disposition", disposition)
	w.Header().Set("Content-Type", "video/mp4")
	path := a.Config.LibraryPath + "/" + id + ".mp4"
	http.ServeFile(w, r, path)
}

// HTTP handler for /t/id
func (a *App) thumbHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	log.Printf("/t/%s", id)
	m, ok := a.Library.Videos[id]
	if !ok {
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=7776000")
	if m.ThumbType == "" {
		w.Header().Set("Content-Type", "image/jpeg")
		http.ServeFile(w, r, "static/defaulticon.jpg")
	} else {
		w.Header().Set("Content-Type", m.ThumbType)
		w.Write(m.Thumb)
	}
}

// HTTP handler for /feed.xml
func (a *App) rssHandler(w http.ResponseWriter, r *http.Request) {
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
	} else {
		host := a.Config.Server.Host
		port := a.Config.Server.Port
		externalURL = fmt.Sprintf("http://%s:%d", host, port)
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
	w.Header().Set("Cache-Control", "public, max-age=7776000")
	w.Header().Set("Content-Type", "text/xml")
	f.WriteRss(w)
}
