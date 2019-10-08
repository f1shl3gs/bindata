package main

import (
	"fmt"
	"strings"
)

func transferFunc(transfers []string) func(string) string {
	if len(transfers) == 0 {
		return func(s string) string {
			return s
		}
	}

	m := make(map[string]string)
	for _, transfer := range transfers {
		fmt.Println("transfer", transfer)
		fields := strings.Split(transfer, ":")
		if len(fields[0]) == 0 {
			continue
		}

		m[fields[0]] = fields[1]
	}

	return func(s string) string {
		for key, value := range m {
			if strings.Contains(s, key) {
				return strings.ReplaceAll(s, key, value)
			}
		}

		return s
	}
}
