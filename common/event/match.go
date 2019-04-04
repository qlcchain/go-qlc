/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package event

// MatchSimple - finds whether the text matches/satisfies the pattern string.
// supports only '*' wildcard in the pattern.
// considers a file system path as a flat name space.
func MatchSimple(pattern, name string) bool {
	if pattern == "" {
		return name == pattern
	}
	if pattern == "*" {
		return true
	}
	rName := make([]rune, 0, len(name))
	rPattern := make([]rune, 0, len(pattern))
	for _, r := range name {
		rName = append(rName, r)
	}
	for _, r := range pattern {
		rPattern = append(rPattern, r)
	}
	// Does only wildcard '*' match.
	return deepMatchRune(rName, rPattern, true)
}

// Match -  finds whether the text matches/satisfies the pattern string.
// supports  '*' and '?' wildcards in the pattern string.
// unlike path.Match(), considers a path as a flat name space while matching the pattern.
// The difference is illustrated in the example here https://play.golang.org/p/Ega9qgD4Qz .
func Match(pattern, name string) (matched bool) {
	if pattern == "" {
		return name == pattern
	}
	if pattern == "*" {
		return true
	}
	rName := make([]rune, 0, len(name))
	rPattern := make([]rune, 0, len(pattern))
	for _, r := range name {
		rName = append(rName, r)
	}
	for _, r := range pattern {
		rPattern = append(rPattern, r)
	}
	// Does extended wildcard '*' and '?' match.
	return deepMatchRune(rName, rPattern, false)
}

func deepMatchRune(str, pattern []rune, simple bool) bool {
	for len(pattern) > 0 {
		switch pattern[0] {
		default:
			if len(str) == 0 || str[0] != pattern[0] {
				return false
			}
		case '?':
			if len(str) == 0 && !simple {
				return false
			}
		case '*':
			return deepMatchRune(str, pattern[1:], simple) ||
				(len(str) > 0 && deepMatchRune(str[1:], pattern, simple))
		}
		str = str[1:]
		pattern = pattern[1:]
	}
	return len(str) == 0 && len(pattern) == 0
}
