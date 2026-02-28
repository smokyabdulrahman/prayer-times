package api

import "testing"

func TestHijriDate_Format(t *testing.T) {
	tests := []struct {
		name string
		h    HijriDate
		want string
	}{
		{
			name: "full date",
			h: HijriDate{
				Day:         "10",
				Month:       HijriMonth{Number: 8, En: "Sha'ban"},
				Year:        "1447",
				Designation: HijriDesignation{Abbreviated: "AH"},
			},
			want: "10 Sha'ban 1447 AH",
		},
		{
			name: "missing abbreviated defaults to AH",
			h: HijriDate{
				Day:   "1",
				Month: HijriMonth{Number: 1, En: "Muharram"},
				Year:  "1448",
			},
			want: "1 Muharram 1448 AH",
		},
		{
			name: "empty day returns empty",
			h: HijriDate{
				Month: HijriMonth{En: "Ramadan"},
				Year:  "1447",
			},
			want: "",
		},
		{
			name: "empty month returns empty",
			h: HijriDate{
				Day:  "15",
				Year: "1447",
			},
			want: "",
		},
		{
			name: "empty year returns empty",
			h: HijriDate{
				Day:   "15",
				Month: HijriMonth{En: "Ramadan"},
			},
			want: "",
		},
		{
			name: "all empty returns empty",
			h:    HijriDate{},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.h.Format()
			if got != tt.want {
				t.Errorf("HijriDate.Format() = %q, want %q", got, tt.want)
			}
		})
	}
}
