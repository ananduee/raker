package main

import (
	"encoding/json"
	"fmt"

	"github.com/ananduee/raker/controllers/rankings"
)

func main() {
	ganna := rankings.SaavnFetcher{
		URL: "https://www.jiosaavn.com/featured/weekly-top-songs/8MT-LQlP35c_",
	}
	songs, err := ganna.Get()
	if err != nil {
		panic(err)
	}
	out, err := json.Marshal(songs)
	if err != nil {
		panic(err)
	}

	fmt.Println("this is just aamzing")

	fmt.Println(string(out))
}
