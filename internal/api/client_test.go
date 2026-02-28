package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// sampleResponse returns a valid Al Adhan API response for testing.
func sampleResponse() Response {
	return Response{
		Code:   200,
		Status: "OK",
		Data: Data{
			Timings: Timings{
				Fajr:       "05:17",
				Sunrise:    "06:48",
				Dhuhr:      "12:13",
				Asr:        "15:02",
				Sunset:     "17:39",
				Maghrib:    "17:39",
				Isha:       "19:10",
				Imsak:      "05:07",
				Midnight:   "00:14",
				Firstthird: "22:02",
				Lastthird:  "02:25",
			},
			Date: DateInfo{
				Readable:  "28 Feb 2026",
				Timestamp: "1772262000",
			},
			Meta: Meta{
				Latitude:  51.5074,
				Longitude: -0.1278,
				Timezone:  "Europe/London",
				Method:    MethodInfo{ID: 2, Name: "ISNA"},
				School:    "STANDARD",
			},
		},
	}
}

func TestNewClient(t *testing.T) {
	c := NewClient()
	if c == nil {
		t.Fatal("NewClient returned nil")
	}
	if c.BaseURL != defaultBaseURL {
		t.Errorf("BaseURL = %q, want %q", c.BaseURL, defaultBaseURL)
	}
}

func TestFetchByCoordinates_Success(t *testing.T) {
	resp := sampleResponse()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request path contains /timings/ and date format DD-MM-YYYY.
		if !strings.Contains(r.URL.Path, "/timings/28-02-2026") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		// Verify query params.
		q := r.URL.Query()
		if q.Get("latitude") == "" {
			t.Error("missing latitude param")
		}
		if q.Get("longitude") == "" {
			t.Error("missing longitude param")
		}
		if q.Get("method") != "2" {
			t.Errorf("method = %q, want %q", q.Get("method"), "2")
		}
		if q.Get("school") != "1" {
			t.Errorf("school = %q, want %q", q.Get("school"), "1")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := NewClient()
	c.BaseURL = server.URL

	date := time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC)
	got, err := c.FetchByCoordinates(date, 51.5074, -0.1278, 2, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Data.Timings.Fajr != "05:17" {
		t.Errorf("Fajr = %q, want %q", got.Data.Timings.Fajr, "05:17")
	}
	if got.Data.Meta.Timezone != "Europe/London" {
		t.Errorf("Timezone = %q, want %q", got.Data.Meta.Timezone, "Europe/London")
	}
}

func TestFetchByCoordinates_NoMethodOrSchool(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		// method=-1 and school=-1 should not be sent.
		if q.Get("method") != "" {
			t.Errorf("method should not be set, got %q", q.Get("method"))
		}
		if q.Get("school") != "" {
			t.Errorf("school should not be set, got %q", q.Get("school"))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(sampleResponse())
	}))
	defer server.Close()

	c := NewClient()
	c.BaseURL = server.URL

	date := time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC)
	_, err := c.FetchByCoordinates(date, 51.5074, -0.1278, -1, -1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFetchByCity_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/timingsByCity/") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("city") != "London" {
			t.Errorf("city = %q, want %q", q.Get("city"), "London")
		}
		if q.Get("country") != "UK" {
			t.Errorf("country = %q, want %q", q.Get("country"), "UK")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(sampleResponse())
	}))
	defer server.Close()

	c := NewClient()
	c.BaseURL = server.URL

	date := time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC)
	got, err := c.FetchByCity(date, "London", "UK", -1, -1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Data.Timings.Asr != "15:02" {
		t.Errorf("Asr = %q, want %q", got.Data.Timings.Asr, "15:02")
	}
}

func TestFetchByCoordinates_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	c := NewClient()
	c.BaseURL = server.URL

	date := time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC)
	_, err := c.FetchByCoordinates(date, 51.5, -0.1, -1, -1)
	if err == nil {
		t.Fatal("expected error for HTTP 503, got nil")
	}
	if !strings.Contains(err.Error(), "503") {
		t.Errorf("error should mention 503, got: %v", err)
	}
}

func TestFetchByCoordinates_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	c := NewClient()
	c.BaseURL = server.URL

	date := time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC)
	_, err := c.FetchByCoordinates(date, 51.5, -0.1, -1, -1)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
	if !strings.Contains(err.Error(), "decode") {
		t.Errorf("error should mention decode, got: %v", err)
	}
}

func TestFetchByCoordinates_APIErrorCode(t *testing.T) {
	resp := Response{Code: 400, Status: "Bad Request"}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := NewClient()
	c.BaseURL = server.URL

	date := time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC)
	_, err := c.FetchByCoordinates(date, 51.5, -0.1, -1, -1)
	if err == nil {
		t.Fatal("expected error for API code 400, got nil")
	}
	if !strings.Contains(err.Error(), "400") {
		t.Errorf("error should mention 400, got: %v", err)
	}
}

func TestFetchByCoordinates_ConnectionRefused(t *testing.T) {
	c := NewClient()
	c.BaseURL = "http://127.0.0.1:1" // nothing listening

	date := time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC)
	_, err := c.FetchByCoordinates(date, 51.5, -0.1, -1, -1)
	if err == nil {
		t.Fatal("expected error for connection refused, got nil")
	}
}

func TestFetchByCoordinates_DateFormat(t *testing.T) {
	var capturedPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(sampleResponse())
	}))
	defer server.Close()

	c := NewClient()
	c.BaseURL = server.URL

	// Test that the date is formatted as DD-MM-YYYY.
	date := time.Date(2026, 3, 5, 0, 0, 0, 0, time.UTC)
	_, err := c.FetchByCoordinates(date, 0, 0, -1, -1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(capturedPath, "/timings/05-03-2026") {
		t.Errorf("date format wrong in path: %s (expected DD-MM-YYYY)", capturedPath)
	}
}
