# Goutube
Go library to get download links for Youtube videos.

## Features

* Extracts links for below formats:
	* video/webm
	* video/mp4
	* video/x-flv
	* video/3gpp
	* audio/webm
	* audio/mp4 (MPEG)


## Installation

`go get github.com/sreejithr/goutube`


## Usage

Below is the barebones basic.

```
import "github.com/sreejithr/goutube"

// Channel to report failure
errorChan := make(chan error)

resultChan := goutube.Youtube("https://www.youtube.com/watch?v=dQw4w9WgXcQ")

for _, link := range <-resultChan {
	// `link` is a struct btw
	fmt.Println(link)
}
```

With error management.

```
// Channel to report failure
errorChan := make(chan error)

resultChan := goutube.Youtube("https://www.youtube.com/watch?v=dQw4w9WgXcQ")

select {
case links := <-resultChan:
    for _, link := range <-resultChan {
    	fmt.Println(link)
    }
case err := <-errorChan:
    fmt.Println(err)
}
```
## API

### Constants
```
type MediaType int
const (
	VIDEO MediaType = iota
	AUDIO
)
```

### func Youtube

`func Youtube(link string, err chan error) result chan Link`

Get download links for the given youtube url. Might return multiple links of different quality and format. Check `Type` of link which maybe `goutube.VIDEO` or `goutube.AUDIO`

### type Link

```
type Link struct {
	URL          string
	Type         MediaType
	Signature    string
	Quality      string
	Format       MediaFormat
}
```

### type MediaFormat

```
type MediaFormat struct {
	Type         string
	VideoCodec   string
	AudioCodec   string
	Raw          string
}
```

## TODO

* Better error management (especially in network outage).
* Download from minified URLs.
* Return file size of each link.
* Download all videos from a page (embedded in pages and Youtube search results)
* Download whole playlists.

## Contributing

* Can I?
	* Yes, please :)
	
* How?
	* Feel free to post issues, take up issues, ask questions, propose features.

* I'm new to open source. So...
	* [Learn Git](http://git-scm.com/book/en/Git-Basics).
	* Fork the repo.
	* Make pull request.
	* Welcome aboard :)