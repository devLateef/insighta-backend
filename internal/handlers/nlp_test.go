package handlers

import (
	"testing"
)

func TestParseNaturalLanguage(t *testing.T) {
	tests := []struct {
		query       string
		wantOK      bool
		wantGender  string
		wantMinAge  int
		wantMaxAge  int
		wantGroup   string
		wantCountry string
	}{
		{
			query:       "young males from nigeria",
			wantOK:      true,
			wantGender:  "male",
			wantMinAge:  16,
			wantMaxAge:  24,
			wantCountry: "NG",
		},
		{
			query:      "females above 30",
			wantOK:     true,
			wantGender: "female",
			wantMinAge: 30,
		},
		{
			query:       "people from angola",
			wantOK:      true,
			wantCountry: "AO",
		},
		{
			query:       "adult males from kenya",
			wantOK:      true,
			wantGender:  "male",
			wantGroup:   "adult",
			wantCountry: "KE",
		},
		{
			query:      "male and female teenagers above 17",
			wantOK:     true,
			wantGroup:  "teenager",
			wantMinAge: 17,
		},
		{
			query:      "men over 25",
			wantOK:     true,
			wantGender: "male",
			wantMinAge: 25,
		},
		{
			query:      "women under 40",
			wantOK:     true,
			wantGender: "female",
			wantMaxAge: 40,
		},
		{
			query:       "seniors from ghana",
			wantOK:      true,
			wantGroup:   "senior",
			wantCountry: "GH",
		},
		{
			query:       "nigerian males",
			wantOK:      true,
			wantGender:  "male",
			wantCountry: "NG",
		},
		{
			query:  "xyzzy foobar",
			wantOK: false,
		},
		{
			query:  "",
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			f, ok := parseNaturalLanguage(tt.query)
			if ok != tt.wantOK {
				t.Errorf("parseNaturalLanguage(%q) ok=%v, want %v", tt.query, ok, tt.wantOK)
				return
			}
			if !ok {
				return
			}
			if tt.wantGender != "" && f.Gender != tt.wantGender {
				t.Errorf("gender: got %q, want %q", f.Gender, tt.wantGender)
			}
			if tt.wantMinAge != 0 && f.MinAge != tt.wantMinAge {
				t.Errorf("min_age: got %d, want %d", f.MinAge, tt.wantMinAge)
			}
			if tt.wantMaxAge != 0 && f.MaxAge != tt.wantMaxAge {
				t.Errorf("max_age: got %d, want %d", f.MaxAge, tt.wantMaxAge)
			}
			if tt.wantGroup != "" && f.AgeGroup != tt.wantGroup {
				t.Errorf("age_group: got %q, want %q", f.AgeGroup, tt.wantGroup)
			}
			if tt.wantCountry != "" && f.CountryID != tt.wantCountry {
				t.Errorf("country_id: got %q, want %q", f.CountryID, tt.wantCountry)
			}
		})
	}
}
