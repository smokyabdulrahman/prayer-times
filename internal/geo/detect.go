package geo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Location holds geographic coordinates detected from the user's IP.
type Location struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lon"`
	City      string  `json:"city"`
	Country   string  `json:"country"`
	Timezone  string  `json:"timezone"`
}

// ipAPIResponse maps the response from ip-api.com.
type ipAPIResponse struct {
	Status   string  `json:"status"`
	Message  string  `json:"message"`
	Lat      float64 `json:"lat"`
	Lon      float64 `json:"lon"`
	City     string  `json:"city"`
	Country  string  `json:"country"`
	Timezone string  `json:"timezone"`
}

// geoAPIURL is the geolocation API endpoint. It is a variable (not a constant)
// so that tests can override it with an httptest server URL.
var geoAPIURL = "http://ip-api.com/json/?fields=status,message,lat,lon,city,country,timezone"

// DetectLocation uses ip-api.com to determine the user's location from their
// public IP address. This is a free service that requires no API key.
func DetectLocation() (*Location, error) {
	client := &http.Client{Timeout: 5 * time.Second}

	resp, err := client.Get(geoAPIURL)
	if err != nil {
		return nil, fmt.Errorf("geolocation request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("geolocation API returned status %d", resp.StatusCode)
	}

	var result ipAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode geolocation response: %w", err)
	}

	if result.Status != "success" {
		return nil, fmt.Errorf("geolocation failed: %s", result.Message)
	}

	return &Location{
		Latitude:  result.Lat,
		Longitude: result.Lon,
		City:      result.City,
		Country:   result.Country,
		Timezone:  result.Timezone,
	}, nil
}
