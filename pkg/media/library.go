package media

import (
	"io/ioutil"
	"sort"
	"strings"
)

type Library struct {
	Videos map[string]*Video
}

func NewLibrary() *Library {
	lib := &Library{
		Videos: make(map[string]*Video),
	}
	return lib
}

func (lib *Library) Import(path string) error {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}
	for _, info := range files {
		name := info.Name()
		v, err := ParseVideo(path + "/" + name)
		if err != nil {
			// Ignore files that can't be parsed
			continue
		}
		// Set modified date property
		v.Modified = info.ModTime().Format("2006-01-02")
		// Default title is filename
		if v.Title == "" {
			v.Title = name
		}
		// ID is name without extension
		idx := strings.LastIndex(name, ".")
		if idx == -1 {
			idx = len(name)
		}
		v.ID = name[:idx]
		lib.Videos[v.ID] = v
	}
	return nil
}

func (lib *Library) Playlist() Playlist {
	pl := make(Playlist, 0)
	for _, v := range lib.Videos {
		pl = append(pl, v)
	}
	sort.Sort(pl)
	return pl
}
