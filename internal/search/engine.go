package search

import (
	"fmt"
	"imagine/internal/entities"
	"regexp"
	"strings"

	"gorm.io/gorm"
)

type Engine struct{}

func NewEngine() *Engine {
	return &Engine{}
}

// Apply applies the search criteria to the Image query
func (e *Engine) Apply(db *gorm.DB, criteria SearchCriteria) *gorm.DB {
	query := db.Model(&entities.Image{})

	// 1. Text Search (Name OR Description OR Keywords)
	if len(criteria.Text) > 0 {
		term := "%" + strings.Join(criteria.Text, " ") + "%"
		// Postgres ILIKE for case-insensitive search
		// Note: accessing JSONB array elements as text for searching might require different syntax depending on exact requirement
		// For keywords (JSON array), check if the array contains the value or convert to text.
		// Here we attempt a broad text match.
		query = query.Where(
			"name ILIKE ? OR description ILIKE ? OR image_metadata->>'keywords' ILIKE ?",
			term, term, term,
		)
	}

	// 2. Metadata Filters
	if val, ok := criteria.Filters["rating"]; ok {
		op, num := parseOperator(val)
		// Cast JSONB string to numeric for comparison
		query = query.Where(fmt.Sprintf("(image_metadata->>'rating')::numeric %s ?", op), num)
	}

	// 3. EXIF Filters
	exifKeys := []string{"iso", "f_number", "make", "model", "f", "aperture"} // Added 'f' and 'aperture' to be explicitly handled
	for _, key := range exifKeys {
		if val, ok := criteria.Filters[key]; ok {
			dbKey := key
			if key == "f" || key == "aperture" { // Map 'f' and 'aperture' to 'f_number' in DB
				dbKey = "f_number"
			}
			query = query.Where(fmt.Sprintf("exif->>'%s' = ?", dbKey), val)
		}
	}

	// Aspect Ratio / Orientation
	if val, ok := criteria.Filters["orientation"]; ok {
		switch strings.ToLower(val) {
		case "landscape":
			query = query.Where("width > height")
		case "portrait":
			query = query.Where("height > width")
		case "square":
			query = query.Where("width = height")
		}
	}

	// File type
	if val, ok := criteria.Filters["ext"]; ok {
		query = query.Where("image_metadata->>'file_type' = ?", val)
	}
	if val, ok := criteria.Filters["type"]; ok {
		query = query.Where("image_metadata->>'file_type' = ?", val)
	}

	// 4. User Filters (e.g. owner:john)
	if user, ok := criteria.Filters["owner"]; ok {
		query = query.Joins("JOIN users ON users.uid = images.owner_id").
			Where("users.username = ?", user)
	}

	// Status
	if val, ok := criteria.Filters["is"]; ok {
		switch val {
		case "private":
			query = query.Where("private = ?", true)
		case "public":
			query = query.Where("private = ?", false)
		}
	}

	// Favourited
	if val, ok := criteria.Filters["favourited"]; ok {
		switch val {
		case "true":
			query = query.Where("favourited = ?", true)
		case "false":
			query = query.Where("favourited = ? OR favourited IS NULL", false)
		}
	} else if val, ok := criteria.Filters["favorite"]; ok {
		switch val {
		case "true":
			query = query.Where("favourited = ?", true)
		case "false":
			query = query.Where("favourited = ? OR favourited IS NULL", false)
		}
	}

	// 5. Date Filters
	if !criteria.DateRange.Min.IsZero() {
		query = query.Where("taken_at >= ?", criteria.DateRange.Min)
	}
	if !criteria.DateRange.Max.IsZero() {
		query = query.Where("taken_at <= ?", criteria.DateRange.Max)
	}

	return query
}

// ApplyCollections applies the search criteria to the Collection query
func (e *Engine) ApplyCollections(db *gorm.DB, criteria SearchCriteria) *gorm.DB {
	query := db.Model(&entities.Collection{})

	// 1. Text Search (Name OR Description)
	if len(criteria.Text) > 0 {
		term := "%" + strings.Join(criteria.Text, " ") + "%"
		query = query.Where("name ILIKE ? OR description ILIKE ?", term, term)
	}

	// 2. User Filters
	if user, ok := criteria.Filters["owner"]; ok {
		query = query.Joins("JOIN users ON users.uid = collections.owner_id").
			Where("users.username = ?", user)
	}

	// Status
	if val, ok := criteria.Filters["is"]; ok {
		switch val {
		case "private":
			query = query.Where("private = ?", true)
		case "public":
			query = query.Where("private = ?", false)
		}
	}

	// Favourited
	if val, ok := criteria.Filters["favourited"]; ok {
		switch val {
		case "true":
			query = query.Where("favourited = ?", true)
		case "false":
			query = query.Where("favourited = ? OR favourited IS NULL", false)
		}
	} else if val, ok := criteria.Filters["favorite"]; ok {
		switch val {
		case "true":
			query = query.Where("favourited = ?", true)
		case "false":
			query = query.Where("favourited = ? OR favourited IS NULL", false)
		}
	}

	return query
}

// parseOperator extracts operator and value from string like ">=5"
// Default operator is "="
func parseOperator(input string) (string, string) {
	// Simple regex to split operator and number
	re := regexp.MustCompile(`^([<>]=?|=)?(.*)$`)
	matches := re.FindStringSubmatch(input)

	if len(matches) < 3 {
		return "=", input
	}

	op := matches[1]
	val := matches[2]

	switch op {
	case "<", ">", "<=", ">=", "=":
		return op, val
	default:
		return "=", val
	}
}
