package helpers

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// MicrosecondsStr 将 time.Duration 类型（nano seconds 为单位）
// 输出为小数点后 3 位的 ms （microsecond 毫秒，千分之一秒）
func MicrosecondsStr(elapsed time.Duration) string {
	return fmt.Sprintf("%.3fms", float64(elapsed.Nanoseconds())/1e6)
}

var slugRe = regexp.MustCompile(`[^a-z0-9]+`)

// Slugify converts s into a URL-safe lowercase slug.
// Non-ASCII characters (e.g. Chinese) that produce an empty slug after
// stripping are replaced with a short random suffix so we never return the
// ambiguous string "default".
func Slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = slugRe.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		id, _ := uuid.NewRandom()
		return id.String()[:8]
	}
	return s
}
