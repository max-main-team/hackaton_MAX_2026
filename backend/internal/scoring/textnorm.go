package scoring

import (
	"regexp"
	"strings"
)

var (
	pageNumberRe   = regexp.MustCompile(`(?i)^(?:страница|стр\.?|page)?\s*[-—–|]?\s*\d{1,4}\s*(?:из\s*\d{1,4})?\s*[-—–|]?$`)
	hyphenEndRe    = regexp.MustCompile(`([A-Za-zА-Яа-яЁё])-\s*$`)
	spaceRunRe     = regexp.MustCompile(`[ \t\x{00A0}]{2,}`)
	contactNoiseRe = regexp.MustCompile(`(?im)^\s*(?:телефон|тел\.?|e-?mail|email|почта|github|gitlab|telegram)\s*[:.].{0,120}$`)
	bareEmailRe    = regexp.MustCompile(`^\S+@\S+\.\S+$`)
	barePhoneRe    = regexp.MustCompile(`^\+?[\d\s()\-]{10,20}$`)
)

func NormalizeResumeText(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")

	lines := strings.Split(s, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			if len(out) > 0 && out[len(out)-1] != "" {
				out = append(out, "")
			}
			continue
		}
		if pageNumberRe.MatchString(line) || contactNoiseRe.MatchString(line) ||
			bareEmailRe.MatchString(line) || barePhoneRe.MatchString(line) {
			continue
		}
		if n := len(out); n > 0 && out[n-1] != "" {
			if prev := out[n-1]; hyphenEndRe.MatchString(prev) && joinableHyphen(line) {
				out[n-1] = prev + line
				continue
			}
		}
		out = append(out, line)
	}

	var b strings.Builder
	for _, line := range out {
		if line == "" {
			b.WriteByte('\n')
			continue
		}
		line = spaceRunRe.ReplaceAllString(line, " ")
		b.WriteString(line)
		b.WriteByte('\n')
	}
	return strings.TrimSpace(b.String())
}

func joinableHyphen(next string) bool {
	next = strings.TrimSpace(next)
	if next == "" {
		return false
	}
	first := []rune(next)[0]
	return first >= 'a' && first <= 'z' || first >= 'а' && first <= 'я' || first == 'ё'
}

func TruncateText(s string, max int) string {
	if len(s) <= max {
		return s
	}
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max])
}
