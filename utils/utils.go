package utils

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

var specialBytes [16]byte

func init() {
	for _, b := range []byte(`\.+*?()|[]{}^$-!=#_<>~`) {
		specialBytes[b%16] |= 1 << (b / 16)
	}
}

// SanitizeTelegramString - sanitize string to escape specials charatcters
func SanitizeTelegramString(rawString string) string {
	var sanitizer strings.Builder
	sanitizer.Grow(32)
	urlRegexp := regexp.MustCompile(`\[[^[]+\]\(https?:\/\/[^\s]+\)`)
	urlAliasRegexp := regexp.MustCompile(`\[(.+?)\]\((.*)\)`)
	urls := urlRegexp.FindAllString(rawString, -1)
	// No urls found
	if len(urls) == 0 {
		return quoteMeta(rawString, specialBytes)
	}

	tail := rawString
	for i, url := range urls {
		urlAlias := urlAliasRegexp.FindStringSubmatch(url)[1]
		sanitizedUrl := strings.ReplaceAll(url, urlAlias, quoteMeta(urlAlias, specialBytes))
		prefix := strings.Split(tail, url)[0]
		tail = strings.Split(tail, url)[1]
		// Add last part
		if i == len(urls)-1 {
			fmt.Fprint(&sanitizer, quoteMeta(prefix, specialBytes), sanitizedUrl, quoteMeta(tail, specialBytes))
		} else {
			fmt.Fprint(&sanitizer, quoteMeta(prefix, specialBytes), sanitizedUrl)
		}
	}
	return sanitizer.String()
}

func quoteMeta(s string, special [16]byte) string {
	// A byte loop is correct because all metacharacters are ASCII.
	var i int
	for i = 0; i < len(s); i++ {
		if s[i] < utf8.RuneSelf && special[s[i]%16]&(1<<(s[i]/16)) != 0 {
			break
		}
	}
	// No meta characters found, so return original string.
	if i >= len(s) {
		return s
	}

	b := make([]byte, 2*len(s)-i)
	copy(b, s[:i])
	j := i
	for ; i < len(s); i++ {
		if s[i] < utf8.RuneSelf && special[s[i]%16]&(1<<(s[i]/16)) != 0 {
			b[j] = '\\'
			j++
		}
		b[j] = s[i]
		j++
	}
	return string(b[:j])
}
