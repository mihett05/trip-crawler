package utils

import "strings"

// BuildPrefixGroups универсальная функция рекурсивной группировки по префиксу.
func BuildPrefixGroups[T any](items []T, extractName func(T) string, currentPrefixLen, maxGroupSize int) map[string][]T {
	groups := make(map[string][]T)

	for _, item := range items {
		name := extractName(item)
		runes := []rune(name)
		pfx := name

		if len(runes) >= currentPrefixLen {
			pfx = string(runes[:currentPrefixLen])
		}

		pfx = strings.ToLower(pfx)
		groups[pfx] = append(groups[pfx], item)
	}

	finalGroups := make(map[string][]T)

	for pfx, group := range groups {
		hasLongerWords := false
		for _, item := range group {
			if len([]rune(extractName(item))) > currentPrefixLen {
				hasLongerWords = true
				break
			}
		}

		if len(group) > maxGroupSize && hasLongerWords {
			subGroups := BuildPrefixGroups(group, extractName, currentPrefixLen+1, maxGroupSize)
			for subPfx, subGroup := range subGroups {
				finalGroups[subPfx] = subGroup
			}
		} else {
			finalGroups[pfx] = group
		}
	}

	return finalGroups
}
