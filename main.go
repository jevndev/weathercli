package main

import (
	"errors"
	"fmt"
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

func main() {
	arguments, err := getCommandLineArguments()
	if err != nil {
		// A location is required to get the weather
		fmt.Printf("\033[0;31m%v\033[0m\n\n", err.Error())
		pflag.Usage()
		os.Exit(1)
	}

	fmt.Println(arguments.apiKey, arguments.location)
}
