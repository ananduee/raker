package rankings

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ananduee/raker/data"
	"github.com/gocolly/colly"
)

// MirchiTop20Fetcher will implement MusicRankingFetcher for mirchi top20.
type MirchiTop20Fetcher struct {
}

// Get list of top 20 songs from mirchi
func (f *MirchiTop20Fetcher) Get() (*data.Playlist, error) {
	songs := make([]data.Song, 20)
	c := colly.NewCollector()
	var callbackErr *data.Error
	c.OnHTML(".top01", func(e *colly.HTMLElement) {
		if callbackErr != nil {
			return
		}
		song := data.Song{}
		song.Title = e.ChildText("h2")
		albumAndArtist := e.ChildText("h3")
		song.Album = strings.Split(albumAndArtist, "\n")[0]
		rank, err := strconv.Atoi(e.ChildText(".circle"))
		if err == nil {
			songs[rank-1] = song
		} else {
			callbackErr = data.NewError("temporary", fmt.Sprintf("Failed to find rank for song: %s", song.Title))
		}
	})
	if callbackErr != nil {
		return nil, callbackErr
	}
	c.Visit("https://www.radiomirchi.com/more/mirchi-top-20/")
	c.Wait()
	return &data.Playlist{
		Provider: "mirchi",
		Type:     "living_list",
		Songs:    songs,
	}, nil
}
