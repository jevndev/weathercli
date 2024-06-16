package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/pflag"
)

const weatherEnvironmentVariableKey = "OWMAPIKEY"

type programArguments struct {
	apiKey   string
	location string
}

func validateArguments(arguments programArguments) (err error) {
	if arguments.location == "" {
		return errors.New("a location is required to get the weather")
	}
	if arguments.apiKey == "" {
		return errors.New("an openstreetmap api key is required to get the weather")
	}
	return nil
}

func getCommandLineArguments() (arguments programArguments, err error) {
	location := pflag.String(
		"location",
		"",
		"The location to get the weather from",
	)
	apiKey := pflag.StringP(
		"api-key",
		"k",
		os.Getenv(weatherEnvironmentVariableKey),
		fmt.Sprintf(
			"Your API key for OpenWeatherMap. If not provided, defaults to the \"%v\" environment variable",
			weatherEnvironmentVariableKey,
		),
	)

	pflag.Parse()

	parsedArguments := programArguments{*apiKey, *location}

	if err := validateArguments(parsedArguments); err != nil {
		return programArguments{}, err
	}

	return parsedArguments, nil
}

func formatGeoCodeRequest(locationName string, apiKey string) (geocodeRequestUrl string) {
	return fmt.Sprintf("http://api.openweathermap.org/geo/1.0/direct?q=%s&limit=1&appid=%s", locationName, apiKey)
}

type LatLon struct {
	Latitude  float64
	Longitude float64
}

type GeoCodeResponseItem struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lon"`
}

func requestLatLonFromUrl(geocodeRequestUrl string) (latlon LatLon, err error) {
	apiResponse, err := http.Get(geocodeRequestUrl)
	if err != nil {
		return LatLon{}, err
	}
	defer apiResponse.Body.Close()

	var responseItems []GeoCodeResponseItem

	jsonDecodeErr := json.NewDecoder(apiResponse.Body).Decode(&responseItems)

	if jsonDecodeErr != nil {
		return LatLon{}, jsonDecodeErr
	}

	if len(responseItems) != 1 {
		return LatLon{}, errors.New("got an unexpected number of returned locations from openstreetmap")
	}

	return LatLon{}, nil
}

func requestLatLonOfLocation(location string, apikey string) (latlon LatLon, err error) {
	requestUrl := formatGeoCodeRequest(location, apikey)
	return requestLatLonFromUrl(requestUrl)
}
func main() {
	arguments, err := getCommandLineArguments()
	if err != nil {
		// A location is required to get the weather
		fmt.Printf("\033[0;31m%v\033[0m\n\n", err.Error())
		pflag.Usage()
		os.Exit(1)
	}

	locationLatLon, latLonRequestError := requestLatLonOfLocation(arguments.location, arguments.apiKey)

	if latLonRequestError != nil {
		fmt.Printf("\033[0;31m%v\033[0m\n\n", err.Error())
		os.Exit(1)
	}
}
