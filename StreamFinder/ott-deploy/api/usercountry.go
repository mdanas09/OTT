package api

import (
	"encoding/json"
	"github.com/cevatbarisyilmaz/ip2country"
	"log"
	"net"
	"net/http"
	"strings"
)

type UserCountryResponse struct {
	Country  string `json:"country,omitempty"`
	Detected bool   `json:"detected"`
}

func UserCountry(writer http.ResponseWriter, request *http.Request) {
	var country string

	// Render and other reverse proxies may place the original client IP in
	// X-Forwarded-For. Use it when present and fall back to RemoteAddr locally.
	host := request.Header.Get("X-Forwarded-For")
	if host == "" {
		host, _, _ = net.SplitHostPort(request.RemoteAddr)
	} else if idx := strings.IndexByte(host, ','); idx >= 0 {
		host = strings.TrimSpace(host[:idx])
	}

	if host != "" {
		country, _ = ip2country.Country(net.ParseIP(host))
		country = strings.ToLower(country)
	}
	data := &UserCountryResponse{Country: country, Detected: country != ""}
	response, err := json.Marshal(data)
	if err != nil {
		log.Println(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	writer.Write(response)
}
