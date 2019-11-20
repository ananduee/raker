package main

import (
	"encoding/json"
	"fmt"

	"github.com/ananduee/raker/controllers/rankings"
)

func main() {
	songs := rankings.GetAmazonMusicTopSongs()
	out, err := json.Marshal(songs)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(out))
}
