package main

import (
	"fmt"
	"net/http"
	"net/url"
	"io/ioutil"
	"strings"
	"regexp"
	"encoding/json"
	"github.com/robertkrimen/otto"
)

type YoutubeConfigArgs struct {
	VideoID           string
	VideoFormats      string
	VideoAdaptFormats string
	VideoManifestURL  string
	PlayerSource      string
}

type Video struct {
	URL          string
	Signature    string
	Quality      string
	Format       string
}

func MakeDownloadLink(link string, signature string) string {
	downloadLink, _ := url.QueryUnescape(link + "&signature=" + signature)
	fmt.Printf("%T", downloadLink)
	return downloadLink
}

func ExtractVideoFormats(args YoutubeConfigArgs) []Video {
	text := args.VideoFormats + "," + args.VideoAdaptFormats

	videos := make([]Video, 0)
	for _, a := range strings.Split(text, ",") {
		v := make(map[string]string)
		for _, b := range strings.Split(a, "&") {
			pair := strings.Split(b, "=")
			if len(pair) >= 2 {
				v[pair[0]] = pair[1]
			}
		}
		signature := DecryptSignature(args, v["s"])
		format, _ := url.QueryUnescape(v["type"])

		videos = append(videos,
			Video{ URL: MakeDownloadLink(v["url"], signature), Signature: signature,
				Quality: v["quality"], Format: format })
	}
	return videos
}

func GetYoutubeConfigArgs(url string) YoutubeConfigArgs {
	resp, _ := http.Get(url)
	page, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

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

	args := videoOptions["args"].(map[string]interface{})
	videoID := args["video_id"].(string)
	videoFormats := args["url_encoded_fmt_stream_map"].(string)
	videoAdaptFormats := args["adaptive_fmts"].(string)
	videoManifestURL := args["dashmpd"].(string)

	jsURL := videoOptions["assets"].(map[string]interface{})["js"].(string)

	if strings.Index(jsURL, "//") != -1 {
		jsURL = "http:" + jsURL
	}
	resp, _ = http.Get(jsURL)
	page, _ = ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	jsScriptBody := string(page[:])

	return YoutubeConfigArgs{
		VideoID: videoID, VideoFormats: videoFormats,
		VideoAdaptFormats: videoAdaptFormats, VideoManifestURL: videoManifestURL,
		PlayerSource: jsScriptBody}
}

func ExtractSignatureFunc(args YoutubeConfigArgs) string {
	// In the script body, the `foo(bar)` in .set("signature", foo(bar)) is the
	// signature function
	funcNameRegex1 := regexp.MustCompile(`.set\("signature"\s*,\s*(.*)\(`)
	funcNameRegex2 := regexp.MustCompile(`.set\("sig"\s*,\s*(.*)\(`)

	funcName := funcNameRegex1.FindStringSubmatch(args.PlayerSource)[1]
	if len(funcName) == 0 {
		funcName = funcNameRegex2.FindStringSubmatch(args.PlayerSource)[1]
	}

	// We extract the function body
	bodyRegex := regexp.MustCompile("function \\s*" + funcName +
		"\\s*\\([\\w$]*\\)\\s*{[\\w$]*=[\\w$]*\\.split\\(\"\"\\);(.+);return [\\w$]*\\.join")
	body := bodyRegex.FindStringSubmatch(args.PlayerSource)[1]

	return body
}

func ExtractHelperObject(args YoutubeConfigArgs, sigBody string) string {
	// We need to extract whole of the helper object (`fG` in eg above).
	funcNameRegex := regexp.MustCompile(`(\w*).\.*`)
	firstFuncCall := strings.Split(sigBody, ";")[0]
	objName := funcNameRegex.FindStringSubmatch(firstFuncCall)[1]

	objRegex := regexp.MustCompile("var \\s*" + objName);
	startIndex := objRegex.FindStringIndex(args.PlayerSource)[0]

	endIndex := startIndex
	expectSemicolon := false
	for i, c := range args.PlayerSource[startIndex:] {
		if c == '}' {
			expectSemicolon = true
			continue
		}
		if expectSemicolon {
			if c == ';' {
				endIndex = startIndex + i + 1
				break
			}
			expectSemicolon = false
		}
	}
	objCode := args.PlayerSource[startIndex: endIndex]

	return objCode
}

func DecryptSignature(args YoutubeConfigArgs, signature string) string {
	sigBody := ExtractSignatureFunc(args)
	helperObj := ExtractHelperObject(args, sigBody)

	jsvm := otto.New()
	jsvm.Set("a", signature)
	// TODO: Handle error
	_, err := jsvm.Run(helperObj + ";a = a.split(\"\");" + sigBody + ";a = a.join(\"\");")

	fmt.Println(err)

	js_a, _ := jsvm.Get("a")
	value, _ := js_a.ToString()

	return value
}
