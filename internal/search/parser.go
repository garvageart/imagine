package search

import (
	"regexp"
	"strings"
	"time"
)

// DateRange holds the parsed date range
type DateRange struct {
	Min time.Time
	Max time.Time
}

// SearchCriteria holds the parsed search parameters
type SearchCriteria struct {
	Text      []string
	DateRange DateRange
	Filters   map[string]string
}

// Regex to capture key:value pairs.
// Group 1: Key ([a-zA-Z0-9_]+)
// Group 2: Value ("..." or '...' or unquoted)
var filterRegex = regexp.MustCompile(`([a-zA-Z0-9_]+):(".*?"|'.*?'|[^"\s]+)`)

// ParseQuery parses a raw query string into a structured SearchCriteria
func ParseQuery(input string) SearchCriteria {
	criteria := SearchCriteria{
		Text:    make([]string, 0),
		Filters: make(map[string]string),
	}

	// 1. Find all filter matches
	matches := filterRegex.FindAllStringSubmatch(input, -1)
	for _, match := range matches {
		// match[0] is the full match (e.g. "rating:5")
		// match[1] is the key (e.g. "rating")
		// match[2] is the value (e.g. "5", "'foo'", "\"bar\"")

		key := strings.ToLower(match[1])
		value := match[2]

		// Remove quotes if present
		if len(value) >= 2 {
			if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
				(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
				value = value[1 : len(value)-1]
			}
		}

		// Handle Date Filters
		switch key {
		case "after":
			if t, err := parseDate(value); err == nil {
				criteria.DateRange.Min = t
			}
		case "before":
			if t, err := parseDate(value); err == nil {
				criteria.DateRange.Max = t
			}
		case "date":
			if t, err := parseDate(value); err == nil {
				criteria.DateRange.Min = t
				criteria.DateRange.Max = t.Add(24 * time.Hour)
			}
		default:
			criteria.Filters[key] = value
		}
	}

	// 2. Remove filters from input to get remaining text
	// We replace each match with a space to prevent accidentally merging surrounding words
	cleanedInput := filterRegex.ReplaceAllString(input, " ")

	// 3. Extract remaining text (tokens)
	tokens := strings.FieldsSeq(cleanedInput)
	for t := range tokens {
		t = strings.TrimSpace(t)
		// Trim quotes from tokens if they were part of a phrase that got split
		t = strings.Trim(t, "\"'")
		if t != "" {
			criteria.Text = append(criteria.Text, t)
		}
	}

	return criteria
}

func parseDate(value string) (time.Time, error) {
	// DD-MM-YYYY
	return time.Parse("02-01-2006", value)
}
