package rankings

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"

	"github.com/ananduee/raker/data"
	"github.com/gocolly/colly"
)

// AmazonMusicPlaylistFetcher will fetch all songs from gaana playlist
type AmazonMusicPlaylistFetcher struct {
	URL          string
	Asin         string // Asin code of playlist
	MaxSongs     int
	PlaylistType string
	LookupURL    string
}

type amazonMusicJSONResponse struct {
	PlaylistList []struct {
		Tracks []struct {
			Album struct {
				Title string `json:"title"`
			} `json:"album"`
			Title     string `json:"title"`
			ItemIndex int    `json:"itemIndex"`
		} `json:"tracks"`
	} `json:"playlistList"`
}

type amazonApplicationContextConfiguration struct {
	DeviceID        string `json:"deviceId"`
	DeviceType      string `json:"deviceType"`
	SessionID       string `json:"sessionId"`
	CSRFTokenConfig struct {
		CSRFToken string `json:"csrf_token"`
		CSRFRnd   string `json:"csrf_rnd"`
		CSRFTs    string `json:"csrf_ts"`
	} `json:"CSRFTokenConfig"`
}

// Get list of songs from configured playlist
func (p *AmazonMusicPlaylistFetcher) Get() (*data.Playlist, error) {
	songs := make([]data.Song, p.MaxSongs)
	c := colly.NewCollector()
	var callbackErr error

	// Amazon music is slightly complicated as it gives bunch of device specific
	// information in javascript which is later used to fetch list of songs.
	c.OnHTML("script", func(e *colly.HTMLElement) {
		if callbackErr != nil {
			return
		}
		scriptContent := strings.Trim(e.DOM.Text(), " \n")
		if !strings.HasPrefix(scriptContent, "var applicationContextConfiguration") {
			return
		}
		applicationConfig := strings.Replace(strings.SplitN(scriptContent, "\n", 2)[0], "var applicationContextConfiguration =", "", 1)
		if strings.HasSuffix(applicationConfig, ";") {
			applicationConfig = strings.ReplaceAll(applicationConfig, ";", "")
		}
		var applicationConfigJSON amazonApplicationContextConfiguration
		json.Unmarshal([]byte(applicationConfig), &applicationConfigJSON)

		deviceID := applicationConfigJSON.DeviceID
		deviceType := applicationConfigJSON.DeviceType
		sessionID := applicationConfigJSON.SessionID

		requestFormat := `
			{
				"asins": [
					"{{.Asin}}"
				],
				"features": [
					"collectionLibraryAvailability",
					"expandTracklist",
					"playlistLibraryAvailability",
					"trackLibraryAvailability",
					"hasLyrics"
				],
				"requestedContent": "PRIME",
				"deviceId": "{{.DeviceID}}",
				"deviceType": "{{.DeviceType}}",
				"musicTerritory": "IN",
				"sessionId": "{{.SessionId}}"
			}
		`

		requestBody, err := processString(requestFormat, map[string]interface{}{"DeviceID": deviceID, "DeviceType": deviceType, "SessionId": sessionID, "Asin": p.Asin})

		if err != nil {
			callbackErr = err
			return
		}

		postColly := c.Clone()
		postColly.OnRequest(func(r *colly.Request) {
			r.Headers.Set("Content-Type", "application/json")
			csrfTokenConfig := applicationConfigJSON.CSRFTokenConfig
			r.Headers.Set("csrf-token", csrfTokenConfig.CSRFToken)
			r.Headers.Set("csrf-rnd", csrfTokenConfig.CSRFRnd)
			r.Headers.Set("csrf-ts", csrfTokenConfig.CSRFTs)
			r.Headers.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.14; rv:68.0) Gecko/20100101 Firefox/68.0")
			r.Headers.Set("X-Requested-With", "XMLHttpRequest")
			r.Headers.Set("X-Amz-Target", "com.amazon.musicensembleservice.MusicEnsembleService.lookup")
			r.Headers.Set("Referer", p.URL)
			r.Headers.Set("Content-Encoding", "amz-1.0")
		})
		postColly.OnResponse(func(r *colly.Response) {
			var songsResponseJSON amazonMusicJSONResponse
			if r.StatusCode == 200 {
				json.Unmarshal(r.Body, &songsResponseJSON)
				if songsResponseJSON.PlaylistList != nil && songsResponseJSON.PlaylistList[0].Tracks != nil {
					for _, song := range songsResponseJSON.PlaylistList[0].Tracks {
						localSong := data.Song{
							Title: song.Title,
							Album: song.Album.Title,
						}
						songs[song.ItemIndex] = localSong
					}
				} else {
					js, _ := json.Marshal(songsResponseJSON)
					callbackErr = data.NewError("temporary", fmt.Sprintf("playListList is nill or tracks is nil js - %s status - %s", string(js), string(r.Body)))
				}
			} else {
				callbackErr = data.NewError("temporary", fmt.Sprintf("status code is not 200. body %s code %d", string(r.Body), r.StatusCode))
			}
		})
		err = postColly.PostRaw(p.LookupURL, []byte(requestBody))
		postColly.Wait()

		if err != nil {
			callbackErr = err
		}
	})

	c.Visit(p.URL)
	c.Wait()

	if callbackErr != nil {
		return nil, callbackErr
	}

	return &data.Playlist{
		Provider: "amazon",
		Type:     p.PlaylistType,
		Songs:    songs,
	}, nil
}

func processString(str string, vars interface{}) (string, error) {
	tmpl, err := template.New("tmpl").Parse(str)

	if err != nil {
		return "", err
	}

	return process(tmpl, vars)
}

func process(t *template.Template, vars interface{}) (string, error) {
	var tmplBytes bytes.Buffer

	err := t.Execute(&tmplBytes, vars)
	if err != nil {
		return "", nil
	}
	return tmplBytes.String(), nil
}
