package util

import (
	"regexp"
	"strings"
)

const (
	StartsWithNumberREString = "^[0-9].*$"
)

var (
	StartsWithNumberRE = regexp.MustCompile(StartsWithNumberREString)
)

func StringCoalesce(args ...string) string {
	for _, str := range args {
		if strings.TrimSpace(str) != "" {
			return str
		}
	}
	return ""
}

func StartsWithNumber(str string) bool {
	return StartsWithNumberRE.MatchString(str)
}
