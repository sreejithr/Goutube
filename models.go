package goutube

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

type Result struct {
	Done         chan bool
	Error        chan error
	Links        []Link
}
