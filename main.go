package main

import (
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	var sc = SpotifyClient{Token: os.Getenv("SPOTIFY_TOKEN"),
		ApiUrl:     "https://api.spotify.com/v1/",
		HttpClient: &http.Client{},
	}

	var yc = YeelightClient{
		Address: os.Getenv("YEELIGHT_ADDRESS"),
	}

	ch := make(chan PlaybackState)
	go sc.broadcastCurrentTrackAudioAnalysis(ch, &yc)
	sc.checkCurrentState(ch)

}
