package media

import (
	"os"
	"strings"

	"github.com/dhowden/tag"
)

type Video struct {
	ID          string
	Title       string
	Album       string
	Description string
	Thumb       []byte
	ThumbType   string
	Modified    string
}

func ParseVideo(path string) (*Video, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	modified := info.ModTime().Format("2006-01-02")
	name := info.Name()
	// ID is name without extension
	idx := strings.LastIndex(name, ".")
	if idx == -1 {
		idx = len(name)
	}
	id := name[:idx]
	m, err := tag.ReadFrom(f)
	if err != nil {
		return nil, err
	}
	title := m.Title()
	// Default title is filename
	if title == "" {
		title = name
	}
	v := &Video{
		ID:          id,
		Title:       title,
		Album:       m.Album(),
		Description: m.Comment(),
		Modified:    modified,
	}
	// Add thumbnail (if exists)
	p := m.Picture()
	if p != nil {
		v.Thumb = p.Data
		v.ThumbType = p.MIMEType
	}
	return v, nil
}
