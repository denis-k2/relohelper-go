package main

import (
	"net/url"
	"reflect"
	"testing"
)

// TestParseInclude tests the include query parameter parser.
func TestParseInclude(t *testing.T) {
	t.Run("empty include", func(t *testing.T) {
		got, err := parseInclude(url.Values{}, newIncludeSet("country"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 0 {
			t.Fatalf("expected empty set, got %v", got)
		}
	})

	t.Run("normalize and dedupe", func(t *testing.T) {
		qs := url.Values{"include": []string{" country,NUMBEO_COST, country "}}
		got, err := parseInclude(qs, newIncludeSet("country", "numbeo_cost"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !got.Has("country") || !got.Has("numbeo_cost") || len(got) != 2 {
			t.Fatalf("unexpected include set: %v", got)
		}
	})

	t.Run("unsupported value", func(t *testing.T) {
		qs := url.Values{"include": []string{"country,foo"}}
		_, err := parseInclude(qs, newIncludeSet("country"))
		if err == nil {
			t.Fatal("expected validation error")
		}
	})

	t.Run("empty token", func(t *testing.T) {
		qs := url.Values{"include": []string{"country, ,numbeo_cost"}}
		_, err := parseInclude(qs, newIncludeSet("country", "numbeo_cost"))
		if err == nil {
			t.Fatal("expected validation error")
		}
	})
}

// TestParseIDsInt64 tests parsing and validation for integer ids query parameter.
func TestParseIDsInt64(t *testing.T) {
	t.Run("not present", func(t *testing.T) {
		got, present, err := parseIDsInt64(url.Values{}, "ids", 200)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if present {
			t.Fatal("expected present=false")
		}
		if got != nil {
			t.Fatalf("expected nil ids, got %v", got)
		}
	})

	t.Run("parse normalize dedupe", func(t *testing.T) {
		qs := url.Values{"ids": []string{"1, 2,2,3 "}}
		got, present, err := parseIDsInt64(qs, "ids", 200)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !present {
			t.Fatal("expected present=true")
		}
		want := []int64{1, 2, 3}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

	t.Run("present but empty", func(t *testing.T) {
		qs := url.Values{"ids": []string{""}}
		_, present, err := parseIDsInt64(qs, "ids", 200)
		if !present {
			t.Fatal("expected present=true")
		}
		if err == nil {
			t.Fatal("expected validation error")
		}
	})

	t.Run("reject non positive", func(t *testing.T) {
		qs := url.Values{"ids": []string{"0,2"}}
		_, _, err := parseIDsInt64(qs, "ids", 200)
		if err == nil {
			t.Fatal("expected validation error")
		}
	})

	t.Run("reject non integer", func(t *testing.T) {
		qs := url.Values{"ids": []string{"1,a"}}
		_, _, err := parseIDsInt64(qs, "ids", 200)
		if err == nil {
			t.Fatal("expected validation error")
		}
	})

	t.Run("reject empty token", func(t *testing.T) {
		qs := url.Values{"ids": []string{"1,,2"}}
		_, _, err := parseIDsInt64(qs, "ids", 200)
		if err == nil {
			t.Fatal("expected validation error")
		}
	})

	t.Run("enforce max", func(t *testing.T) {
		qs := url.Values{"ids": []string{"1,2,3"}}
		_, _, err := parseIDsInt64(qs, "ids", 2)
		if err == nil {
			t.Fatal("expected validation error")
		}
	})

	t.Run("empty key is invalid", func(t *testing.T) {
		_, _, err := parseIDsInt64(url.Values{}, "", 2)
		if err == nil {
			t.Fatal("expected validation error")
		}
	})
}

// TestParseIDsString tests parsing and validation for string ids query parameter.
func TestParseIDsString(t *testing.T) {
	t.Run("not present", func(t *testing.T) {
		got, present, err := parseIDsString(url.Values{}, "ids", 200)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if present {
			t.Fatal("expected present=false")
		}
		if got != nil {
			t.Fatalf("expected nil ids, got %v", got)
		}
	})

	t.Run("normalize uppercase dedupe", func(t *testing.T) {
		qs := url.Values{"ids": []string{"rus, usa ,RUS"}}
		got, present, err := parseIDsString(qs, "ids", 200)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !present {
			t.Fatal("expected present=true")
		}
		want := []string{"RUS", "USA"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

	t.Run("present but empty", func(t *testing.T) {
		qs := url.Values{"ids": []string{""}}
		_, present, err := parseIDsString(qs, "ids", 200)
		if !present {
			t.Fatal("expected present=true")
		}
		if err == nil {
			t.Fatal("expected validation error")
		}
	})

	t.Run("reject empty token", func(t *testing.T) {
		qs := url.Values{"ids": []string{"RUS,,USA"}}
		_, _, err := parseIDsString(qs, "ids", 200)
		if err == nil {
			t.Fatal("expected validation error")
		}
	})

	t.Run("enforce max", func(t *testing.T) {
		qs := url.Values{"ids": []string{"RUS,USA,CAN"}}
		_, _, err := parseIDsString(qs, "ids", 2)
		if err == nil {
			t.Fatal("expected validation error")
		}
	})

	t.Run("empty key is invalid", func(t *testing.T) {
		_, _, err := parseIDsString(url.Values{}, "", 2)
		if err == nil {
			t.Fatal("expected validation error")
		}
	})
}
