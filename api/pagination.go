package main

import "fmt"

type Pagination struct {
	Href   string `json:"href"`
	Prev   string `json:"prev"`
	Next   string `json:"next"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
	Count  int    `json:"count"`
}

func CreatePaginationURL(pathname string, limit, offset int) string {
	return fmt.Sprintf("%s?limit=%d&offset=%d", pathname, limit, offset)
}
