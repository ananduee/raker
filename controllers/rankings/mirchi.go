package rankings

import (
	"strconv"
	"strings"

	"github.com/gocolly/colly"
)

// MirchiSong represents one song from mirchi.
type MirchiSong struct {
	Title string
	Album string
	Rank  int
}

// GetMirchiHindiSongs returns ranked list of songs.
func GetMirchiHindiSongs() []MirchiSong {
	songs := []MirchiSong{}
	c := colly.NewCollector()
	c.OnHTML(".top01", func(e *colly.HTMLElement) {
		song := MirchiSong{}
		song.Title = e.ChildText("h2")
		albumAndArtist := e.ChildText("h3")
		song.Album = strings.Split(albumAndArtist, "\n")[0]
		rank, err := strconv.Atoi(e.ChildText(".circle"))
		if err == nil {
			song.Rank = rank
		}
		songs = append(songs, song)
	})
	c.Visit("https://www.radiomirchi.com/more/mirchi-top-20/")
	c.Wait()
	return songs
}
