package main

import (
	"encoding/json"
	"fmt"

	"github.com/ananduee/raker/controllers/rankings"
)

func main() {
	ganna := rankings.AmazonMusicPlaylistFetcher{
		URL:          "https://music.amazon.in/playlists/B081HYGQ5S",
		MaxSongs:     50,
		PlaylistType: "living_list",
		Asin:         "B081HYGQ5S",
		LookupURL:    "https://music.amazon.in/EU/api/muse/legacy/lookup",
	}
	songs, err := ganna.Get()
	if err != nil {
		panic(err)
	}
	out, err := json.Marshal(songs)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(out))
}
