package parse

import (
	"testing"
)

func TestGetTokenAt(t *testing.T) {
	input := `
const MY_CONST = 4 # IGNORED
FOO 	thing(BLAH)	BAZ // COMMENT
 MyFoo
`
	tests := []struct {
		line     int
		column   int
		expected string
	}{
		{line: 0, column: -1, expected: ""},
		{line: -1, column: 0, expected: ""},
		{line: 0, column: 0, expected: ""},
		{line: 0, column: 1, expected: ""},
		{line: 1, column: 0, expected: "const"},
		{line: 1, column: 2, expected: "const"},
		{line: 1, column: 5, expected: "const"},
		{line: 1, column: 6, expected: "MY_CONST"},
		{line: 1, column: 13, expected: "MY_CONST"},
		{line: 1, column: 14, expected: "MY_CONST"},
		{line: 1, column: 25, expected: ""},
		{line: 1, column: 15, expected: ""},
		{line: 1, column: 18, expected: "4"},
		{line: 2, column: 0, expected: "FOO"},
		{line: 2, column: 3, expected: "FOO"},
		{line: 2, column: 4, expected: ""},
		{line: 2, column: 5, expected: "thing"},
		{line: 2, column: 10, expected: "thing"},
		{line: 2, column: 11, expected: "BLAH"},
		{line: 2, column: 15, expected: "BLAH"},
		{line: 2, column: 19, expected: "BAZ"},
		{line: 2, column: 20, expected: "BAZ"},
		{line: 2, column: 31, expected: ""},
		{line: 2, column: 190, expected: ""},
		{line: 3, column: 3, expected: "MyFoo"},
		{line: 44, column: 0, expected: ""},
	}
	for i, tt := range tests {
		result := GetTokenAt(input, tt.line, tt.column)
		if result != tt.expected {
			t.Errorf("Test Case %d: Expected: '%s', Got: '%s'", i, tt.expected, result)
		}
	}
}
