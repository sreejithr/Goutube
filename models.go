package goutube

import "fmt"

type MediaType int

const (
	VIDEO MediaType = iota
	AUDIO
)

func (t MediaType) String() string {
	var s string
	switch t {
	case VIDEO:
		s = "VIDEO"
	case AUDIO:
		s = "AUDIO"
	}
	return s
}

type MediaFormat struct {
	Type         string
	VideoCodec   string
	AudioCodec   string
	Raw          string
}

type Link struct {
	URL          string
	Type         MediaType
	Signature    string
	Quality      string
	Format       MediaFormat
}

func (t Link) String() string {
	quality := t.Quality
	if len(quality) == 0 {
		quality = "NA"
	}
	s := fmt.Sprintf("URL: %s,\n Type: %s,\n Quality: %s,\n Format(Raw): %s\n",
		t.URL, t.Type, quality, t.Format.Raw)
	return s
}

type Result struct {
	Done         chan bool
	Error        chan error
	Links        []Link
}
