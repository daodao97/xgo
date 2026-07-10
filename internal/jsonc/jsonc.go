package jsonc

import "encoding/json"

// ToJSON removes JSONC comments and trailing commas so the result can be
// parsed by the standard encoding/json package.
func ToJSON(src []byte) []byte {
	return removeTrailingCommas(stripComments(src))
}

// Valid reports whether data is valid JSON after JSONC cleanup.
func Valid(data []byte) bool {
	return json.Valid(ToJSON(data))
}

func stripComments(src []byte) []byte {
	out := make([]byte, 0, len(src))
	inString := false
	escaped := false

	for i := 0; i < len(src); i++ {
		ch := src[i]

		if inString {
			out = append(out, ch)
			if escaped {
				escaped = false
				continue
			}
			switch ch {
			case '\\':
				escaped = true
			case '"':
				inString = false
			}
			continue
		}

		switch ch {
		case '"':
			inString = true
			out = append(out, ch)
		case '#':
			i = skipLineComment(src, i+1, &out)
		case '/':
			if i+1 >= len(src) {
				out = append(out, ch)
				continue
			}
			switch src[i+1] {
			case '/':
				i = skipLineComment(src, i+2, &out)
			case '*':
				i = skipBlockComment(src, i+2, &out)
			default:
				out = append(out, ch)
			}
		default:
			out = append(out, ch)
		}
	}

	return out
}

func skipLineComment(src []byte, start int, out *[]byte) int {
	for i := start; i < len(src); i++ {
		if src[i] == '\n' {
			*out = append(*out, '\n')
			return i
		}
	}
	return len(src) - 1
}

func skipBlockComment(src []byte, start int, out *[]byte) int {
	for i := start; i < len(src); i++ {
		if src[i] == '\n' {
			*out = append(*out, '\n')
			continue
		}
		if src[i] == '*' && i+1 < len(src) && src[i+1] == '/' {
			return i + 1
		}
	}
	return len(src) - 1
}

func removeTrailingCommas(src []byte) []byte {
	out := make([]byte, 0, len(src))
	inString := false
	escaped := false

	for i := 0; i < len(src); i++ {
		ch := src[i]

		if inString {
			out = append(out, ch)
			if escaped {
				escaped = false
				continue
			}
			switch ch {
			case '\\':
				escaped = true
			case '"':
				inString = false
			}
			continue
		}

		if ch == '"' {
			inString = true
			out = append(out, ch)
			continue
		}

		if ch == ',' {
			next := nextNonSpace(src, i+1)
			if next < len(src) && (src[next] == ']' || src[next] == '}') {
				continue
			}
		}

		out = append(out, ch)
	}

	return out
}

func nextNonSpace(src []byte, start int) int {
	for i := start; i < len(src); i++ {
		switch src[i] {
		case ' ', '\t', '\r', '\n':
			continue
		default:
			return i
		}
	}
	return len(src)
}
