package goutube

import (
	"fmt"
	"net/http"
	"io/ioutil"
	"strings"
	"encoding/json"
)

type YoutubeArgs struct {
	VideoID string
	VideoFormats string
	VideoAdaptFormats string
	VideoManifestURL string
	PlayerSource string
}

func getYoutubeArgs (url string) YoutubeArgs {
	resp, _ := http.Get(url)
	page, _ := ioutil.ReadAll(resp.Body)

	bodyTxt := string(page[:])
	startIndex := strings.Index(bodyTxt, "ytplayer.config = ") + 18

	openCount := 0
	closeCount := 0
	endIndex := startIndex
	for i, c := range bodyTxt[startIndex:] {
		if c == '{' {
			openCount++
		} else if c == '}' {
			closeCount++
		}
		if openCount == closeCount {
			endIndex = startIndex + i + 1
			break
		}
	}

	var videoOptions map[string]interface{}
	videoOptionsJSON := bodyTxt[startIndex: endIndex]
	json.Unmarshal([]byte(videoOptionsJSON), &videoOptions)

	args := videoOptions["args"]
	videoID := args.(map[string]interface{})["video_id"].(string)
	videoFormats := args.(map[string]interface{})["url_encoded_fmt_stream_map"].(string)
	videoAdaptFormats := args.(map[string]interface{})["adaptive_fmts"].(string)
	videoManifestURL := args.(map[string]interface{})["dashmpd"].(string)
	jsURL := videoOptions["assets"].(map[string]interface{})["js"].(string)

	if strings.Index(jsURL, "//") != -1 {
		jsURL = "http:" + jsURL
	}
	resp, _ = http.Get(jsURL)
	page, _ = ioutil.ReadAll(resp.Body)
	jsScriptBody := string(page[:])

	return YoutubeArgs{
		VideoID: videoID, VideoFormats: videoFormats,
		VideoAdaptFormats: videoAdaptFormats, VideoManifestURL: videoManifestURL,
		ScriptBody: jsScriptBody}
}

func getYoutubeDownloadLink(args YoutubeArgs) string {
	
}

