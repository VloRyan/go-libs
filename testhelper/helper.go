package testhelper

import (
	"io"
	"os"
	"testing"
	"time"
)

var FixedNow = ParseTime(nil, "2024-06-26T20:33:44Z")

func ParseTime(t *testing.T, value string) time.Time {
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		t.Fatal(err)
	}
	return parsed
}

func Ptr[T any](s T) *T {
	return &s
}

func ReadFile(t *testing.T, name string) string {
	f, err := os.Open(name)
	if err != nil {
		t.Fatal(err)
	}
	defer func(jsonFile *os.File) {
		_ = jsonFile.Close()
	}(f)

	byteValue, err := io.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}
	return string(byteValue)
}
