package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"slices"

	"github.com/spf13/pflag"
)

const weatherEnvironmentVariableKey = "OWMAPIKEY"

var availableUnits = [...]string{
	"imperial",
	"standard",
	"metric",
}

var unitToTemperatureSuffix = map[string]string{
	"imperial": "F",
	"standard": "K",
	"metric":   "C",
}

var DescriptionToEmojiMap = map[string]string{
	"Clear":        "☀️",
	"Clouds":       "🌤️",
	"Drizzle":      "🌦️",
	"Rain":         "🌧️",
	"Thunderstorm": "⛈️",
	"Snow":         "🌨️",
	"Mist":         "🌫️",
}

type programArguments struct {
	apiKey      string
	location    string
	units       string
	enableEmoji bool
}

func validateArguments(arguments programArguments) (err error) {
	if arguments.location == "" {
		return errors.New("a location is required to get the weather")
	}
	if arguments.apiKey == "" {
		return errors.New("an openstreetmap api key is required to get the weather")
	}

	if !slices.Contains(availableUnits[:], arguments.units) {
		return errors.New("provided units are invalid")
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
	units := pflag.StringP(
		"units",
		"u",
		availableUnits[0],
		fmt.Sprintf("The units to use when formatting the output. One of %v", availableUnits),
	)

	enableEmoji := pflag.BoolP(
		"enable-emoji",
		"e",
		false,
		"Whether to enable emojis representing the weather condition",
	)

	pflag.Parse()

	parsedArguments := programArguments{*apiKey, *location, *units, *enableEmoji}

	if err := validateArguments(parsedArguments); err != nil {
		return programArguments{}, err
	}

	return parsedArguments, nil
}

func formatGeoCodeRequest(locationName string, apiKey string) (geocodeRequestUrl string) {
	escapedLocationName := url.QueryEscape(locationName)
	return fmt.Sprintf("http://api.openweathermap.org/geo/1.0/direct?q=%s&limit=1&appid=%s", escapedLocationName, apiKey)
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

	mostRelevantLatLonResponse := LatLon(responseItems[0])

	return mostRelevantLatLonResponse, nil
}

func requestLatLonOfLocation(location string, apikey string) (latlon LatLon, err error) {
	requestUrl := formatGeoCodeRequest(location, apikey)
	return requestLatLonFromUrl(requestUrl)
}

func formatWeatherRequest(latlon LatLon, apikey string, units string) (requestUrl string) {
	return fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?lat=%f&lon=%f&appid=%s&units=%s", latlon.Latitude, latlon.Longitude, apikey, units)
}

type WeatherResponseMain struct {
	Temp float32 `json:"temp"`
}

type WeatherResponseWeather struct {
	Main        string `json:"main"`
	Description string `json:"description"`
}

type WeatherResponse struct {
	MainResponse    WeatherResponseMain      `json:"main"`
	WeatherResponse []WeatherResponseWeather `json:"weather"`
}

type Weather struct {
	temperature float32
	condition   string
	description string
}

func requestWeatherFromUrl(weatherRequestUrl string) (weather Weather, err error) {
	apiResponse, err := http.Get(weatherRequestUrl)
	if err != nil {
		return Weather{}, err
	}

	defer apiResponse.Body.Close()

	var response WeatherResponse

	jsonDecodeErr := json.NewDecoder(apiResponse.Body).Decode(&response)

	if jsonDecodeErr != nil {
		return Weather{}, err
	}

	if len(response.WeatherResponse) == 0 {
		return Weather{}, errors.New("OpenWeatherMap didn't return any weather responses")
	}

	mostRelevantWeatherResponse := response.WeatherResponse[0]

	return Weather{
		response.MainResponse.Temp,
		mostRelevantWeatherResponse.Main,
		mostRelevantWeatherResponse.Description,
	}, nil
}

func requestWeatherForLatLon(latlon LatLon, apikey string, weatherUnits string) (weather Weather, err error) {
	weatherRequestUrl := formatWeatherRequest(latlon, apikey, weatherUnits)
	return requestWeatherFromUrl(weatherRequestUrl)
}

func formatWeather(weather Weather, temperatureSuffix string) (formattedWeather string) {
	return fmt.Sprintf("%s, %.2f°%s", weather.description, weather.temperature, temperatureSuffix)
}

func formatWeatherWithEmoji(weather Weather, temperatureSuffix string, emoji string) (formattedWeather string) {
	return fmt.Sprintf("%s %s, %.2f°%s", emoji, weather.description, weather.temperature, temperatureSuffix)
}

func makeFormattedWeatherString(weather Weather, temperatureSuffix string, enableEmoji bool) (formattedWeather string) {
	if enableEmoji {
		conditionEmoji, conditionInMap := DescriptionToEmojiMap[weather.condition]
		if !conditionInMap {
			conditionEmoji = "🛰️"
		}
		return formatWeatherWithEmoji(weather, temperatureSuffix, conditionEmoji)
	}
	return formatWeather(weather, temperatureSuffix)
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
		fmt.Printf("%v\033[0m\n\n", latLonRequestError.Error())
		os.Exit(1)
	}

	weather, weatherRequestError := requestWeatherForLatLon(
		locationLatLon, arguments.apiKey, arguments.units,
	)

	if weatherRequestError != nil {
		fmt.Printf("\033[0;31m%v\033[0m\n\n", weatherRequestError.Error())
		os.Exit(1)
	}

	temperatureSuffix, temperatureSuffixValid := unitToTemperatureSuffix[arguments.units]

	if !temperatureSuffixValid {
		fmt.Printf("\033[0;31mGot an unexpected unit specification which has no associated temperature suffix\033[0m\n\n")
		os.Exit(1)
	}

	formattedWeatherString := makeFormattedWeatherString(weather, temperatureSuffix, arguments.enableEmoji)

	fmt.Println(formattedWeatherString)
}
