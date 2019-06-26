package media

import (
	"os"

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
	m, err := tag.ReadFrom(f)
	if err != nil {
		return nil, err
	}
	v := &Video{
		Title:       m.Title(),
		Album:       m.Album(),
		Description: m.Comment(),
	}
	// Add thumbnail (if exists)
	p := m.Picture()
	if p != nil {
		v.Thumb = p.Data
		v.ThumbType = p.MIMEType
	}
	return v, nil
}
