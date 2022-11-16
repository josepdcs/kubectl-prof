package lists

import "strings"

func TrimSpace(list []string) []string {
	for i := range list {
		list[i] = strings.TrimSpace(list[i])
	}
	return list
}
