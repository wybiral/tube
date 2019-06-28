package app

import (
	"html/template"
	"log"
	"net"
	"net/http"

	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/mux"
	"github.com/wybiral/tube/pkg/media"
)

type App struct {
	Config    *Config
	Library   *media.Library
	Watcher   *fsnotify.Watcher
	Templates *template.Template
	Listener  net.Listener
	Router    *mux.Router
}

func NewApp(cfg *Config) (*App, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	a := &App{
		Config: cfg,
	}
	a.Library = media.NewLibrary()
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	a.Watcher = w
	ln, err := newListener(cfg.Server)
	if err != nil {
		return nil, err
	}
	a.Listener = ln
	a.Templates = template.Must(template.ParseGlob("templates/*"))
	r := mux.NewRouter().StrictSlash(true)
	r.HandleFunc("/", a.indexHandler).Methods("GET")
	r.HandleFunc("/v/{id}.mp4", a.videoHandler).Methods("GET")
	r.HandleFunc("/t/{id}", a.thumbHandler).Methods("GET")
	r.HandleFunc("/{id}", a.pageHandler).Methods("GET")
	fsHandler := http.StripPrefix(
		"/static/",
		http.FileServer(http.Dir("./static/")),
	)
	r.PathPrefix("/static/").Handler(fsHandler).Methods("GET")
	a.Router = r
	return a, nil
}

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

func (a *App) indexHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("/")
	pl := a.Library.Playlist()
	http.Redirect(w, r, "/"+pl[0].ID, 302)
}

func (a *App) pageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	log.Printf("/%s", id)
	playing, ok := a.Library.Videos[id]
	if !ok {
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
