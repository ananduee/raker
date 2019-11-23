package rankings

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/ananduee/raker/data"
	"github.com/gocolly/colly"
)

// GannaPlaylistFetcher will fetch all songs from gaana playlist
type GannaPlaylistFetcher struct {
	URL          string
	MaxSongs     int
	PlaylistType string
}

type gaanaSongSpanJSON struct {
	ID         string `json:"id"`
	AlbumTitle string `json:"albumtitle"`
}

// Get list of songs from configured playlist
func (p *GannaPlaylistFetcher) Get() (*data.Playlist, error) {
	songs := make([]data.Song, p.MaxSongs)
	songsAlbum := make(map[string]string)
	c := colly.NewCollector()
	var callbackErr *data.Error
	c.OnHTML(".s_l", func(e *colly.HTMLElement) {
		if callbackErr != nil {
			return
		}
		dataValue := e.Attr("data-value")
		if !strings.HasPrefix(dataValue, "song") {
			return
		}

		song := data.Song{}
		song.ID = strings.Replace(dataValue, "song", "", 1)
		song.Title = e.ChildText(".playlist_thumb_det > .sng_c")
		rank, err := strconv.Atoi(e.ChildText("._c"))
		if err != nil {
			callbackErr = data.NewError("temporary", fmt.Sprintf("Failed to find rank for song: %s", song.Title))
		} else if rank > p.MaxSongs {
			callbackErr = data.NewError("permanent", fmt.Sprintf("Rank found for song %s = %d is outside max songs range.", song.Title, rank))
		} else {
			songs[rank-1] = song
		}
	})
	c.OnHTML("span", func(e *colly.HTMLElement) {
		spanID := e.Attr("id")
		if !strings.HasPrefix(spanID, "parent-row-song") {
			return
		}
		var songJSON gaanaSongSpanJSON
		json.Unmarshal([]byte(e.DOM.Text()), &songJSON)
		songID := songJSON.ID
		albumName := songJSON.AlbumTitle
		songsAlbum[songID] = albumName
	})
	c.Visit(p.URL)
	c.Wait()
	if callbackErr != nil {
		return nil, callbackErr
	}
	for i, song := range songs {
		album := songsAlbum[song.ID]
		songs[i].Album = album
	}
	return &data.Playlist{
		Provider: "ganna",
		Type:     p.PlaylistType,
		Songs:    songs,
	}, nil
}
