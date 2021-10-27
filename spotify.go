package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type SpotifyClient struct {
	Token      string
	ApiUrl     string
	HttpClient *http.Client
}

type PlaybackState struct {
	ProgressMs int  `json:"progress_ms"`
	IsPlaying  bool `json:"is_playing"`
	Item       struct {
		Id       string `json:"id"`
		Duration int    `json:"duration"`
	} `json:"item"`
}

type TrackAudioAnalysis struct {
	Bars []Bar `json:"bars"`
}
type Bar struct {
	Start      float32 `json:"start"`
	Duration   float32 `json:"duration"`
	Confidence float32 `json:"confidence"`
}

func (sc *SpotifyClient) getPlaybackState() PlaybackState {
	var currentTrackUrl = "me/player"
	request, err := sc.generateRequest(currentTrackUrl)
	if err != nil {
		return PlaybackState{}
	}
	body, err := sc.doRequest(request)
	if err != nil {
		return PlaybackState{}
	}
	playbackState := PlaybackState{}
	err = json.Unmarshal(body, &playbackState)
	if err != nil {
		return PlaybackState{}
	}
	return playbackState
}

func (sc *SpotifyClient) getTrackAudioAnalysis(id string) TrackAudioAnalysis {
	var currentTrackUrl = fmt.Sprintf("audio-analysis/%s", id)
	request, err := sc.generateRequest(currentTrackUrl)
	if err != nil {
		return TrackAudioAnalysis{}
	}
	body, err := sc.doRequest(request)
	if err != nil {
		return TrackAudioAnalysis{}
	}
	analysis := TrackAudioAnalysis{}
	err = json.Unmarshal(body, &analysis)
	if err != nil {
		return TrackAudioAnalysis{}
	}
	return analysis
}

func (sc *SpotifyClient) generateRequest(url string) (*http.Request, error) {
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s%s", sc.ApiUrl, url), nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", sc.Token))
	return request, nil
}

func (sc *SpotifyClient) doRequest(request *http.Request) ([]byte, error) {
	response, err := sc.HttpClient.Do(request)
	if err != nil {
		return nil, err
	}
	if response.Body != nil {
		defer response.Body.Close()
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (sc *SpotifyClient) checkCurrentState(chOut chan<- PlaybackState) {
	ticker := time.NewTicker(1 * time.Second)
	for range ticker.C {
		state := sc.getPlaybackState()
		chOut <- state
	}
}

func (sc *SpotifyClient) broadcastCurrentTrackAudioAnalysis(chIn <-chan PlaybackState, yc *YeelightClient) {
	trackId := ""
	var trackAudioAnalysis TrackAudioAnalysis = TrackAudioAnalysis{}
	progressMs := 0
	isPlaying := false

	go func() {
		for {
			_trackId := trackId
			_progressMs := progressMs
			for _, val := range trackAudioAnalysis.Bars {
				if !isPlaying || _trackId != trackId || progressMs < _progressMs {
					break
				}

				_progressMs = progressMs
				if val.Start*1000 < float32(progressMs) {
					continue
				}

				go yc.convertAndSendRgbMessage(val.Confidence)
				time.Sleep(time.Duration(val.Duration*1000) * time.Millisecond)
			}
		}
	}()

	for msg := range chIn {
		if trackId != msg.Item.Id {
			trackAudioAnalysis = sc.getTrackAudioAnalysis(msg.Item.Id)
			progressMs = msg.ProgressMs
			trackId = msg.Item.Id
			continue
		}
		if !msg.IsPlaying {
			progressMs = msg.ProgressMs
			isPlaying = false
			continue
		}
		if msg.IsPlaying {
			progressMs = msg.ProgressMs
			isPlaying = true
			continue
		}
	}

}
