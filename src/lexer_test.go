package main

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestLexer(t *testing.T) {
	for _, test := range []struct {
		name     string
		input    string
		expected []item
	}{
		{
			name:  "command with arguments",
			input: "foo bar baz",
			expected: []item{
				str(t, "foo"),
				space(t),
				str(t, "bar"),
				space(t),
				str(t, "baz"),
				eof(t),
			},
		},
		{
			name:  "spaces",
			input: "foo bar            baz ",
			expected: []item{
				str(t, "foo"),
				space(t),
				str(t, "bar"),
				spaceMultiple(t, len("            ")),
				str(t, "baz"),
				space(t),
				eof(t),
			},
		},
		{
			name:  "quoted",
			input: "foo \"bar baz\"",
			expected: []item{
				str(t, "foo"),
				space(t),
				quoted(t, "bar baz"),
				eof(t),
			},
		},
		{
			name:  "slash",
			input: "foo bar/baz",
			expected: []item{
				str(t, "foo"),
				space(t),
				str(t, "bar/baz"),
				eof(t),
			},
		},
		{
			name:  "slash and quotes",
			input: "foo \"bar/ baz\"/qux",
			expected: []item{
				str(t, "foo"),
				space(t),
				quoted(t, "bar/ baz"),
				str(t, "/qux"),
				eof(t),
			},
		},
		{
			name:  "and",
			input: "foo bar&&baz && qux",
			expected: []item{
				str(t, "foo"),
				space(t),
				str(t, "bar"),
				and(t),
				str(t, "baz"),
				space(t),
				and(t),
				space(t),
				str(t, "qux"),
				eof(t),
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, items := lex("test lexer", test.input)
			actual := drain(t, items)

			if diff := cmp.Diff(test.expected, actual); diff != "" {
				t.Errorf("non-zero diff (-expected +actual):\n%s", diff)
			}
		})
	}
}

func str(t *testing.T, s string) item {
	t.Helper()

	return item{
		Type:  itemString,
		Value: s,
	}
}

func quoted(t *testing.T, s string) item {
	t.Helper()

	return item{
		Type:  itemQuotedString,
		Value: fmt.Sprintf("%q", s),
	}
}

func and(t *testing.T) item {
	t.Helper()

	return item{
		Type:  itemAnd,
		Value: "&&",
	}
}

func eof(t *testing.T) item {
	t.Helper()

	return item{
		Type: itemEOF,
	}
}

func space(t *testing.T) item {
	t.Helper()

	return spaceMultiple(t, 1)
}

func spaceMultiple(t *testing.T, count int) item {
	t.Helper()

	return item{
		Type:  itemSpace,
		Value: strings.Repeat(" ", count),
	}
}

func drain(t *testing.T, items chan item) []item {
	t.Helper()

	var out []item

	for item := range items {
		out = append(out, item)
	}

	return out
}
