package rankings

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"

	"github.com/gocolly/colly"
)

// AmazonMusicSong represents one song
type AmazonMusicSong struct {
	Title string
	Album string
	Rank  int
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

// GetAmazonMusicTopSongs returns ranked list of songs.
func GetAmazonMusicTopSongs() []AmazonMusicSong {
	songs := []AmazonMusicSong{}
	c := colly.NewCollector()

	// Amazon music is slightly complicated as it gives bunch of device specific
	// information in javascript which is later used to fetch list of songs.
	c.OnHTML("script", func(e *colly.HTMLElement) {
		scriptContent := strings.Trim(e.DOM.Text(), " \n")
		if !strings.HasPrefix(scriptContent, "var applicationContextConfiguration") {
			return
		}
		applicationConfig := strings.Replace(strings.SplitN(scriptContent, "\n", 2)[0], "var applicationContextConfiguration =", "", 1)
		if strings.HasSuffix(applicationConfig, ";") {
			applicationConfig = strings.ReplaceAll(applicationConfig, ";", "")
		}
		var applicationConfigJSON map[string]interface{}
		json.Unmarshal([]byte(applicationConfig), &applicationConfigJSON)

		deviceID := applicationConfigJSON["deviceId"].(string)
		deviceType := applicationConfigJSON["deviceType"].(string)
		sessionID := applicationConfigJSON["sessionId"].(string)
		fmt.Println("deviceId", deviceID, "deviceType", deviceType, "sessionId", sessionID)

		requestFormat := `
			{
				"asins": [
					"B081HYGQ5S"
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

		requestBody := processString(requestFormat, map[string]interface{}{"DeviceID": deviceID, "DeviceType": deviceType, "SessionId": sessionID})

		postColly := c.Clone()
		postColly.OnRequest(func(r *colly.Request) {
			r.Headers.Set("Content-Type", "application/json")
			csrfTokenConfig := applicationConfigJSON["CSRFTokenConfig"].(map[string]interface{})
			r.Headers.Set("csrf-token", csrfTokenConfig["csrf_token"].(string))
			r.Headers.Set("csrf-rnd", csrfTokenConfig["csrf_rnd"].(string))
			r.Headers.Set("csrf-ts", csrfTokenConfig["csrf_ts"].(string))
			r.Headers.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.14; rv:68.0) Gecko/20100101 Firefox/68.0")
			r.Headers.Set("X-Requested-With", "XMLHttpRequest")
			r.Headers.Set("X-Amz-Target", "com.amazon.musicensembleservice.MusicEnsembleService.lookup")
			r.Headers.Set("Referer", "https://music.amazon.in/playlists/B081J5DZ5W")
			r.Headers.Set("Content-Encoding", "amz-1.0")
		})
		postColly.OnResponse(func(r *colly.Response) {
			var songsResponseJSON amazonMusicJSONResponse
			if r.StatusCode == 200 {
				json.Unmarshal(r.Body, &songsResponseJSON)
				if songsResponseJSON.PlaylistList != nil && songsResponseJSON.PlaylistList[0].Tracks != nil {
					for _, song := range songsResponseJSON.PlaylistList[0].Tracks {
						localSong := AmazonMusicSong{
							Title: song.Title,
							Album: song.Album.Title,
							Rank:  song.ItemIndex + 1,
						}
						songs = append(songs, localSong)
					}
				} else {
					js, _ := json.Marshal(songsResponseJSON)
					fmt.Println("playListList is nill or tracks is nil", "status", string(js), string(r.Body))
				}
			} else {
				fmt.Println("status code is not 200. body", string(r.Body), "status", r.StatusCode)
			}
		})
		err := postColly.PostRaw("https://music.amazon.in/EU/api/muse/legacy/lookup", []byte(requestBody))
		postColly.Wait()

		if err != nil {
			panic(err)
		}
	})

	c.Visit("https://music.amazon.in/playlists/B081HYGQ5S")
	c.Wait()
	return songs
}

func processString(str string, vars interface{}) string {
	tmpl, err := template.New("tmpl").Parse(str)

	if err != nil {
		panic(err)
	}
	return process(tmpl, vars)
}

func process(t *template.Template, vars interface{}) string {
	var tmplBytes bytes.Buffer

	err := t.Execute(&tmplBytes, vars)
	if err != nil {
		panic(err)
	}
	return tmplBytes.String()
}
