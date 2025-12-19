package search

import (
	"reflect"
	"testing"
	"time"
)

func TestParseQuery(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected SearchCriteria
	}{
		{
			name:  "Basic text only",
			input: "hello world",
			expected: SearchCriteria{
				Text:    []string{"hello", "world"},
				Filters: map[string]string{},
			},
		},
		{
			name:  "Basic filter only",
			input: "rating:5",
			expected: SearchCriteria{
				Text: []string{},
				Filters: map[string]string{
					"rating": "5",
				},
			},
		},
		{
			name:  "Mixed text and filter",
			input: "sunset rating:5 beach",
			expected: SearchCriteria{
				Text: []string{"sunset", "beach"},
				Filters: map[string]string{
					"rating": "5",
				},
			},
		},
		{
			name:  "Quoted value",
			input: "title:\"auckland park\" park",
			expected: SearchCriteria{
				Text: []string{"park"},
				Filters: map[string]string{
					"title": "auckland park",
				},
			},
		},
		{
			name:  "Single quoted value",
			input: "tag:'maboneng precinct'",
			expected: SearchCriteria{
				Text: []string{},
				Filters: map[string]string{
					"tag": "maboneng precinct",
				},
			},
		},
		{
			name:  "Comparison operator",
			input: "rating:>=4",
			expected: SearchCriteria{
				Text: []string{},
				Filters: map[string]string{
					"rating": ">=4",
				},
			},
		},
		{
			name:  "Complex mix",
			input: "  party narowbi   is:public  rating:5   \"johannesburg\" ",
			expected: SearchCriteria{
				Text: []string{"party", "narowbi", "johannesburg"}, // Quotes in text are split by Fields currently
				Filters: map[string]string{
					"is":     "public",
					"rating": "5",
				},
			},
		},
		{
			name:  "Date filters",
			input: "after:01-01-2023 before:31-12-2023",
			expected: SearchCriteria{
				Text: []string{},
				DateRange: DateRange{
					Min: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					Max: time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC),
				},
				Filters: map[string]string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseQuery(tt.input)
			
			if len(got.Filters) == 0 && len(tt.expected.Filters) == 0 {
				got.Filters = map[string]string{}
				tt.expected.Filters = map[string]string{}
			}

			if !reflect.DeepEqual(got.Filters, tt.expected.Filters) {
				t.Errorf("ParseQuery() Filters = %v, want %v", got.Filters, tt.expected.Filters)
			}

			// sort text slices for comparison as order might vary slightly depending on split implementation details
			// although fields preserves order
			if !reflect.DeepEqual(got.Text, tt.expected.Text) {
				t.Errorf("ParseQuery() Text = %v, want %v", got.Text, tt.expected.Text)
			}

			if !got.DateRange.Min.Equal(tt.expected.DateRange.Min) {
				t.Errorf("ParseQuery() DateRange.Min = %v, want %v", got.DateRange.Min, tt.expected.DateRange.Min)
			}
			if !got.DateRange.Max.Equal(tt.expected.DateRange.Max) {
				t.Errorf("ParseQuery() DateRange.Max = %v, want %v", got.DateRange.Max, tt.expected.DateRange.Max)
			}
		})
	}
}