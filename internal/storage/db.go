package storage

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"insighta/internal/models"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func Init(dsn string) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("failed to open db:", err)
	}

	db.SetMaxOpenConns(5) // Neon free tier has limited connections
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		log.Fatal("failed to ping db:", err)
	}

	DB = db
	log.Println("Database connected")
}

// ─── Users ────────────────────────────────────────────────────────────────────

func UpsertUser(u *models.User) error {
	now := time.Now()
	_, err := DB.Exec(`
		INSERT INTO users (id, github_id, username, email, avatar_url, role, is_active, last_login_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (github_id) DO UPDATE SET
			username     = EXCLUDED.username,
			email        = EXCLUDED.email,
			avatar_url   = EXCLUDED.avatar_url,
			last_login_at = EXCLUDED.last_login_at
	`, u.ID, u.GithubID, u.Username, u.Email, u.AvatarURL, u.Role, u.IsActive, now, u.CreatedAt)
	return err
}

func GetUserByGithubID(githubID string) (*models.User, error) {
	u := &models.User{}
	err := DB.QueryRow(`
		SELECT id, github_id, username, email, avatar_url, role, is_active, last_login_at, created_at
		FROM users WHERE github_id = $1
	`, githubID).Scan(
		&u.ID, &u.GithubID, &u.Username, &u.Email, &u.AvatarURL,
		&u.Role, &u.IsActive, &u.LastLoginAt, &u.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func GetUserByID(id string) (*models.User, error) {
	u := &models.User{}
	err := DB.QueryRow(`
		SELECT id, github_id, username, email, avatar_url, role, is_active, last_login_at, created_at
		FROM users WHERE id = $1
	`, id).Scan(
		&u.ID, &u.GithubID, &u.Username, &u.Email, &u.AvatarURL,
		&u.Role, &u.IsActive, &u.LastLoginAt, &u.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// ─── Refresh Tokens ───────────────────────────────────────────────────────────

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", h)
}

func StoreRefreshToken(userID, token string, expiresAt time.Time) error {
	hash := hashToken(token)
	_, err := DB.Exec(`
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`, userID, hash, expiresAt)
	return err
}

func ValidateAndRotateRefreshToken(token string) (string, error) {
	hash := hashToken(token)
	var userID string
	var expiresAt time.Time

	err := DB.QueryRow(`
		SELECT user_id, expires_at FROM refresh_tokens
		WHERE token_hash = $1
	`, hash).Scan(&userID, &expiresAt)
	if err != nil {
		return "", fmt.Errorf("refresh token not found")
	}

	if time.Now().After(expiresAt) {
		// Clean up expired token
		_, _ = DB.Exec(`DELETE FROM refresh_tokens WHERE token_hash = $1`, hash)
		return "", fmt.Errorf("refresh token expired")
	}

	// Invalidate immediately (rotation)
	_, err = DB.Exec(`DELETE FROM refresh_tokens WHERE token_hash = $1`, hash)
	if err != nil {
		return "", err
	}

	return userID, nil
}

func InvalidateRefreshToken(token string) error {
	hash := hashToken(token)
	_, err := DB.Exec(`DELETE FROM refresh_tokens WHERE token_hash = $1`, hash)
	return err
}

func InvalidateAllUserTokens(userID string) error {
	_, err := DB.Exec(`DELETE FROM refresh_tokens WHERE user_id = $1`, userID)
	return err
}

// ─── Profiles ─────────────────────────────────────────────────────────────────

func CreateProfile(p *models.Profile) error {
	_, err := DB.Exec(`
		INSERT INTO profiles (id, name, gender, gender_probability, age, age_group, country_id, country_name, country_probability, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, p.ID, p.Name, p.Gender, p.GenderProbability, p.Age, p.AgeGroup,
		p.CountryID, p.CountryName, p.CountryProbability, p.CreatedAt)
	return err
}

// SeedProfile inserts a profile, silently skipping duplicates (by id or name).
func SeedProfile(p *models.Profile) error {
	_, err := DB.Exec(`
		INSERT INTO profiles (id, name, gender, gender_probability, age, age_group, country_id, country_name, country_probability, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT DO NOTHING
	`, p.ID, p.Name, p.Gender, p.GenderProbability, p.Age, p.AgeGroup,
		p.CountryID, p.CountryName, p.CountryProbability, p.CreatedAt)
	return err
}

func GetProfileByID(id string) (*models.Profile, error) {
	p := &models.Profile{}
	err := DB.QueryRow(`
		SELECT id, name, gender, gender_probability, age, age_group, country_id, country_name, country_probability, created_at
		FROM profiles WHERE id = $1
	`, id).Scan(
		&p.ID, &p.Name, &p.Gender, &p.GenderProbability, &p.Age, &p.AgeGroup,
		&p.CountryID, &p.CountryName, &p.CountryProbability, &p.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func ListProfiles(f models.ProfileFilter) ([]*models.Profile, int, error) {
	where := []string{"1=1"}
	filterArgs := []any{}
	idx := 1

	if f.Gender != "" {
		where = append(where, fmt.Sprintf("gender = $%d", idx))
		filterArgs = append(filterArgs, f.Gender)
		idx++
	}
	if f.CountryID != "" {
		where = append(where, fmt.Sprintf("country_id = $%d", idx))
		filterArgs = append(filterArgs, strings.ToUpper(f.CountryID))
		idx++
	}
	if f.AgeGroup != "" {
		where = append(where, fmt.Sprintf("age_group = $%d", idx))
		filterArgs = append(filterArgs, f.AgeGroup)
		idx++
	}
	if f.MinAge > 0 {
		where = append(where, fmt.Sprintf("age >= $%d", idx))
		filterArgs = append(filterArgs, f.MinAge)
		idx++
	}
	if f.MaxAge > 0 {
		where = append(where, fmt.Sprintf("age <= $%d", idx))
		filterArgs = append(filterArgs, f.MaxAge)
		idx++
	}
	if f.MinGenderProbability > 0 {
		where = append(where, fmt.Sprintf("gender_probability >= $%d", idx))
		filterArgs = append(filterArgs, f.MinGenderProbability)
		idx++
	}
	if f.MinCountryProbability > 0 {
		where = append(where, fmt.Sprintf("country_probability >= $%d", idx))
		filterArgs = append(filterArgs, f.MinCountryProbability)
		idx++
	}

	whereClause := strings.Join(where, " AND ")

	// Count — uses only filter args (no pagination args)
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM profiles WHERE %s", whereClause)
	if err := DB.QueryRow(countQuery, filterArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Sort
	allowedSort := map[string]bool{
		"age":                 true,
		"created_at":          true,
		"gender_probability":  true,
		"country_probability": true,
	}
	sortBy := "created_at"
	if allowedSort[f.SortBy] {
		sortBy = f.SortBy
	}
	order := "DESC"
	if strings.ToUpper(f.Order) == "ASC" {
		order = "ASC"
	}

	// Pagination
	if f.Limit <= 0 {
		f.Limit = 10
	}
	if f.Limit > 50 {
		f.Limit = 50
	}
	if f.Page <= 0 {
		f.Page = 1
	}
	offset := (f.Page - 1) * f.Limit

	// Build query args: filter args + pagination args (separate slice to avoid mutation)
	queryArgs := make([]any, len(filterArgs), len(filterArgs)+2)
	copy(queryArgs, filterArgs)
	queryArgs = append(queryArgs, f.Limit, offset)

	query := fmt.Sprintf(`
		SELECT id, name, gender, gender_probability, age, age_group,
		       country_id, country_name, country_probability, created_at
		FROM profiles
		WHERE %s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, sortBy, order, idx, idx+1)

	rows, err := DB.Query(query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var profiles []*models.Profile
	for rows.Next() {
		p := &models.Profile{}
		if err := rows.Scan(
			&p.ID, &p.Name, &p.Gender, &p.GenderProbability, &p.Age, &p.AgeGroup,
			&p.CountryID, &p.CountryName, &p.CountryProbability, &p.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		profiles = append(profiles, p)
	}

	return profiles, total, nil
}

func DeleteProfile(id string) error {
	result, err := DB.Exec(`DELETE FROM profiles WHERE id = $1`, id)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("profile not found")
	}
	return nil
}
