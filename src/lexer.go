package main

import (
	"fmt"
	"unicode"
	"unicode/utf8"
)

/**
ARG = TEXT
PROGRAM = TEXT ARG*
PROGRAM_GROUP = PROGRAM && PROGRAM
**/

type itemType int

const (
	itemString itemType = iota
	itemQuotedString

	itemError

	itemEOF
)

const EOF = -1

type stateFn func(*lexer) stateFn

type lexer struct {
	name string // used for errors

	input string // the string being scanned
	start int    // start position of this item
	pos   int    // current position in the input

	width int       // width of last rune read
	items chan item // channel of scanned items
}

func lex(name, input string) (*lexer, chan item) {
	l := &lexer{
		name:  name,
		input: input,
		items: make(chan item),
	}

	go l.run()
	return l, l.items
}

func (l *lexer) emit(t itemType) {
	l.items <- item{t, l.input[l.start:l.pos]}
	l.start = l.pos
}

func (l *lexer) run() {
	for state := lexText; state != nil; {
		state = state(l)
	}
	close(l.items)
}

func (l *lexer) next() rune {
	if l.pos >= len(l.input) {
		l.width = 0
		return EOF
	}

	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = w

	l.pos += l.width
	return r
}

func (l *lexer) ignore() {
	l.start = l.pos
}

func (l *lexer) backup() {
	l.pos -= l.width
}

func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- item{
		itemError,
		fmt.Sprintf(format, args...),
	}

	return nil
}

// skip over whitespace, emit next tokens
func lexText(l *lexer) stateFn {
	l.width = 0

	r := l.next()

	if isSpace(r) {
		l.ignore()
		return lexText
	}

	if r == EOF {
		l.emit(itemEOF)
		return nil
	}

	if r == '"' {
		return lexQuoted
	}

	l.backup()
	return lexToken

}

func lexToken(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case isAlphaNumeric(r):
			// collect
		default:
			l.backup()
			if !l.atTerminator() {
				return l.errorf("bad character %#U", r)
			}

			l.emit(itemString)
			return lexText
		}
	}
}

func lexQuoted(l *lexer) stateFn {
	for {
		switch l.next() {
		case EOF:
			return l.errorf("unterminated quoted string")
		case '"':
			l.emit(itemQuotedString)
			return lexText
		}
	}
}

func (l *lexer) atTerminator() bool {
	r := l.peek()
	if isSpace(r) {
		return true
	}

	return r == EOF
}

func isAlphaNumeric(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}

func isSpace(r rune) bool {
	return unicode.IsSpace(r)
}

type item struct {
	Type  itemType
	Value string
}

func (i item) String() string {
	switch i.Type {
	case itemEOF:
		return "EOF"
	case itemError:
		return i.Value
	}

	return fmt.Sprintf("%q", i.Value)
}
