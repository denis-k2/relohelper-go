package main

import (
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

type IncludeSet map[string]struct{}

func newIncludeSet(values ...string) IncludeSet {
	set := make(IncludeSet, len(values))
	for _, value := range values {
		set[value] = struct{}{}
	}

	return set
}

func (s IncludeSet) Has(value string) bool {
	_, ok := s[value]
	return ok
}

func validateAllowedQueryParams(qs url.Values, allowed IncludeSet) error {
	if len(qs) == 0 {
		return nil
	}

	keys := make([]string, 0, len(qs))
	for key := range qs {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		if !allowed.Has(key) {
			return fmt.Errorf("unknown query parameter %q", key)
		}
	}

	return nil
}

func parseInclude(qs url.Values, allowed IncludeSet) (IncludeSet, error) {
	raw := strings.TrimSpace(qs.Get("include"))
	if raw == "" {
		return newIncludeSet(), nil
	}

	include := newIncludeSet()
	for _, token := range strings.Split(raw, ",") {
		item := strings.ToLower(strings.TrimSpace(token))
		if item == "" {
			return nil, fmt.Errorf("include contains an empty value")
		}
		if !allowed.Has(item) {
			return nil, fmt.Errorf("include contains unsupported value %q", item)
		}

		include[item] = struct{}{}
	}

	return include, nil
}

func parseIDsInt64(qs url.Values, key string, max int) ([]int64, bool, error) {
	if key == "" {
		return nil, false, fmt.Errorf("query parameter key must be provided")
	}
	if max <= 0 {
		max = 100
	}

	if !qs.Has(key) {
		return nil, false, nil
	}
	raw := strings.TrimSpace(qs.Get(key))
	if raw == "" {
		return nil, true, fmt.Errorf("%s contains an empty value", key)
	}

	ids := make([]int64, 0)
	seen := make(map[int64]struct{})

	for _, token := range strings.Split(raw, ",") {
		item := strings.TrimSpace(token)
		if item == "" {
			return nil, true, fmt.Errorf("%s contains an empty value", key)
		}

		id, err := strconv.ParseInt(item, 10, 64)
		if err != nil {
			return nil, true, fmt.Errorf("%s contains non-integer value %q", key, item)
		}
		if id <= 0 {
			return nil, true, fmt.Errorf("%s must contain only positive integers", key)
		}
		if _, ok := seen[id]; ok {
			continue
		}

		seen[id] = struct{}{}
		ids = append(ids, id)
		if len(ids) > max {
			return nil, true, fmt.Errorf("%s cannot contain more than %d unique values", key, max)
		}
	}

	return ids, true, nil
}

func parseIDsString(qs url.Values, key string, max int) ([]string, bool, error) {
	if key == "" {
		return nil, false, fmt.Errorf("query parameter key must be provided")
	}
	if max <= 0 {
		max = 100
	}

	if !qs.Has(key) {
		return nil, false, nil
	}
	raw := strings.TrimSpace(qs.Get(key))
	if raw == "" {
		return nil, true, fmt.Errorf("%s contains an empty value", key)
	}

	ids := make([]string, 0)
	seen := make(map[string]struct{})

	for _, token := range strings.Split(raw, ",") {
		item := strings.ToUpper(strings.TrimSpace(token))
		if item == "" {
			return nil, true, fmt.Errorf("%s contains an empty value", key)
		}
		if _, ok := seen[item]; ok {
			continue
		}

		seen[item] = struct{}{}
		ids = append(ids, item)
		if len(ids) > max {
			return nil, true, fmt.Errorf("%s cannot contain more than %d unique values", key, max)
		}
	}

	return ids, true, nil
}
