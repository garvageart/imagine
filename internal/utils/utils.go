package utils

import (
	"encoding/json"
	"log"
	"maps"
	mRand "math/rand"
	"os"
	"reflect"
	"slices"
	"strings"

	"go.uber.org/zap"
)

// A struct to hold default date values useful for formatting Golang dates
type DateFormatValues struct {
	Year        string
	MonthDigit  string
	MonthString string
	Day         string
	Hour        string
	Minute      string
	Second      string
}

var DefaultDateTimeValues = &DateFormatValues{
	Year:        "2006",
	MonthDigit:  "01",
	MonthString: "January",
	Day:         "02",
	Hour:        "15",
	Minute:      "04",
	Second:      "05",
}

const (
	DefaultDateTimeFormat = "02012006_150405"
	VersionFileName = "version.txt"
)

func GetAppVersion() string {
	data, err := os.ReadFile(VersionFileName)
	if err != nil {
		panic(err)
	}

	return strings.Split(string(data), "\n")[0]
}

// Utility function to log and panic if an error occurs
func FailOnError(err error, msg string) {

	log.Panicf("%s: %s", msg, err)
}

func FileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}

	return !info.IsDir()
}

// Converts a map to a slice of zap.Fields
func MapToZapFields(inputMap map[string]any) []zap.Field {
	fields := make([]zap.Field, 0, len(inputMap))

	for key, val := range inputMap {
		fields = append(fields, zap.Any(key, val))
	}

	return fields
}

func MapToSlice(inputMap map[string]any) []any {
	return slices.Collect(maps.Values(inputMap))
}

func IsEnvironment(env string) bool {
	return os.Getenv("env") == env ||
		os.Getenv("ENV") == env ||
		os.Getenv("environment") == env
}

func JsonToMap(jsonString string) (map[string]any, error) {
	result := make(map[string]any)

	err := json.Unmarshal([]byte(jsonString), &result)
	if err != nil {
		return nil, err
	}

	return result, err
}

// From https://gist.github.com/bxcodec/c2a25cfc75f6b21a0492951706bc80b8
func StructToMap(item interface{}) map[string]interface{} {
	res := map[string]interface{}{}
	if item == nil {
		return res
	}

	v := reflect.TypeOf(item)
	reflectValue := reflect.ValueOf(item)
	reflectValue = reflect.Indirect(reflectValue)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	for i := 0; i < v.NumField(); i++ {
		tag := v.Field(i).Tag.Get("json")
		field := reflectValue.Field(i).Interface()
		if tag != "" && tag != "-" {
			if v.Field(i).Type.Kind() == reflect.Struct {
				res[tag] = StructToMap(field)
			} else {
				res[tag] = field
			}
		}
	}

	return res
}

func RandomInt(min, max int) int {
	return min + mRand.Intn(max-min)
}

// StringPtr returns a pointer to the provided string.
func StringPtr(s string) *string { return &s }
