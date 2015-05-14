package main

import (
	"fmt"
)

// TODO: Check if valid Youtube url. If not (minified urls), try following them

/**
video <- goutube.Youtube(link)
args := video.args
downloadLink := video.
*/
func main() {
 	args, err := GetYoutubeConfigArgs("https://www.youtube.com/watch?v=iZq3i94mSsQ")
	if err != nil {
		return
	}
	results, _ := GetResults(args)

	for _, result := range results {
		fmt.Println(result.Format.Type)
	}
}
