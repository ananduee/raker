package data

// Error represents error in performing an action. Based on the code error can be retried
type Error struct {
	Err  string
	Code string
}

// MusicRankingFetcher will fetch list of songs
type MusicRankingFetcher interface {
	Get() (Playlist, error)
}

func (e *Error) Error() string {
	return e.Err
}

// NewError creates new instance of error
func NewError(code string, msg string) *Error {
	return &Error{
		Err:  msg,
		Code: code,
	}
}
