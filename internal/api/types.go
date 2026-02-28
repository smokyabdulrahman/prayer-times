package api

// Response represents the top-level Al Adhan API response.
type Response struct {
	Code   int    `json:"code"`
	Status string `json:"status"`
	Data   Data   `json:"data"`
}

// Data holds the prayer timings, date info, and metadata.
type Data struct {
	Timings Timings  `json:"timings"`
	Date    DateInfo `json:"date"`
	Meta    Meta     `json:"meta"`
}

// Timings contains all prayer and event times as HH:MM strings.
// The API may include a timezone suffix like " (BST)" which we strip during parsing.
type Timings struct {
	Fajr       string `json:"Fajr"`
	Sunrise    string `json:"Sunrise"`
	Dhuhr      string `json:"Dhuhr"`
	Asr        string `json:"Asr"`
	Sunset     string `json:"Sunset"`
	Maghrib    string `json:"Maghrib"`
	Isha       string `json:"Isha"`
	Imsak      string `json:"Imsak"`
	Midnight   string `json:"Midnight"`
	Firstthird string `json:"Firstthird"`
	Lastthird  string `json:"Lastthird"`
}

// DateInfo contains date representations.
type DateInfo struct {
	Readable  string `json:"readable"`
	Timestamp string `json:"timestamp"`
}

// Meta contains request metadata returned by the API.
type Meta struct {
	Latitude  float64    `json:"latitude"`
	Longitude float64    `json:"longitude"`
	Timezone  string     `json:"timezone"`
	Method    MethodInfo `json:"method"`
	School    string     `json:"school"`
}

// MethodInfo identifies the calculation method used.
type MethodInfo struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
