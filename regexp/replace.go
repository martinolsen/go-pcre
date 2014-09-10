package regexp

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

func (re *Regexp) Expand(dst []byte, template []byte, src []byte, match []int) []byte { panic("TODO") }

func (re *Regexp) ExpandString(dst []byte, template string, src string, match []int) []byte {
	return re.Expand(dst, []byte(template), []byte(src), match)
}

func (re *Regexp) ReplaceAll(src, repl []byte) []byte {
	return re.replaceAll(src, func(dst []byte, match []int) []byte {
		return re.expand(dst, string(repl), src, match)
	})
}

func (re *Regexp) ReplaceAllString(src, repl string) string {
	return string(re.ReplaceAll([]byte(src), []byte(repl)))
}

func (re *Regexp) ReplaceAllLiteral(src, repl []byte) []byte {
	return re.replaceAll(src, func(dst []byte, _ []int) []byte {
		return append(dst, repl...)
	})
}

func (re *Regexp) expand(dst []byte, template string, src []byte, match []int) []byte {
	for len(template) > 0 {
		i := strings.Index(template, "$")
		if i < 0 {
			break
		}
		dst = append(dst, template[:i]...)
		template = template[i:]
		if len(template) > 1 && template[1] == '$' {
			// Treat $$ as $.
			dst = append(dst, '$')
			template = template[2:]
			continue
		}
		name, num, rest, ok := extract(template)
		if !ok {
			// Malformed; treat $ as raw text.
			dst = append(dst, '$')
			template = template[1:]
			continue
		}
		template = rest
		if num >= 0 {
			if 2*num+1 < len(match) && match[2*num] >= 0 {
				dst = append(dst, src[match[2*num]:match[2*num+1]]...)
			}
		} else {
			for i, namei := range re.SubexpNames() {
				if name == namei && 2*i+1 < len(match) && match[2*i] >= 0 {
					dst = append(dst, src[match[2*i]:match[2*i+1]]...)
					break
				}
			}
		}
	}
	return append(dst, template...)
}

// extract returns the name from a leading "$name" or "${name}" in str.
// If it is a number, extract returns num set to that number; otherwise num = -1.
func extract(str string) (name string, num int, rest string, ok bool) {
	if len(str) < 2 || str[0] != '$' {
		return
	}
	brace := false
	if str[1] == '{' {
		brace = true
		str = str[2:]
	} else {
		str = str[1:]
	}
	i := 0
	for i < len(str) {
		rune, size := utf8.DecodeRuneInString(str[i:])
		if !unicode.IsLetter(rune) && !unicode.IsDigit(rune) && rune != '_' {
			break
		}
		i += size
	}
	if i == 0 {
		// empty name is not okay
		return
	}
	name = str[:i]
	if brace {
		if i >= len(str) || str[i] != '}' {
			// missing closing brace
			return
		}
		i++
	}

	// Parse number.
	num = 0
	for i := 0; i < len(name); i++ {
		if name[i] < '0' || '9' < name[i] || num >= 1e8 {
			num = -1
			break
		}
		num = num*10 + int(name[i]) - '0'
	}
	// Disallow leading zeros.
	if name[0] == '0' && len(name) > 1 {
		num = -1
	}

	rest = str[i:]
	ok = true
	return
}

func (re *Regexp) replaceAll(src []byte, repl func([]byte, []int) []byte) []byte {
	var dst []byte
	locs := re.FindAllSubmatchIndex(src, -1)
	var srci int
	for i := 0; i < len(locs); i++ {
		dst = append(dst, src[srci:locs[i][0]]...)
		dst = repl(dst, locs[i])
		srci = locs[i][1]
	}
	if srci < len(src) {
		dst = append(dst, src[srci:]...)
	}
	return dst
}

func (re *Regexp) ReplaceAllLiteralString(src, repl string) string {
	return string(re.ReplaceAllLiteral([]byte(src), []byte(repl)))
}

func (re *Regexp) ReplaceAllFunc(src []byte, repl func([]byte) []byte) []byte {
	return re.replaceAll(src, func(dst []byte, match []int) []byte {
		return append(dst, repl(src[match[0]:match[1]])...)
	})
}

func (re *Regexp) ReplaceAllStringFunc(src string, repl func(string) string) string {
	return string(re.replaceAll([]byte(src), func(dst []byte, match []int) []byte {
		return append(dst, []byte(repl(string(src[match[0]:match[1]])))...)
	}))
}
