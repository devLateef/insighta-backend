package handlers

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"insighta/internal/models"
	"insighta/internal/storage"

	"github.com/gin-gonic/gin"
)

// GET /api/profiles/export?format=csv
func ExportCSV(c *gin.Context) {
	format := c.DefaultQuery("format", "csv")
	if format != "csv" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "only csv format is supported"})
		return
	}

	// Apply same filters as GET /api/profiles but fetch all (no pagination limit)
	filter, httpErr := parseFilter(c)
	if httpErr != nil {
		c.JSON(httpErr.Code, gin.H{"status": "error", "message": httpErr.Message})
		return
	}
	filter.Page = 1
	filter.Limit = 100000 // export all matching

	profiles, _, err := storage.ListProfiles(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("profiles_%s.csv", timestamp)

	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Header("Content-Type", "text/csv")
	c.Header("Cache-Control", "no-cache")

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	// Header row — exact columns per spec
	writer.Write([]string{
		"id", "name", "gender", "gender_probability",
		"age", "age_group", "country_id", "country_name",
		"country_probability", "created_at",
	})

	for _, p := range profiles {
		writer.Write(profileToCSVRow(p))
	}
}

func profileToCSVRow(p *models.Profile) []string {
	return []string{
		p.ID,
		p.Name,
		p.Gender,
		strconv.FormatFloat(p.GenderProbability, 'f', 4, 64),
		strconv.Itoa(p.Age),
		p.AgeGroup,
		p.CountryID,
		p.CountryName,
		strconv.FormatFloat(p.CountryProbability, 'f', 4, 64),
		p.CreatedAt.Format(time.RFC3339),
	}
}
