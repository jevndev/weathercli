package main

import "testing"

func TestValidateArgumentsShouldErrorWithNoLocation(t *testing.T) {
	const arbitraryNonEmptyAPIKey = "keymckeyface"
	err := validateArguments(programArguments{
		location: "", apiKey: arbitraryNonEmptyAPIKey,
	})
	if err == nil {
		t.Fail()
	}
}

func TestValidateArgumentsShouldErrorWithNoApiKey(t *testing.T) {
	const arbitraryNonEmptyLocation = "Nowhere"
	err := validateArguments(programArguments{
		location: arbitraryNonEmptyLocation, apiKey: "",
	})
	if err == nil {
		t.Fail()
	}
}
