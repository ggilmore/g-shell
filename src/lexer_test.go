package main

import (
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
				{
					Type:  itemString,
					Value: "foo",
				},
				{
					Type:  itemString,
					Value: "bar",
				},
				{
					Type:  itemString,
					Value: "baz",
				},
				{
					Type: itemEOF,
				},
			},
		},
		{
			name:  "spaces",
			input: "foo bar            baz",
			expected: []item{
				{
					Type:  itemString,
					Value: "foo",
				},
				{
					Type:  itemString,
					Value: "bar",
				},
				{
					Type:  itemString,
					Value: "baz",
				},
				{
					Type: itemEOF,
				},
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

func drain(t *testing.T, items chan item) []item {
	t.Helper()

	var out []item

	for item := range items {
		out = append(out, item)
	}

	return out
}
