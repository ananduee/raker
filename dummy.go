package main

import (
	"encoding/json"
	"fmt"

	"github.com/ananduee/raker/controllers/rankings"
)

func main() {
	f := rankings.MirchiTop20Fetcher{}
	songs, err := f.Get()
	if err != nil {
		panic(err)
	}
	out, err := json.Marshal(songs)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(out))
}
