package data

// Node represents one point in linked lit.
type Node struct {
	Song *Song
	Next *Node
	Rank int
}

// SortedPlayList stores rank wise songs
type SortedPlayList struct {
	root *Node
}

// NewSortedPlayList creates an empty playlist.
func NewSortedPlayList() *SortedPlayList {
	return &SortedPlayList{root: nil}
}

// Add a new element to the list.
func (list *SortedPlayList) Add(song *Song, rank int) {
	if list.root == nil {
		list.root = &Node{
			Song: song,
			Rank: rank,
		}
		return
	}
	var prevNode *Node = nil
	nextNode := list.root
	// find nodes between which this node should be inserted.
	for nextNode != nil && rank > nextNode.Rank {
		prevNode = nextNode
		nextNode = nextNode.Next
	}
	newNode := &Node{
		Song: song,
		Rank: rank,
		Next: nextNode,
	}
	if prevNode == nil {
		// we need to change the root.
		list.root = newNode
	} else {
		prevNode.Next = newNode
	}
}

// ToSlice converts linked list to a slice
func (list *SortedPlayList) ToSlice() []Song {
	songs := make([]Song, 0)
	root := list.root
	for root != nil {
		songs = append(songs, *root.Song)
		root = root.Next
	}
	return songs
}
