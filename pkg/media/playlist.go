package media

// Playlist holds an array of videos capable of sorting by Timestamp.
type Playlist []*Video

// Len returns length of array (for sorting).
func (p Playlist) Len() int {
	return len(p)
}

// Swap swaps two values in array by index (for sorting).
func (p Playlist) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

// Less returns true if p[i] Timestamp is after p[j] (for sorting).
func (p Playlist) Less(i, j int) bool {
	return p[i].Timestamp.After(p[j].Timestamp)
}
