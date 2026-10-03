package main

import (
	"fmt"
	"strings"
)

type statement struct {
	text string // comments removed, whitespace collapsed
	line int    // 1-based line of its first token
}

// splitStatements splits src on top-level semicolons, so plpgsql bodies and quoted text never end a statement.
func splitStatements(src string) []statement {
	var out []statement
	var cur strings.Builder
	line, start := 1, 0
	flush := func() {
		if t := strings.Join(strings.Fields(cur.String()), " "); t != "" {
			out = append(out, statement{t, start})
		}
		cur.Reset()
		start = 0
	}
	emit := func(s string) {
		if start == 0 && strings.TrimSpace(s) != "" {
			start = line
		}
		cur.WriteString(s)
	}
	for i := 0; i < len(src); {
		c := src[i]
		var n int // bytes consumed by a quoted token or comment
		switch {
		case c == '-' && strings.HasPrefix(src[i:], "--"):
			n = strings.IndexByte(src[i:], '\n')
			if n < 0 {
				n = len(src) - i
			}
			line += strings.Count(src[i:i+n], "\n")
			cur.WriteByte(' ')
			i += n
			continue
		case c == '/' && strings.HasPrefix(src[i:], "/*"):
			n = blockComment(src[i:])
			line += strings.Count(src[i:i+n], "\n")
			cur.WriteByte(' ')
			i += n
			continue
		case c == '\'':
			n = quoted(src[i:], '\'', i > 0 && (src[i-1] == 'E' || src[i-1] == 'e') && (i < 2 || !identChar(src[i-2])))
		case c == '"':
			n = quoted(src[i:], '"', false)
		case c == '$' && (i == 0 || !identChar(src[i-1])):
			n = dollarQuoted(src[i:])
		case c == ';':
			flush()
			i++
			continue
		}
		if n == 0 {
			n = 1
		}
		emit(src[i : i+n])
		line += strings.Count(src[i:i+n], "\n")
		i += n
	}
	flush()
	return out
}

// PostgreSQL nests block comments, so the first */ does not always end one.
func blockComment(s string) int {
	depth := 0
	for i := 0; i < len(s)-1; i++ {
		switch s[i : i+2] {
		case "/*":
			depth++
			i++
		case "*/":
			depth--
			i++
			if depth == 0 {
				return i + 1
			}
		}
	}
	return len(s)
}

// quoted returns the length of the quoted token at s; a doubled quote escapes itself.
func quoted(s string, q byte, backslash bool) int {
	for i := 1; i < len(s); i++ {
		switch {
		case backslash && s[i] == '\\':
			i++
		case s[i] == q && i+1 < len(s) && s[i+1] == q:
			i++
		case s[i] == q:
			return i + 1
		}
	}
	return len(s)
}

// Returning 0 for a non-tag lets a positional parameter such as $1 pass through as ordinary text.
func dollarQuoted(s string) int {
	end := strings.IndexByte(s[1:], '$')
	if end < 0 {
		return 0
	}
	tag := s[:end+2]
	for j, r := range tag[1 : len(tag)-1] {
		if !(r == '_' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || j > 0 && r >= '0' && r <= '9') {
			return 0
		}
	}
	body := strings.Index(s[len(tag):], tag)
	if body < 0 {
		return len(s)
	}
	return len(tag) + body + len(tag)
}

func identChar(c byte) bool {
	return c == '_' || c == '$' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9'
}

var txControl = []string{"BEGIN", "START TRANSACTION", "COMMIT", "END", "ROLLBACK", "ABORT", "PREPARE TRANSACTION"}

// checkWrap requires one BEGIN … COMMIT around the file, since golang-migrate runs each file outside a transaction.
func checkWrap(file, src string) []violation {
	stmts := splitStatements(src)
	if len(stmts) == 0 {
		return []violation{{7, "wrap", file, 0, "empty: wrap the statements in BEGIN; … COMMIT;"}}
	}
	var vs []violation
	first, last := stmts[0], stmts[len(stmts)-1]
	if !strings.EqualFold(first.text, "BEGIN") {
		vs = append(vs, violation{7, "wrap", file, first.line, "the first statement must be BEGIN"})
	}
	if len(stmts) == 1 || !strings.EqualFold(last.text, "COMMIT") {
		vs = append(vs, violation{7, "wrap", file, last.line, "the last statement must be COMMIT"})
	}
	if len(stmts) < 3 {
		return vs
	}
	for _, s := range stmts[1 : len(stmts)-1] {
		if kw, ok := endsTransaction(s.text); ok {
			vs = append(vs, violation{7, "wrap", file, s.line, fmt.Sprintf("%s inside BEGIN … COMMIT leaves the migration half applied", kw)})
		}
	}
	return vs
}

func endsTransaction(text string) (string, bool) {
	up := strings.ToUpper(text) + " "
	if strings.HasPrefix(up, "ROLLBACK TO ") {
		return "", false
	}
	for _, kw := range txControl {
		if strings.HasPrefix(up, kw+" ") {
			return kw, true
		}
	}
	return "", false
}
