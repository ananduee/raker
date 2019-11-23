package data

// Song represents a real world song.
type Song struct {
	ID    string
	Title string
	Album string
}

// Playlist is ranked list of songs.
type Playlist struct {
	Provider string
	Type     string
	Songs    []Song
}
