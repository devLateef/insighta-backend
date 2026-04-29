package models

import "time"

type User struct {
	ID          string     `json:"id"`
	GithubID    string     `json:"github_id"`
	Username    string     `json:"username"`
	Email       string     `json:"email"`
	AvatarURL   string     `json:"avatar_url"`
	Role        string     `json:"role"`
	IsActive    bool       `json:"is_active"`
	LastLoginAt *time.Time `json:"last_login_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

type Profile struct {
	ID                 string    `json:"id"`
	Name               string    `json:"name"`
	Gender             string    `json:"gender"`
	GenderProbability  float64   `json:"gender_probability"`
	Age                int       `json:"age"`
	AgeGroup           string    `json:"age_group"`
	CountryID          string    `json:"country_id"`
	CountryName        string    `json:"country_name"`
	CountryProbability float64   `json:"country_probability"`
	CreatedAt          time.Time `json:"created_at"`
}

type ProfileFilter struct {
	Gender                string
	CountryID             string
	AgeGroup              string
	MinAge                int
	MaxAge                int
	MinGenderProbability  float64
	MinCountryProbability float64
	SortBy                string
	Order                 string
	Page                  int
	Limit                 int
}
