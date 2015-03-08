package goutube

type Video interface {
	Link string
}

type Result struct {
	Mp4 Video
}

func Youtube(url string) <-chan Result {
	// Youtube returns a channel which pumps a `Result` whenever it is ready.
	// ...
	result := make(chan Result)

	// TODO: Check if valid Youtube url. If not (minified urls), try following

	
}

