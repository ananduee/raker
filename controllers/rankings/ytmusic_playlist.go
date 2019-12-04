package rankings

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/ananduee/raker/data"
	"github.com/gocolly/colly"
	"github.com/corpix/uarand"
)

// YTMusicPlaylistFetcher fetches songs in a plylist from yt music.
type YTMusicPlaylistFetcher struct {
	URL string
}

type rankedSong struct {
	Rank int
	Song data.Song
}

type ytcConfig struct {
	InitialEndpoint struct {
		BrowseEndpoint struct {
			BrowseID string `json:"browseId"`
		} `json:"browseEndpoint"`
	} `json:"INITIAL_ENDPOINT"`
	PageBuildLabel string `json:"PAGE_BUILD_LABEL"`
	PageCL         string `json:"PAGE_CL"`
	ClientName     string `json:"INNERTUBE_CONTEXT_CLIENT_NAME"`
	ClientVersion  string `json:"INNERTUBE_CLIENT_VERSION"`
	VisitorID      string `json:"VISITOR_DATA"`
	APIKey         string `json:"INNERTUBE_API_KEY"`
}

// Get list of top 20 songs from mirchi
func (y *YTMusicPlaylistFetcher) Get() (*data.Playlist, error) {
	songs := []rankedSong{}
	c := colly.NewCollector()
	var callbackErr error

	c.OnHTML("script ", func(e *colly.HTMLElement) {
		fmt.Println("inside script tag.")
		if callbackErr != nil {
			return
		}

		scriptContent := strings.Trim(e.DOM.Text(), " \n")
		if !strings.Contains(scriptContent, "INNERTUBE_CONTEXT_CLIENT_NAME") {
			fmt.Println("script does not have variable defined skipping.")
			return
		}
		fmt.Println("string has variable present parsing now.")

		config, err := parseYtcConfig(scriptContent)
		if err != nil {
			callbackErr = err
			return
		}

		requestBody := getRequestBody(config.InitialEndpoint.BrowseEndpoint.BrowseID)

		postColly := c.Clone()
		postColly.OnRequest(func(r *colly.Request) {
			r.Headers.Set("X-YouTube-Utc-Offset", "330")
			r.Headers.Set("X-YouTube-Page-Label", config.PageBuildLabel)
			r.Headers.Set("X-YouTube-Page-CL", config.PageCL)
			r.Headers.Set("X-YouTube-Client-Name", config.ClientName)
			r.Headers.Set("X-YouTube-Client-Version", config.ClientVersion)
			r.Headers.Set("X-Goog-Visitor-Id", config.VisitorID)
			r.Headers.Set("Host", "music.youtube.com")
			r.Headers.Set("User-Agent", uarand.GetRandom())
			r.Headers.Set("Referer", y.URL)
			r.Headers.Set("Content-Type", "application/json")
			r.Headers.Set("Accept", "*/*")
		})
		postColly.OnResponse(func(r *colly.Response) {
			fmt.Println("body got",string(r.Body))
		})
		apiURL, err := getPlaylistAPIURL(y, config)
		if err != nil {
			callbackErr = err
			return
		}
		err = postColly.PostRaw(apiURL.String(), []byte(requestBody))
		postColly.Wait()

		if err != nil {
			callbackErr = err
		}
	})

	c.OnResponse(func(response *colly.Response) {
		//fmt.Println("response body", string(response.Body))
	})

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("User-Agent", uarand.GetRandom())
		fmt.Println("Visiting", r.URL)
		fmt.Println("UserAgent", r.Headers.Get("User-Agent"))
	})

	c.Visit(y.URL)
	c.Wait()

	if callbackErr != nil {
		return nil, callbackErr
	}

	sort.Slice(songs, func(i, j int) bool {
		return songs[i].Rank < songs[j].Rank
	})
	songsToReturn := make([]data.Song, len(songs))
	for i, song := range songs {
		songsToReturn[i] = song.Song
	}
	return &data.Playlist{}, nil
}

func parseYtcConfig(str string) (*ytcConfig, error) {
	startPos := strings.Index(str, "{")
	lastPos := strings.LastIndex(str, "}")
	if startPos != -1 && lastPos != -1 {
		cleanStr := str[startPos : lastPos+1]
		jsToJSONReplace := strings.NewReplacer("null", "\"null\"", "\"{", "{", "}\"", "}", "\\\"", "\"")
		parsedJSON := jsToJSONReplace.Replace(cleanStr)
		var config ytcConfig
		json.Unmarshal([]byte(parsedJSON), &config)
		fmt.Println("succesfully parsed js variables.")
		return &config, nil
	}
	return nil, data.NewError("temporary", "Unable to find json values")
}

func getPlaylistAPIURL(y *YTMusicPlaylistFetcher, config *ytcConfig) (*url.URL, error) {
	playlistURL, err := url.Parse(y.URL)
	if err != nil {
		return nil, err
	}
	playlistURL.Path = "youtubei/v1/browse"
	queryParams := playlistURL.Query()
	queryParams.Del("list")
	queryParams.Set("key", config.APIKey)
	queryParams.Set("alt", "json")
	playlistURL.RawQuery = queryParams.Encode()
	return playlistURL, nil
}

func getRequestBody(browseID string) string {
	return `
	{
		"context": {
			"client": {
				"clientName": "WEB_REMIX",
				"clientVersion": "0.1",
				"hl": "en",
				"gl": "IN",
				"experimentIds": [],
				"experimentsToken": "",
				"utcOffsetMinutes": 330,
				"locationInfo": {
					"locationPermissionAuthorizationStatus": "LOCATION_PERMISSION_AUTHORIZATION_STATUS_UNSUPPORTED"
				},
				"musicAppInfo": {
					"musicActivityMasterSwitch": "MUSIC_ACTIVITY_MASTER_SWITCH_INDETERMINATE",
					"musicLocationMasterSwitch": "MUSIC_LOCATION_MASTER_SWITCH_INDETERMINATE",
					"pwaInstallabilityStatus": "PWA_INSTALLABILITY_STATUS_UNKNOWN"
				}
			},
			"capabilities": {},
			"request": {
				"internalExperimentFlags": [
					{
						"key": "force_music_enable_outertube_playlist_detail_browse",
						"value": "true"
					},
					{
						"key": "force_music_enable_outertube_tastebuilder_browse",
						"value": "true"
					},
					{
						"key": "force_music_enable_outertube_search_suggestions",
						"value": "true"
					}
				],
				"sessionIndex": {}
			},
			"activePlayers": {},
			"user": {
				"enableSafetyMode": false
			}
		},
		"browseId": "VLPL4fGSI1pDJn40WjZ6utkIuj2rNg-7iGsq",
		"browseEndpointContextSupportedConfigs": {
			"browseEndpointContextMusicConfig": {
				"pageType": "MUSIC_PAGE_TYPE_PLAYLIST"
			}
		}
	}
	`
}
