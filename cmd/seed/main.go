// cmd/seed/main.go — seeds the profiles table from a JSON file.
//
// Usage:
//
//	go run ./cmd/seed --file=profiles.json
//	go run ./cmd/seed --file=profiles.json --reset   # truncate first
//
// Supports JSON formats:
//   - Plain array:              [ {...}, {...} ]
//   - Wrapped in "data":        { "data": [ {...} ] }
//   - Wrapped in "profiles":    { "profiles": [ {...} ] }
//   - Wrapped in "results":     { "results": [ {...} ] }
//
// Re-running is safe — duplicates are skipped via ON CONFLICT DO NOTHING.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"insighta/configs"
	"insighta/internal/models"
	"insighta/internal/storage"

	"github.com/google/uuid"
)

func main() {
	filePath := flag.String("file", "profiles.json", "path to JSON seed file")
	reset := flag.Bool("reset", false, "truncate profiles table before seeding")
	flag.Parse()

	configs.Load()
	storage.Init(configs.AppConfig.DatabaseURL)

	if *reset {
		log.Println("Truncating profiles table...")
		if _, err := storage.DB.Exec("TRUNCATE TABLE profiles RESTART IDENTITY CASCADE"); err != nil {
			log.Fatal("truncate failed:", err)
		}
	}

	f, err := os.Open(*filePath)
	if err != nil {
		log.Fatalf("cannot open %s: %v", *filePath, err)
	}
	defer f.Close()

	// Decode into a generic value first so we can handle any shape
	var raw json.RawMessage
	if err := json.NewDecoder(f).Decode(&raw); err != nil {
		log.Fatal("failed to decode JSON:", err)
	}

	profiles, err := extractProfiles(raw)
	if err != nil {
		log.Fatal("failed to extract profiles:", err)
	}

	log.Printf("Seeding %d profiles...", len(profiles))

	inserted, skipped, failed := 0, 0, 0

	for i, item := range profiles {
		var p models.Profile
		if err := json.Unmarshal(item, &p); err != nil {
			log.Printf("[%d] parse error: %v — raw: %s", i, err, string(item))
			failed++
			continue
		}

		if p.ID == "" {
			p.ID = uuid.New().String()
		}
		if p.Name == "" {
			log.Printf("[%d] missing name, skipping", i)
			failed++
			continue
		}
		if p.CreatedAt.IsZero() {
			p.CreatedAt = time.Now().UTC()
		}
		if p.AgeGroup == "" {
			p.AgeGroup = classifyAgeGroup(p.Age)
		}

		if err := storage.SeedProfile(&p); err != nil {
			if isUniqueViolation(err) {
				skipped++
			} else {
				log.Printf("[%d] insert error for %q: %v", i, p.Name, err)
				failed++
			}
			continue
		}
		inserted++

		if (inserted+skipped)%100 == 0 {
			fmt.Printf("  progress: %d inserted, %d skipped, %d failed\n", inserted, skipped, failed)
		}
	}

	log.Printf("Done. inserted=%d skipped=%d failed=%d", inserted, skipped, failed)
}

// extractProfiles handles multiple JSON shapes:
//   - plain array
//   - object with "data", "profiles", "results", or any single array field
func extractProfiles(raw json.RawMessage) ([]json.RawMessage, error) {
	trimmed := strings.TrimSpace(string(raw))

	// Plain array
	if strings.HasPrefix(trimmed, "[") {
		var arr []json.RawMessage
		if err := json.Unmarshal(raw, &arr); err != nil {
			return nil, err
		}
		return arr, nil
	}

	// Object — try known wrapper keys first
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(raw, &obj); err != nil {
		return nil, fmt.Errorf("JSON is neither an array nor an object: %v", err)
	}

	for _, key := range []string{"data", "profiles", "results", "items", "records"} {
		if val, ok := obj[key]; ok {
			var arr []json.RawMessage
			if err := json.Unmarshal(val, &arr); err == nil {
				log.Printf("Found profiles under key %q", key)
				return arr, nil
			}
		}
	}

	// Fall back: find the first field that is an array
	for key, val := range obj {
		var arr []json.RawMessage
		if err := json.Unmarshal(val, &arr); err == nil && len(arr) > 0 {
			log.Printf("Found profiles under key %q", key)
			return arr, nil
		}
	}

	return nil, fmt.Errorf("could not find a profiles array in the JSON — keys found: %v", keys(obj))
}

func keys(m map[string]json.RawMessage) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
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

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return strings.Contains(s, "unique") || strings.Contains(s, "duplicate")
}
