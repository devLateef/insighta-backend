package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"

	"insighta/internal/models"
	"insighta/internal/storage"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ─── GET /api/profiles ────────────────────────────────────────────────────────
func GetProfiles(c *gin.Context) {
	filter, httpErr := parseFilter(c)
	if httpErr != nil {
		c.JSON(httpErr.Code, gin.H{"status": "error", "message": httpErr.Message})
		return
	}

	profiles, total, err := storage.ListProfiles(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "internal server error"})
		return
	}

	if profiles == nil {
		profiles = []*models.Profile{}
	}

	c.JSON(http.StatusOK, buildPaginatedResponse(profiles, total, filter, "/api/profiles"))
}

// ─── GET /api/profiles/search ─────────────────────────────────────────────────
func SearchProfiles(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	if q == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "q parameter required"})
		return
	}

	filter, ok := parseNaturalLanguage(q)
	if !ok {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"status": "error", "message": "Unable to interpret query"})
		return
	}

	// Apply pagination from query params
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}
	filter.Page = page
	filter.Limit = limit
	filter.SortBy = c.DefaultQuery("sort_by", "created_at")
	filter.Order = c.DefaultQuery("order", "desc")

	profiles, total, err := storage.ListProfiles(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "internal server error"})
		return
	}

	if profiles == nil {
		profiles = []*models.Profile{}
	}

	c.JSON(http.StatusOK, buildPaginatedResponse(profiles, total, filter, "/api/profiles/search"))
}

// ─── GET /api/profiles/:id ────────────────────────────────────────────────────
func GetProfile(c *gin.Context) {
	id := c.Param("id")
	profile, err := storage.GetProfileByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "profile not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": profile})
}

// ─── POST /api/profiles ───────────────────────────────────────────────────────
func CreateProfile(c *gin.Context) {
	var body struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "name is required"})
		return
	}

	name := strings.TrimSpace(body.Name)
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "name cannot be empty"})
		return
	}

	gender, genderProb := fetchGender(name)
	age, ageGroup := fetchAge(name)
	countryID, countryName, countryProb := fetchNationality(name)

	profile := &models.Profile{
		ID:                 uuid.New().String(),
		Name:               name,
		Gender:             gender,
		GenderProbability:  genderProb,
		Age:                age,
		AgeGroup:           ageGroup,
		CountryID:          countryID,
		CountryName:        countryName,
		CountryProbability: countryProb,
		CreatedAt:          time.Now().UTC(),
	}

	if err := storage.CreateProfile(profile); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "failed to save profile"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"status": "success", "data": profile})
}

// ─── DELETE /api/profiles/:id ─────────────────────────────────────────────────
func DeleteProfile(c *gin.Context) {
	id := c.Param("id")
	if err := storage.DeleteProfile(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "profile not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "profile deleted"})
}

// ─── Filter parsing ───────────────────────────────────────────────────────────

type httpError struct {
	Code    int
	Message string
}

func parseFilter(c *gin.Context) (models.ProfileFilter, *httpError) {
	f := models.ProfileFilter{}

	// Pagination
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		return f, &httpError{http.StatusUnprocessableEntity, "Invalid query parameters"}
	}
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil {
		return f, &httpError{http.StatusUnprocessableEntity, "Invalid query parameters"}
	}
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}
	f.Page = page
	f.Limit = limit

	// Filters
	f.Gender = strings.ToLower(c.Query("gender"))
	f.CountryID = strings.ToUpper(c.Query("country_id"))
	f.AgeGroup = strings.ToLower(c.Query("age_group"))

	if v := c.Query("min_age"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			return f, &httpError{http.StatusUnprocessableEntity, "Invalid query parameters"}
		}
		f.MinAge = n
	}
	if v := c.Query("max_age"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			return f, &httpError{http.StatusUnprocessableEntity, "Invalid query parameters"}
		}
		f.MaxAge = n
	}
	if v := c.Query("min_gender_probability"); v != "" {
		n, err := strconv.ParseFloat(v, 64)
		if err != nil || n < 0 || n > 1 {
			return f, &httpError{http.StatusUnprocessableEntity, "Invalid query parameters"}
		}
		f.MinGenderProbability = n
	}
	if v := c.Query("min_country_probability"); v != "" {
		n, err := strconv.ParseFloat(v, 64)
		if err != nil || n < 0 || n > 1 {
			return f, &httpError{http.StatusUnprocessableEntity, "Invalid query parameters"}
		}
		f.MinCountryProbability = n
	}

	// Sorting
	sortBy := c.DefaultQuery("sort_by", "created_at")
	allowed := map[string]bool{"age": true, "created_at": true, "gender_probability": true, "country_probability": true}
	if sortBy != "" && !allowed[sortBy] {
		return f, &httpError{http.StatusUnprocessableEntity, "Invalid query parameters"}
	}
	f.SortBy = sortBy
	f.Order = c.DefaultQuery("order", "desc")
	if f.Order != "asc" && f.Order != "desc" {
		return f, &httpError{http.StatusUnprocessableEntity, "Invalid query parameters"}
	}

	return f, nil
}

