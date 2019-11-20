package rankings

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/gocolly/colly"
)

// GannaSong represents one song from mirchi.
type GannaSong struct {
	ID    string
	Title string
	Album string
	Rank  int
}

// GetGaanaHindiSongs returns ranked list of songs.
func GetGaanaHindiSongs() []GannaSong {
	songs := []GannaSong{}
	songsAlbum := make(map[string]string)
	c := colly.NewCollector()
	c.OnHTML(".s_l", func(e *colly.HTMLElement) {
		dataValue := e.Attr("data-value")
		if !strings.HasPrefix(dataValue, "song") {
			return
		}

		song := GannaSong{}
		song.ID = strings.Replace(dataValue, "song", "", 1)
		song.Title = e.ChildText(".playlist_thumb_det > .sng_c")
		rank, err := strconv.Atoi(e.ChildText("._c"))
		if err == nil {
			song.Rank = rank
		}
		songs = append(songs, song)
	})
	c.OnHTML("span", func(e *colly.HTMLElement) {
		spanID := e.Attr("id")
		if !strings.HasPrefix(spanID, "parent-row-song") {
			return
		}
		var songJSON map[string]interface{}
		json.Unmarshal([]byte(e.DOM.Text()), &songJSON)
		songID := songJSON["id"].(string)
		albumName := songJSON["albumtitle"].(string)
		songsAlbum[songID] = albumName
	})
	c.Visit("https://gaana.com/playlist/gaana-dj-bollywood-top-50-1")
	c.Wait()
	for i, song := range songs {
		album := songsAlbum[song.ID]
		songs[i].Album = album
	}

	return songs
}
