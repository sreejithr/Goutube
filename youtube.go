package main

import (
	"errors"
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

func MakeDownloadLink(link string, signature string) string {
	downloadLink, _ := url.QueryUnescape(link + "&signature=" + signature);
	return downloadLink
}

func getAVCodecs(s string, t ResultType) (string, string, error) {
	// Returns audio and video codecs in that order.
	// s should look like codec="<video-codec>, <audio-codec>".
	split := strings.Split(s, "=")
	if len(split) < 2 {
		return "", "", errors.New("Invalid codec string supplied")
	}
	codecs := strings.Split(split[1], ",")

	if t == AUDIO {
		audio := codecs[0]
		return audio, "", nil
	}

	if len(codecs) < 2 {
		video := codecs[0]
		return "", video, nil
	}

	audio := codecs[1]
	video := codecs[0]
	return audio, video, nil
}

func extractResultFormat(s string, t ResultType) (ResultFormat, error){
	split := strings.Split(s, ";")
	format := split[0]
	rf := ResultFormat{format, "", "", s}

	if len(split) > 1 {
		a, v, err := getAVCodecs(split[1], t)
		if err != nil {
			return rf, err
		}
		rf.AudioCodec = a
		rf.VideoCodec = v
	}

	return rf, nil
}

func GetResults(args YoutubeConfigArgs) ([]Result, error) {
	text := args.VideoFormats + "," + args.VideoAdaptFormats

	videos := make([]Result, 0)
	for _, a := range strings.Split(text, ",") {
		v := make(map[string]string)
		for _, b := range strings.Split(a, "&") {
			pair := strings.Split(b, "=")
			if len(pair) >= 2 {
				v[pair[0]] = pair[1]
			}
		}
		signature, err := DecryptSignature(args, v["s"])
		if err != nil {
			return nil, err
		}

		format, err := url.QueryUnescape(v["type"])
		if err != nil {
			return nil, err
		}

		var resultType ResultType
		if strings.Index(format, "audio") > -1 {
			resultType = AUDIO
		} else {
			resultType = VIDEO
		}

		resultFormat, err := extractResultFormat(format, resultType)
		if err != nil {
			return make([]Result, 0), err
		}

		videos = append(videos,
			Result{ URL: MakeDownloadLink(v["url"], signature),
				Signature: signature, Quality: v["quality"],
				Type: resultType, Format: resultFormat })
	}
	return videos, nil
}

func GetYoutubeConfigArgs(link string) (YoutubeConfigArgs, error) {
	resp, err := http.Get(link)
	if err != nil {
		return YoutubeConfigArgs{}, err
	}

	page, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return YoutubeConfigArgs{}, err
	}
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

	resp, err = http.Get(jsURL)
	if err != nil {
		return YoutubeConfigArgs{}, errors.New("Failed to get PlayerSource")
	}

	page, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return YoutubeConfigArgs{}, errors.New(
			"Failed to read remote response (PlayerSource)")
	}

	defer resp.Body.Close()
	jsScriptBody := string(page[:])

	return YoutubeConfigArgs{
		VideoID: videoID, VideoFormats: videoFormats,
		VideoAdaptFormats: videoAdaptFormats, VideoManifestURL: videoManifestURL,
		PlayerSource: jsScriptBody}, nil
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

func DecryptSignature(args YoutubeConfigArgs, signature string) (string, error) {
	sigBody := ExtractSignatureFunc(args)
	helperObj := ExtractHelperObject(args, sigBody)

	jsvm := otto.New()
	jsvm.Set("a", signature)

	_, err := jsvm.Run(helperObj + ";a = a.split(\"\");" + sigBody + ";a = a.join(\"\");")

	if err != nil {
		return "", errors.New("Error in JS runtime")
	}

	js_a, err := jsvm.Get("a")
	if err != nil {
		return "", errors.New("Error decrypting signature")
	}

	value, err := js_a.ToString()
	if err != nil {
		return "", errors.New("Error decrypting signature")
	}

	return value, err
}
