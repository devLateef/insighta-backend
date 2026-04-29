package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

var httpClient = &http.Client{Timeout: 8 * time.Second}

// ─── Gender API (genderize.io) ────────────────────────────────────────────────

func fetchGender(name string) (string, float64) {
	firstName := strings.Fields(name)[0]
	url := fmt.Sprintf("https://api.genderize.io/?name=%s", firstName)

	resp, err := httpClient.Get(url)
	if err != nil {
		return "unknown", 0
	}
	defer resp.Body.Close()

	var result struct {
		Gender      string  `json:"gender"`
		Probability float64 `json:"probability"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "unknown", 0
	}

	if result.Gender == "" {
		return "unknown", 0
	}
	return result.Gender, result.Probability
}

// ─── Age API (agify.io) ───────────────────────────────────────────────────────

func fetchAge(name string) (int, string) {
	firstName := strings.Fields(name)[0]
	url := fmt.Sprintf("https://api.agify.io/?name=%s", firstName)

	resp, err := httpClient.Get(url)
	if err != nil {
		return 0, "unknown"
	}
	defer resp.Body.Close()

	var result struct {
		Age int `json:"age"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, "unknown"
	}

	return result.Age, classifyAgeGroup(result.Age)
}

func classifyAgeGroup(age int) string {
	switch {
	case age <= 0:
		return "unknown"
	case age < 13:
		return "child"
	case age < 18:
		return "teenager"
	case age < 60:
		return "adult"
	default:
		return "senior"
	}
}

// ─── Nationality API (nationalize.io) ─────────────────────────────────────────

func fetchNationality(name string) (string, string, float64) {
	firstName := strings.Fields(name)[0]
	url := fmt.Sprintf("https://api.nationalize.io/?name=%s", firstName)

	resp, err := httpClient.Get(url)
	if err != nil {
		return "", "Unknown", 0
	}
	defer resp.Body.Close()

	var result struct {
		Country []struct {
			CountryID   string  `json:"country_id"`
			Probability float64 `json:"probability"`
		} `json:"country"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "Unknown", 0
	}

	if len(result.Country) == 0 {
		return "", "Unknown", 0
	}

	top := result.Country[0]
	countryName := countryCodeToName(top.CountryID)
	return top.CountryID, countryName, top.Probability
}

// countryCodeToName maps ISO 3166-1 alpha-2 codes to country names.
func countryCodeToName(code string) string {
	countries := map[string]string{
		"US": "United States", "GB": "United Kingdom", "NG": "Nigeria",
		"GH": "Ghana", "KE": "Kenya", "ZA": "South Africa", "EG": "Egypt",
		"DE": "Germany", "FR": "France", "IT": "Italy", "ES": "Spain",
		"BR": "Brazil", "MX": "Mexico", "AR": "Argentina", "CO": "Colombia",
		"IN": "India", "CN": "China", "JP": "Japan", "KR": "South Korea",
		"AU": "Australia", "CA": "Canada", "RU": "Russia", "TR": "Turkey",
		"PK": "Pakistan", "BD": "Bangladesh", "PH": "Philippines", "ID": "Indonesia",
		"TH": "Thailand", "VN": "Vietnam", "MY": "Malaysia", "SG": "Singapore",
		"NL": "Netherlands", "BE": "Belgium", "SE": "Sweden", "NO": "Norway",
		"DK": "Denmark", "FI": "Finland", "PL": "Poland", "PT": "Portugal",
		"GR": "Greece", "CZ": "Czech Republic", "HU": "Hungary", "RO": "Romania",
		"UA": "Ukraine", "NZ": "New Zealand", "ZW": "Zimbabwe", "TZ": "Tanzania",
		"UG": "Uganda", "ET": "Ethiopia", "SN": "Senegal", "CI": "Ivory Coast",
		"CM": "Cameroon", "AO": "Angola", "MZ": "Mozambique", "MG": "Madagascar",
		"IL": "Israel", "SA": "Saudi Arabia", "AE": "United Arab Emirates",
		"IQ": "Iraq", "IR": "Iran", "AF": "Afghanistan", "PG": "Papua New Guinea",
	}
	if name, ok := countries[code]; ok {
		return name
	}
	return code
}
