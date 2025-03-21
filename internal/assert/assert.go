package assert

import (
	"reflect"
	"strings"
	"testing"
)

func Equal[T comparable](t *testing.T, actual, expected T) {
	t.Helper()

	if actual != expected {
		t.Errorf("got: %v; want: %v", actual, expected)
	}
}

func DeepEqual(t *testing.T, actual, expected any) {
	t.Helper()

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("got: %v; want: %v", actual, expected)
	}
}

func StringContains(t *testing.T, actual, expectedSubstring string) {
	t.Helper()

	if !strings.Contains(actual, expectedSubstring) {
		t.Errorf("got: %q; expected to contain: %q", actual, expectedSubstring)
	}
}
func NilError(t *testing.T, actual error) {
	t.Helper()

	if actual != nil {
		t.Errorf("got: %v; expected: nil", actual)
	}
}

func NotEmpty(t *testing.T, actual any) {
	t.Helper()

	if actual == nil {
		t.Errorf("expected non-empty value, but got: <nil>")
		return
	}

	v := reflect.ValueOf(actual)
	kind := v.Kind()

	switch kind {
	case reflect.String, reflect.Array, reflect.Slice, reflect.Map, reflect.Chan:
		if v.Len() == 0 {
			t.Errorf("expected non-empty value, but got an empty %s", kind.String())
		}
	default:
		zeroValue := reflect.Zero(reflect.TypeOf(actual)).Interface()
		if reflect.DeepEqual(actual, zeroValue) {
			t.Errorf("expected non-empty (non-zero) value, but got: <%#v>", actual)
		}
	}
}
