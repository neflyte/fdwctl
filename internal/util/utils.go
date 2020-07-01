package util

import "strings"

func StringCoalesce(args ...string) string {
	for _, str := range args {
		if strings.TrimSpace(str) != "" {
			return str
		}
	}
	return ""
}
