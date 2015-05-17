package goutube

// TODO: Check if valid Youtube url. If not (minified urls), try following them

func Youtube(link string, errorChan chan error) <-chan []Link {
	resultChan := make(chan []Link)
	go func() {
		args, err := GetYoutubeConfigArgs(link)
		if err != nil {
			errorChan <- err
			return
		}

		links, err := GetLinks(args)
		if err != nil {
			errorChan <- err
			return
		}
		resultChan <- links
	}()
	return resultChan
}
