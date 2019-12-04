package data

import "testing"

func TestEmptyList(t *testing.T) {
	list := NewSortedPlayList()
	value := list.ToSlice()
	if len(value) != 0 {
		t.Errorf("Expected an empty slice. Got slicne of size %d", len(value))
	}
}

func TestListOfSizeOne(t *testing.T) {
	list := NewSortedPlayList()
	list.Add(&Song{Title: "rank 2"}, 2)
	list.Add(&Song{Title: "rank 1"}, 1)
	//	value := list.ToSlice()
}
