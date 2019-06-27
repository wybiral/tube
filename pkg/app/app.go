package app

import (
	"errors"
	"html/template"
	"log"
	"net"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/wybiral/tube/pkg/media"
)

type App struct {
	Config    *Config
	Library   *media.Library
	Playlist  media.Playlist
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
	lib := media.NewLibrary()
	err := lib.Import(cfg.LibraryPath)
	if err != nil {
		return nil, err
	}
	a.Library = lib
	pl := lib.Playlist()
	if len(pl) == 0 {
		return nil, errors.New("No valid videos found")
	}
	a.Playlist = pl
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
	return http.Serve(a.Listener, a.Router)
}

func (a *App) indexHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("/")
	http.Redirect(w, r, "/"+a.Playlist[0].ID, 302)
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
		Playlist: a.Playlist,
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