// ─── Natural language parser ──────────────────────────────────────────────────
//
// Supported patterns (rule-based, no AI):
//
//	Gender:    "male", "males", "female", "females", "men", "women", "boys", "girls"
//	Age group: "young" (16-24), "teenager"/"teenagers", "adult"/"adults",
//	           "child"/"children", "senior"/"seniors", "old"/"elderly"
//	Age mods:  "above N", "over N", "below N", "under N", "older than N", "younger than N"
//	Country:   "from <country>" → mapped to ISO code
//	           also bare country names/adjectives: "nigerian", "kenyan", etc.
func parseNaturalLanguage(q string) (models.ProfileFilter, bool) {
	f := models.ProfileFilter{}
	words := tokenize(q)

	if len(words) == 0 {
		return f, false
	}

	matched := false

	// ── Gender ────────────────────────────────────────────────────────────────
	for _, w := range words {
		switch w {
		case "male", "males", "men", "man", "boy", "boys":
			f.Gender = "male"
			matched = true
		case "female", "females", "women", "woman", "girl", "girls":
			f.Gender = "female"
			matched = true
		}
	}

	// ── Age group / young keyword ─────────────────────────────────────────────
	for _, w := range words {
		switch w {
		case "young":
			// "young" maps to 16–24 for parsing only
			f.MinAge = 16
			f.MaxAge = 24
			matched = true
		case "teenager", "teenagers", "teen", "teens":
			f.AgeGroup = "teenager"
			matched = true
		case "adult", "adults":
			f.AgeGroup = "adult"
			matched = true
		case "child", "children", "kid", "kids":
			f.AgeGroup = "child"
			matched = true
		case "senior", "seniors", "elderly", "old":
			f.AgeGroup = "senior"
			matched = true
		}
	}

	// ── Age modifiers: "above N", "over N", "below N", "under N" ─────────────
	for i, w := range words {
		if i+1 >= len(words) {
			break
		}
		next := words[i+1]
		n, err := strconv.Atoi(next)
		if err != nil {
			continue
		}
		switch w {
		case "above", "over":
			f.MinAge = n
			matched = true
		case "below", "under":
			f.MaxAge = n
			matched = true
		case "than":
			// "older than N" / "younger than N"
			if i >= 1 {
				switch words[i-1] {
				case "older":
					f.MinAge = n
					matched = true
				case "younger":
					f.MaxAge = n
					matched = true
				}
			}
		}
	}

	// ── Country: "from <country>" or bare adjective ───────────────────────────
	fromIdx := -1
	for i, w := range words {
		if w == "from" {
			fromIdx = i
			break
		}
	}

	if fromIdx >= 0 && fromIdx+1 < len(words) {
		// Collect remaining words after "from" as country name
		countryPhrase := strings.Join(words[fromIdx+1:], " ")
		if code := resolveCountry(countryPhrase); code != "" {
			f.CountryID = code
			matched = true
		}
	}

	// Also check bare country adjectives anywhere in the query
	if f.CountryID == "" {
		for _, w := range words {
			if code := resolveCountryAdjective(w); code != "" {
				f.CountryID = code
				matched = true
				break
			}
		}
	}

	if !matched {
		return f, false
	}

	return f, true
}

// tokenize lowercases and splits on whitespace/punctuation.
func tokenize(s string) []string {
	s = strings.ToLower(s)
	fields := strings.FieldsFunc(s, func(r rune) bool {
		return unicode.IsSpace(r) || r == ',' || r == '.' || r == ';'
	})
	return fields
}

// resolveCountry maps a country name/phrase to an ISO 3166-1 alpha-2 code.
func resolveCountry(name string) string {
	name = strings.TrimSpace(strings.ToLower(name))
	m := countryNameMap()
	// Try full phrase first, then first word
	if code, ok := m[name]; ok {
		return code
	}
	parts := strings.Fields(name)
	if len(parts) > 0 {
		if code, ok := m[parts[0]]; ok {
			return code
		}
	}
	return ""
}

