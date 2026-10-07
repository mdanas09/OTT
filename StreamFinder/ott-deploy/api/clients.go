package api

import (
	"os"

	"github.com/cyruzin/golang-tmdb"
	"github.com/joho/godotenv"
	streaming "github.com/movieofthenight/go-streaming-availability/v4"
)

var streamingAvailabilityClient = createStreamingAvailabilityClient()
var tmdbClient = createTmdbClient()

func createStreamingAvailabilityClient() *streaming.APIClient {
	_ = godotenv.Load()
	apiKey := os.Getenv("STREAMING_AVAILABILITY_API_KEY")
	if apiKey == "" {
		panic("STREAMING_AVAILABILITY_API_KEY is not set")
	}

	return streaming.NewAPIClientFromApiKey(apiKey, nil)
}

func createTmdbClient() *tmdb.Client {
	apiKey := os.Getenv("TMDB_API_KEY")
	client, err := tmdb.Init(apiKey)
	if err != nil {
		panic(err)
	}
	return client
}