package main

import (
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/wybiral/tube/pkg/media"
)

const addr = "127.0.0.1:40404"

var templates *template.Template

var library *media.Library
var playlist media.Playlist

func main() {
	library = media.NewLibrary()
	err := library.Import("./videos")
	if err != nil {
		log.Fatal(err)
	}
	playlist = library.Playlist()
	if len(playlist) == 0 {
		log.Fatal("No valid videos found")
	}
	templates = template.Must(template.ParseGlob("templates/*"))
	r := mux.NewRouter().StrictSlash(true)
	r.HandleFunc("/", index).Methods("GET")
	r.HandleFunc("/v/{id}.mp4", video).Methods("GET")
	r.HandleFunc("/t/{id}", thumb).Methods("GET")
	r.HandleFunc("/{id}", page).Methods("GET")
	r.PathPrefix("/static/").Handler(
		http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))),
	).Methods("GET")
	log.Printf("Serving at %s", addr)
	http.ListenAndServe(addr, r)
}

func index(w http.ResponseWriter, r *http.Request) {
	log.Printf("/index")
	http.Redirect(w, r, "/"+playlist[0].ID, 302)
}

func page(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	log.Printf(id)
	playing, ok := library.Videos[id]
	if !ok {
		log.Print(ok)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	templates.ExecuteTemplate(w, "index.html", &struct {
		Playing  *media.Video
		Playlist media.Playlist
	}{
		Playing:  playing,
		Playlist: playlist,
	})
}

func video(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	log.Print("/v/" + id)
	m, ok := library.Videos[id]
	if !ok {
		return
	}
	title := m.Title
	disposition := "attachment; filename=\"" + title + ".mp4\""
	w.Header().Set("Content-Disposition", disposition)
	w.Header().Set("Content-Type", "video/mp4")
	http.ServeFile(w, r, "./videos/"+id+".mp4")
}

func thumb(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	log.Printf("/t/" + id)
	m, ok := library.Videos[id]
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
