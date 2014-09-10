package regexp

import (
	"regexp"
)

func QuoteMeta(s string) string { return regexp.QuoteMeta(s) }
