package rankings

import (
	"fmt"
	"strconv"

	"github.com/ananduee/raker/data"
	"github.com/gocolly/colly"
)

// SaavnFetcher represents configuration object for fetching songs from saavn.
type SaavnFetcher struct {
	URL string
}

// Get list of songs from a saavn page
func (p *SaavnFetcher) Get() (*data.Playlist, error) {
	playlist := data.NewSortedPlayList()

	c := colly.NewCollector()
	var callbackErr error

	c.OnHTML(".song-wrap", func(e *colly.HTMLElement) {
		if callbackErr != nil {
			return
		}

		song := data.Song{}
		song.Title = e.ChildText(".title > a")
		song.Album = e.ChildText(".meta-album")
		rank, err := strconv.Atoi(e.ChildText(".index"))
		if err != nil {
			callbackErr = data.NewError("temporary", fmt.Sprintf("Failed to find rank for song: %s", song.Title))
		} else {
			playlist.Add(&song, rank)
		}
	})

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.14; rv:68.0) Gecko/20100101 Firefox/68.0")
	})

	c.Visit(p.URL)
	c.Wait()

	if callbackErr != nil {
		return nil, callbackErr
	}

	return &data.Playlist{
		Provider: "saavn",
		Type:     "living_list",
		Songs:    playlist.ToSlice(),
	}, nil
}
