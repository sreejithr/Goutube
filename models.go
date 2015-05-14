package main

type ResultType int

const (
	VIDEO ResultType = iota
	AUDIO
)

func (t ResultType) String() string {
	var s string
	switch t {
	case VIDEO:
		s = "VIDEO"
	case AUDIO:
		s = "AUDIO"
	}
	return s
}

type ResultFormat struct {
	Type         string
	VideoCodec   string
	AudioCodec   string
	Raw          string
}

type Result struct {
	URL          string
	Type         ResultType
	Signature    string
	Quality      string
	Format       ResultFormat
}
