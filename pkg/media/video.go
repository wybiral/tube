package media

import (
	"os"
	"strings"
	"time"

	"github.com/dhowden/tag"
)

// Video represents metadata for a single video.
type Video struct {
	ID          string
	Title       string
	Album       string
	Description string
	Thumb       []byte
	ThumbType   string
	Modified    string
	Size        int64
	Timestamp   time.Time
}

// ParseVideo parses a video file's metadata and returns a Video.
func ParseVideo(path string) (*Video, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	size := info.Size()
	timestamp := info.ModTime()
	modified := timestamp.Format("2006-01-02 03:04 PM")
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
		Size:        size,
		Timestamp:   timestamp,
	}
	// Add thumbnail (if exists)
	p := m.Picture()
	if p != nil {
		v.Thumb = p.Data
		v.ThumbType = p.MIMEType
	}
	return v, nil
}
