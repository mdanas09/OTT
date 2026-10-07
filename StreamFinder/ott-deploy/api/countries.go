package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
)

type Country struct {
	Code     string             `json:"code"`
	Name     string             `json:"name"`
	Services map[string]Service `json:"services"`
}

type Service struct {
	Id            string `json:"id"`
	Name          string `json:"name"`
	DarkThemeLogo string `json:"darkThemeLogo"`
	WhiteLogo     string `json:"whiteLogo"`
	ThemeColor    string `json:"themeColor"`
}

var countries = map[string]Country{}

func init() {
	response, _, err := streamingAvailabilityClient.CountriesAPI.
		GetCountries(context.Background()).
		Execute()

	if err != nil {
		panic(err)
	}

	for countryCode, country := range *response {
		services := map[string]Service{}

		for _, service := range country.Services {
			if service.StreamingOptionTypes.Subscription {
				services[service.Id] = Service{
					Id:            service.Id,
					Name:          service.Name,
					DarkThemeLogo: service.ImageSet.DarkThemeImage,
					WhiteLogo:     service.ImageSet.WhiteImage,
					ThemeColor:    service.ThemeColorCode,
				}
			}
		}

		countries[countryCode] = Country{
			Code:     country.CountryCode,
			Name:     country.Name,
			Services: services,
		}
	}
}

func Countries(writer http.ResponseWriter, request *http.Request) {
	response, err := json.Marshal(countries)
	if err != nil {
		log.Println(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.Write(response)
}