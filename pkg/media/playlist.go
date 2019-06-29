package media

type Playlist []*Video

func (p Playlist) Len() int {
	return len(p)
}

func (p Playlist) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p Playlist) Less(i, j int) bool {
	return p[i].Timestamp.After(p[j].Timestamp)
}
