package media

import (
	"io/ioutil"
	"log"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

type Library struct {
	mu     sync.RWMutex
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
		err = lib.Add(path + "/" + info.Name())
		if err != nil {
			// Ignore files that can't be parsed
			continue
		}
	}
	return nil
}

func (lib *Library) Add(path string) error {
	v, err := ParseVideo(path)
	if err != nil {
		return err
	}
	lib.mu.Lock()
	defer lib.mu.Unlock()
	lib.Videos[v.ID] = v
	log.Println("Added:", path)
	return nil
}

func (lib *Library) Remove(path string) {
	name := filepath.Base(path)
	// ID is name without extension
	idx := strings.LastIndex(name, ".")
	if idx == -1 {
		idx = len(name)
	}
	id := name[:idx]
	lib.mu.Lock()
	defer lib.mu.Unlock()
	_, ok := lib.Videos[id]
	if ok {
		delete(lib.Videos, id)
		log.Println("Removed:", path)
	}
}

func (lib *Library) Playlist() Playlist {
	lib.mu.RLock()
	defer lib.mu.RUnlock()
	pl := make(Playlist, 0)
	for _, v := range lib.Videos {
		pl = append(pl, v)
	}
	sort.Sort(pl)
	return pl
}
