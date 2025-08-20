package utils

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"
)

func CleanString(s string) string {
	s = strings.TrimPrefix(s, "\uFEFF")
	s = strings.TrimSpace(s)
	return strings.Map(func(r rune) rune {
		if unicode.IsGraphic(r) {
			return r
		}
		return -1
	}, s)
}

func ParseTimeString(s string) (time.Time, error) {
	// RFC3339
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	// "2006-01-02 15:04:05"
	if t, err := time.Parse("2006-01-02 15:04:05", s); err == nil {
		return t, nil
	}
	// epoch seconds
	if sec, err := strconv.ParseInt(s, 10, 64); err == nil {
		return time.Unix(sec, 0), nil
	}

	return time.Time{}, fmt.Errorf("unsupported time format: %q", s)
}
