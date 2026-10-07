# Deploy on Render

This folder is the standalone Go Movie Recommender web app from the Streaming Availability API example.

## Required secrets

- `STREAMING_AVAILABILITY_API_KEY`: RapidAPI subscription key for Streaming Availability API.
- `TMDB_API_KEY`: TMDB API key.

Do not commit these keys to GitHub.

## Local run

```bash
go mod download
go run .
```

Open `http://localhost:10000`.

## Render

1. Push the contents of this folder to a new GitHub repository.
2. In Render, choose **New -> Web Service** and connect that repository.
3. Use the Go runtime.
4. Build command: `go build -tags netgo -ldflags '-s -w' -o app`
5. Start command: `./app`
6. Add `STREAMING_AVAILABILITY_API_KEY` and `TMDB_API_KEY` in Render Environment Variables.
7. Deploy.
8. Test `/healthz` and then open the service URL.

`render.yaml` contains the same service configuration if you deploy via Render Blueprint.
