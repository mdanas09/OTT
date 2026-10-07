package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	streaming "github.com/movieofthenight/go-streaming-availability/v4"
)

type Movie struct {
	Title         string `json:"title"`
	Poster        string `json:"poster"`
	StreamingLink string `json:"streamingLink"`
	StreamingLogo string `json:"streamingLogo"`
}

// Recommend recommends movies based on the query parameters.
func Recommend(writer http.ResponseWriter, request *http.Request) {
	searchRequest, country, services := consumeQueryParameters(
		request.URL.Query(),
	)

	searchResponse, _, err := searchRequest.Execute()
	if err != nil {
		log.Println(err)
		http.Error(writer, "Failed to fetch movie recommendations", http.StatusInternalServerError)
		return
	}

	movies := readSearchResponse(searchResponse, country, services)

	response, err := json.Marshal(movies)
	if err != nil {
		log.Println(err)
		http.Error(writer, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.Write(response)
}

// consumeQueryParameters consumes the query parameters and returns the search
// request, the country code and the selected services.
func consumeQueryParameters(
	query url.Values,
) (
	searchRequest streaming.ApiSearchShowsByFiltersRequest,
	country string,
	services map[string]bool,
) {
	country = query.Get("country")

	var catalogs []string

	services = map[string]bool{}

	for i, service := range strings.Split(query.Get("services"), ",") {
		if service == "" {
			continue
		}

		// Limit the number of services to 16.
		if i >= 16 {
			break
		}

		services[service] = true

		// Only include subscription availability.
		catalogs = append(catalogs, fmt.Sprintf("%s.subscription", service))
	}

	// Create a new v4 search request.
	searchRequest = streamingAvailabilityClient.
		ShowsAPI.
		SearchShowsByFilters(context.Background())

	// We only want movies.
	searchRequest = searchRequest.ShowType(streaming.MOVIE)

	// Country in which the availability should be checked.
	searchRequest = searchRequest.Country(country)

	// Search only selected streaming catalogs.
	if len(catalogs) > 0 {
		searchRequest = searchRequest.Catalogs(catalogs)
	}

	// Optional genre filter.
	if query.Has("genre") && query.Get("genre") != "" {
		searchRequest = searchRequest.Genres(
			[]string{query.Get("genre")},
		)
	}

	// Optional keyword filter.
	if query.Has("keyword") && query.Get("keyword") != "" {
		searchRequest = searchRequest.Keyword(
			query.Get("keyword"),
		)
	}

	// Movie type / sorting.
	switch query.Get("movieType") {

	case TrendingNow:
		searchRequest = searchRequest.
			OrderBy("popularity_1week").
			OrderDirection(streaming.DESC)

	case AllTimeClassics:
		searchRequest = searchRequest.
			OrderBy("popularity_alltime").
			OrderDirection(streaming.DESC)

	case OldiesButGoldies:
		searchRequest = searchRequest.
			YearMax(int32(time.Now().Year() - 25)).
			OrderBy("popularity_alltime").
			OrderDirection(streaming.DESC)

	case BestOfRecentYears:
		searchRequest = searchRequest.
			OrderBy("popularity_1year").
			OrderDirection(streaming.DESC)
	}

	return
}

// readSearchResponse reads the search response and returns the movies found.
func readSearchResponse(
	searchResponse *streaming.SearchResult,
	country string,
	services map[string]bool,
) (movies []Movie) {
	movies = make([]Movie, 0)

	if searchResponse == nil {
		return
	}

	// Fetch poster URLs from TMDB in parallel.
	wg := sync.WaitGroup{}
	posters := map[int]string{}
	postersMu := &sync.Mutex{}

	for _, movie := range searchResponse.Shows {
		movie := movie

		wg.Add(1)

		go func() {
			defer wg.Done()

			tmdbID, err := parseTMDBMovieID(movie.TmdbId)
			if err != nil {
				log.Printf(
					"Invalid TMDB ID %q for %q: %v",
					movie.TmdbId,
					movie.Title,
					err,
				)
				return
			}

			poster, err := getPosterUrl(tmdbID)
			if err != nil {
				log.Printf("Failed to get poster for %q", movie.Title)
				return
			}

			postersMu.Lock()
			posters[tmdbID] = poster
			postersMu.Unlock()
		}()
	}

	wg.Wait()

	// Build the final response.
	for _, movie := range searchResponse.Shows {

		tmdbID, err := parseTMDBMovieID(movie.TmdbId)
		if err != nil {
			continue
		}

		poster := posters[tmdbID]

		// Only show movies with posters.
		if poster == "" {
			continue
		}

		// Availability for the selected country.
		streamingOptions := movie.StreamingOptions[country]

		for _, streamingOption := range streamingOptions {

			serviceID := streamingOption.Service.Id

			// Only use services selected by the user.
			if !services[serviceID] {
				continue
			}

			// Only use subscription availability.
			if streamingOption.Type != streaming.SUBSCRIPTION {
				continue
			}

			service, ok := countries[country].Services[serviceID]
			if !ok {
				continue
			}

			movies = append(movies, Movie{
				Title:         movie.Title,
				Poster:        poster,
				StreamingLink: streamingOption.Link,
				StreamingLogo: service.DarkThemeLogo,
			})

			break
		}
	}

	return
}

func parseTMDBMovieID(value string) (int, error) {
	return strconv.Atoi(strings.TrimPrefix(value, "movie/"))
}

// getPosterUrl fetches the poster URL from TMDB by TMDB ID.
// Returns an empty string if the movie does not have a poster.
func getPosterUrl(tmdbID int) (string, error) {
	tmdbInfo, err := tmdbClient.GetMovieDetails(tmdbID, nil)
	if err != nil {
		return "", err
	}

	if len(tmdbInfo.PosterPath) > 0 {
		return fmt.Sprintf(
			"https://image.tmdb.org/t/p/w500/%s",
			tmdbInfo.PosterPath,
		), nil
	}

	return "", nil
}