// resolveCountryAdjective maps demonyms/adjectives to ISO codes.
func resolveCountryAdjective(word string) string {
	adj := map[string]string{
		"nigerian": "NG", "nigerien": "NE", "ghanaian": "GH", "kenyan": "KE",
		"south african": "ZA", "egyptian": "EG", "ethiopian": "ET", "tanzanian": "TZ",
		"ugandan": "UG", "senegalese": "SN", "cameroonian": "CM", "angolan": "AO",
		"zimbabwean": "ZW", "ivorian": "CI", "congolese": "CD",
		"american": "US", "british": "GB", "german": "DE", "french": "FR",
		"italian": "IT", "spanish": "ES", "portuguese": "PT", "dutch": "NL",
		"swedish": "SE", "norwegian": "NO", "danish": "DK", "finnish": "FI",
		"polish": "PL", "greek": "GR", "turkish": "TR", "russian": "RU",
		"ukrainian": "UA", "romanian": "RO", "hungarian": "HU", "czech": "CZ",
		"brazilian": "BR", "mexican": "MX", "argentinian": "AR", "colombian": "CO",
		"peruvian": "PE", "chilean": "CL", "venezuelan": "VE",
		"indian": "IN", "chinese": "CN", "japanese": "JP", "korean": "KR",
		"pakistani": "PK", "bangladeshi": "BD", "indonesian": "ID", "filipino": "PH",
		"thai": "TH", "vietnamese": "VN", "malaysian": "MY", "singaporean": "SG",
		"australian": "AU", "canadian": "CA", "new zealander": "NZ",
		"saudi": "SA", "emirati": "AE", "iraqi": "IQ", "iranian": "IR",
		"israeli": "IL", "lebanese": "LB",
		"beninese": "BJ", "togolese": "TG", "malian": "ML", "burkinabe": "BF",
		"guinean": "GN", "sierra leonean": "SL", "liberian": "LR",
		"rwandan": "RW", "burundian": "BI", "zambian": "ZM", "malawian": "MW",
		"mozambican": "MZ", "malagasy": "MG", "sudanese": "SD", "libyan": "LY",
		"tunisian": "TN", "algerian": "DZ", "moroccan": "MA",
	}
	return adj[word]
}

func countryNameMap() map[string]string {
	return map[string]string{
		"nigeria": "NG", "niger": "NE", "ghana": "GH", "kenya": "KE",
		"south africa": "ZA", "egypt": "EG", "ethiopia": "ET", "tanzania": "TZ",
		"uganda": "UG", "senegal": "SN", "cameroon": "CM", "angola": "AO",
		"zimbabwe": "ZW", "ivory coast": "CI", "cote d'ivoire": "CI",
		"congo": "CD", "democratic republic of congo": "CD", "drc": "CD",
		"benin": "BJ", "togo": "TG", "mali": "ML", "burkina faso": "BF",
		"guinea": "GN", "sierra leone": "SL", "liberia": "LR",
		"rwanda": "RW", "burundi": "BI", "zambia": "ZM", "malawi": "MW",
		"mozambique": "MZ", "madagascar": "MG", "sudan": "SD", "libya": "LY",
		"tunisia": "TN", "algeria": "DZ", "morocco": "MA",
		"united states": "US", "usa": "US", "america": "US",
		"united kingdom": "GB", "uk": "GB", "britain": "GB", "england": "GB",
		"germany": "DE", "france": "FR", "italy": "IT", "spain": "ES",
		"portugal": "PT", "netherlands": "NL", "belgium": "BE",
		"sweden": "SE", "norway": "NO", "denmark": "DK", "finland": "FI",
		"poland": "PL", "greece": "GR", "turkey": "TR", "russia": "RU",
		"ukraine": "UA", "romania": "RO", "hungary": "HU", "czech republic": "CZ",
		"brazil": "BR", "mexico": "MX", "argentina": "AR", "colombia": "CO",
		"peru": "PE", "chile": "CL", "venezuela": "VE",
		"india": "IN", "china": "CN", "japan": "JP", "south korea": "KR", "korea": "KR",
		"pakistan": "PK", "bangladesh": "BD", "indonesia": "ID", "philippines": "PH",
		"thailand": "TH", "vietnam": "VN", "malaysia": "MY", "singapore": "SG",
		"australia": "AU", "canada": "CA", "new zealand": "NZ",
		"saudi arabia": "SA", "uae": "AE", "united arab emirates": "AE",
		"iraq": "IQ", "iran": "IR", "israel": "IL", "lebanon": "LB",
	}
}

// ─── Pagination response builder ──────────────────────────────────────────────

func buildPaginatedResponse(profiles []*models.Profile, total int, f models.ProfileFilter, basePath string) gin.H {
	totalPages := 0
	if f.Limit > 0 {
		totalPages = (total + f.Limit - 1) / f.Limit
	}

	self := fmt.Sprintf("%s?page=%d&limit=%d", basePath, f.Page, f.Limit)
	var next, prev any
	if f.Page < totalPages {
		next = fmt.Sprintf("%s?page=%d&limit=%d", basePath, f.Page+1, f.Limit)
	}
	if f.Page > 1 {
		prev = fmt.Sprintf("%s?page=%d&limit=%d", basePath, f.Page-1, f.Limit)
	}

	return gin.H{
		"status":      "success",
		"page":        f.Page,
		"limit":       f.Limit,
		"total":       total,
		"total_pages": totalPages,
		"links": gin.H{
			"self": self,
			"next": next,
			"prev": prev,
		},
		"data": profiles,
	}
}
