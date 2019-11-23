package main

import (
	"encoding/json"
	"fmt"

	"github.com/ananduee/raker/controllers/rankings"
)

func main() {
	ganna := rankings.GannaPlaylistFetcherFetcher{
		URL:          "https://gaana.com/playlist/gaana-dj-bollywood-top-50-1",
		MaxSongs:     50,
		PlaylistType: "living_list",
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
