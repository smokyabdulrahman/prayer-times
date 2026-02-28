package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const defaultBaseURL = "https://api.aladhan.com/v1"

// Client communicates with the Al Adhan prayer times API.
type Client struct {
	httpClient *http.Client
	// BaseURL is the API base URL. Defaults to the Al Adhan API.
	// Exported for testing with httptest.
	BaseURL string
}

// NewClient creates a new API client with sensible defaults.
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		BaseURL: defaultBaseURL,
	}
}

// FetchByCoordinates fetches prayer times for the given date and coordinates.
func (c *Client) FetchByCoordinates(date time.Time, lat, lon float64, method, school int) (*Response, error) {
	dateStr := date.Format("02-01-2006")
	endpoint := fmt.Sprintf("%s/timings/%s", c.BaseURL, dateStr)

	params := url.Values{}
	params.Set("latitude", fmt.Sprintf("%f", lat))
	params.Set("longitude", fmt.Sprintf("%f", lon))
	if method >= 0 {
		params.Set("method", fmt.Sprintf("%d", method))
	}
	if school >= 0 {
		params.Set("school", fmt.Sprintf("%d", school))
	}

	return c.doRequest(endpoint, params)
}

// FetchByCity fetches prayer times for the given date, city, and country.
func (c *Client) FetchByCity(date time.Time, city, country string, method, school int) (*Response, error) {
	dateStr := date.Format("02-01-2006")
	endpoint := fmt.Sprintf("%s/timingsByCity/%s", c.BaseURL, dateStr)

	params := url.Values{}
	params.Set("city", city)
	params.Set("country", country)
	if method >= 0 {
		params.Set("method", fmt.Sprintf("%d", method))
	}
	if school >= 0 {
		params.Set("school", fmt.Sprintf("%d", school))
	}

	return c.doRequest(endpoint, params)
}

func (c *Client) doRequest(endpoint string, params url.Values) (*Response, error) {
	reqURL := fmt.Sprintf("%s?%s", endpoint, params.Encode())

	resp, err := c.httpClient.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var apiResp Response
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode API response: %w", err)
	}

	if apiResp.Code != 200 {
		return nil, fmt.Errorf("API error: code=%d status=%s", apiResp.Code, apiResp.Status)
	}

	return &apiResp, nil
}
