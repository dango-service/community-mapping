package mapping

import (
	"regexp"
	"strings"
	"unicode"
)

var puncRegex = regexp.MustCompile(`[^\p{L}\p{N}\s]+`)
var spacesRegex = regexp.MustCompile(`\s+`)

// CleanTitle lowercases, removes punctuation, and collapses whitespace.
func CleanTitle(s string) string {
	s = strings.ToLower(s)
	// Remove media annotations
	s = strings.ReplaceAll(s, "(manga)", "")
	s = strings.ReplaceAll(s, "(manhwa)", "")
	s = strings.ReplaceAll(s, "(manhua)", "")
	s = strings.ReplaceAll(s, "(novel)", "")
	s = strings.ReplaceAll(s, "[webtoon]", "")
	s = puncRegex.ReplaceAllString(s, " ")
	s = spacesRegex.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

// TokenizeTitle splits clean text into distinct meaningful word tokens.
func TokenizeTitle(s string) map[string]bool {
	clean := CleanTitle(s)
	tokens := make(map[string]bool)
	for _, word := range strings.Fields(clean) {
		if len(word) > 1 || unicode.IsDigit(rune(word[0])) {
			tokens[word] = true
		}
	}
	return tokens
}

// TitleSimilarity computes the Jaccard index between two titles (0.0 to 1.0).
func TitleSimilarity(a, b string) float64 {
	cleanA := CleanTitle(a)
	cleanB := CleanTitle(b)
	if cleanA == "" || cleanB == "" {
		return 0.0
	}
	if cleanA == cleanB {
		return 1.0
	}

	tokensA := TokenizeTitle(a)
	tokensB := TokenizeTitle(b)

	if len(tokensA) == 0 || len(tokensB) == 0 {
		return 0.0
	}

	intersection := 0
	for t := range tokensA {
		if tokensB[t] {
			intersection++
		}
	}

	union := len(tokensA)
	for t := range tokensB {
		if !tokensA[t] {
			union++
		}
	}

	if union == 0 {
		return 0.0
	}
	return float64(intersection) / float64(union)
}

// ExtractYear parses a 4-digit year from an ISO date string (e.g. "2018-03-04" -> 2018).
func ExtractYear(dateStr string) int {
	dateStr = strings.TrimSpace(dateStr)
	if len(dateStr) >= 4 {
		var year int
		if _, err := stringToInt(dateStr[:4], &year); err == nil && year >= 1900 && year <= 2100 {
			return year
		}
	}
	return 0
}

func stringToInt(s string, out *int) (bool, error) {
	var val int
	for _, c := range s {
		if c < '0' || c > '9' {
			return false, nil
		}
		val = val*10 + int(c-'0')
	}
	*out = val
	return true, nil
}
