package goutube

import "testing"

func TestYoutubeLinkFetching(t *testing.T) {
	errorChan := make(chan error)
	result := Youtube("https://www.youtube.com/watch?v=iZq3i94mSsQ",
		errorChan)

	select {
	case _ = <-result:
	case err := <-errorChan:
		t.Error(err)
	}
}
