package main

import (
	"fmt"
)

// TODO: Check if valid Youtube url. If not (minified urls), try following

/**
video <-goutube.Youtube(link)
args := video.args
downloadLink := video.
*/

func main() {
 	args := GetYoutubeConfigArgs("https://www.youtube.com/watch?v=iZq3i94mSsQ")
	videos := ExtractVideoFormats(args)
	fmt.Println(videos[0].URL)
}
