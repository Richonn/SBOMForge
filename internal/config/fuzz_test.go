package config

import (
	"testing"
)

func FuzzParseBool(f *testing.F) {
	f.Add("true")
	f.Add("false")
	f.Add("1")
	f.Add("0")
	f.Add("yes")
	f.Add("")
	f.Add("TRUE\n")
	f.Add("  true  ")

	f.Fuzz(func(t *testing.T, s string) {
		_ = parseBool(s)
	})
}

func FuzzGetEnvDefault(f *testing.F) {
	f.Add("INPUT_FORMAT", "spdx-json")
	f.Add("", "default")
	f.Add("KEY_WITH_SPECIAL_CHARS!@#", "fallback")

	f.Fuzz(func(t *testing.T, key, def string) {
		_ = getEnvDefault(key, def)
	})
}
