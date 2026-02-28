package geo

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDetectLocation_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := ipAPIResponse{
			Status:   "success",
			Lat:      51.5074,
			Lon:      -0.1278,
			City:     "London",
			Country:  "United Kingdom",
			Timezone: "Europe/London",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Override the package-level URL for testing.
	origURL := geoAPIURL
	geoAPIURL = server.URL
	defer func() { geoAPIURL = origURL }()

	loc, err := DetectLocation()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if loc.Latitude != 51.5074 {
		t.Errorf("Latitude = %v, want %v", loc.Latitude, 51.5074)
	}
	if loc.Longitude != -0.1278 {
		t.Errorf("Longitude = %v, want %v", loc.Longitude, -0.1278)
	}
	if loc.City != "London" {
		t.Errorf("City = %q, want %q", loc.City, "London")
	}
	if loc.Country != "United Kingdom" {
		t.Errorf("Country = %q, want %q", loc.Country, "United Kingdom")
	}
	if loc.Timezone != "Europe/London" {
		t.Errorf("Timezone = %q, want %q", loc.Timezone, "Europe/London")
	}
}

func TestDetectLocation_APIFailureStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := ipAPIResponse{
			Status:  "fail",
			Message: "reserved range",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	origURL := geoAPIURL
	geoAPIURL = server.URL
	defer func() { geoAPIURL = origURL }()

	_, err := DetectLocation()
	if err == nil {
		t.Fatal("expected error for failed status, got nil")
	}
	if !strings.Contains(err.Error(), "reserved range") {
		t.Errorf("error should contain message, got: %v", err)
	}
}

func TestDetectLocation_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer server.Close()

	origURL := geoAPIURL
	geoAPIURL = server.URL
	defer func() { geoAPIURL = origURL }()

	_, err := DetectLocation()
	if err == nil {
		t.Fatal("expected error for HTTP 500, got nil")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("error should mention 500, got: %v", err)
	}
}

func TestDetectLocation_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("not json at all"))
	}))
	defer server.Close()

	origURL := geoAPIURL
	geoAPIURL = server.URL
	defer func() { geoAPIURL = origURL }()

	_, err := DetectLocation()
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
	if !strings.Contains(err.Error(), "decode") {
		t.Errorf("error should mention decode, got: %v", err)
	}
}

func TestDetectLocation_ConnectionRefused(t *testing.T) {
	origURL := geoAPIURL
	geoAPIURL = "http://127.0.0.1:1" // nothing listening
	defer func() { geoAPIURL = origURL }()

	_, err := DetectLocation()
	if err == nil {
		t.Fatal("expected error for connection refused, got nil")
	}
}
