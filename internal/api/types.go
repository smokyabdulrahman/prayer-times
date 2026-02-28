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
	Readable  string        `json:"readable"`
	Timestamp string        `json:"timestamp"`
	Hijri     HijriDate     `json:"hijri"`
	Gregorian GregorianDate `json:"gregorian"`
}

// HijriDate represents the Hijri (Islamic) date from the API response.
type HijriDate struct {
	Date        string           `json:"date"` // e.g. "10-08-1447"
	Day         string           `json:"day"`
	Month       HijriMonth       `json:"month"`
	Year        string           `json:"year"`
	Designation HijriDesignation `json:"designation"`
}

// HijriMonth represents the month in the Hijri calendar.
type HijriMonth struct {
	Number int    `json:"number"`
	En     string `json:"en"` // English name, e.g. "Shaʿbān"
	Ar     string `json:"ar"` // Arabic name
}

// HijriDesignation contains the calendar designation labels.
type HijriDesignation struct {
	Abbreviated string `json:"abbreviated"` // "AH"
	Expanded    string `json:"expanded"`    // "Anno Hegirae"
}

// Format returns the Hijri date as "DD MonthName YYYY AH".
func (h HijriDate) Format() string {
	if h.Day == "" || h.Month.En == "" || h.Year == "" {
		return ""
	}
	abbr := h.Designation.Abbreviated
	if abbr == "" {
		abbr = "AH"
	}
	return h.Day + " " + h.Month.En + " " + h.Year + " " + abbr
}

// GregorianDate represents the Gregorian date from the API response.
type GregorianDate struct {
	Date    string         `json:"date"` // e.g. "28-02-2026"
	Day     string         `json:"day"`
	Weekday GregorianDay   `json:"weekday"`
	Month   GregorianMonth `json:"month"`
	Year    string         `json:"year"`
}

// GregorianDay contains the weekday name.
type GregorianDay struct {
	En string `json:"en"` // e.g. "Saturday"
}

// GregorianMonth contains the month details.
type GregorianMonth struct {
	Number int    `json:"number"`
	En     string `json:"en"` // e.g. "February"
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

// CalendarResponse represents the Al Adhan calendar API response.
// The calendar endpoint returns an array of daily data objects for a whole month.
type CalendarResponse struct {
	Code   int    `json:"code"`
	Status string `json:"status"`
	Data   []Data `json:"data"`
}
