package main

import (
	"fmt"
	"os"

	"github.com/spf13/pflag"
)

const weatherEnvironmentVariableKey = "OWMAPIKEY"

func main() {
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

	if *location == "" {
		fmt.Print("\033[0;31mA location is required to get the weather\033[0m\n\n")
		pflag.Usage()
		os.Exit(1)
	}

	fmt.Println(*location, *apiKey)
}
