package apt

import (
	"strconv"
	"strings"
	"unicode/utf8"
)

type tokenType int

const (
	openParam tokenType = iota
	closeParam
	operator
	constant
)

const eof rune = -1

type token struct {
	typ   tokenType
	value string
}

type lexer struct {
	input             string
	start, pos, width int
	tokens            chan token
}

type stateFunc func(*lexer) stateFunc

func (l *lexer) next() (r rune) {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return r
}

func (l *lexer) backup() {
	l.pos -= l.width
}

func (l *lexer) ignore() {
	l.start = l.pos
}

func (l *lexer) emit(t tokenType) {
	l.tokens <- token{t, l.input[l.start:l.pos]}
	l.start = l.pos
}

func isWhiteSpace(r rune) bool {
	return r == ' ' || r == '\n' || r == '\t' || r == '\r'
}

func isStartOfNumber(r rune) bool {
	return (r >= '0' && r <= '9') || r == '-' || r == '.'
}

func (l *lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

func (l *lexer) acceptRun(valid string) {
	for strings.IndexRune(valid, l.next()) >= 0 {}
	l.backup()
}

func lexNumber(l *lexer) stateFunc {
	l.accept("-.")
	digits := "0123456789"
	l.acceptRun(digits)
	if l.accept(".") {
		l.acceptRun(digits)
	}

	if l.input[l.start:l.pos] == "-" {
		l.emit(operator)
	} else {
		l.emit(constant)
	}

	return determineToken
}

func lexOp(l *lexer) stateFunc {
	l.acceptRun("+-/*abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	l.emit(operator)
	return determineToken
}

func determineToken(l *lexer) stateFunc {
	for {
		switch r := l.next(); {
		case isWhiteSpace(r):
			l.ignore()
		case r == '(':
			l.emit(openParam)
		case r == ')':
			l.emit(closeParam)
		case isStartOfNumber(r):
			return lexNumber
		case r == eof:
			return nil
		default:
			return lexOp
		}
	}
} 

func (l *lexer) run() {
	for state := determineToken; state != nil; {
		state = state(l)
	}
	close(l.tokens)
}

func stringToNode(s string) Node {
	switch s {
	case "+":
		return NewOpPlus()
	case "-":
		return NewOpMinus()
	case "*":
		return NewOpMultiplies()
	case "/":
		return NewOpDivide()
	case "atan2":
		return NewOpAtan2()
	case "atan":
		return NewOpAtan()
	case "sin":
		return NewOpSin()
	case "cos":
		return NewOpCos()
	case "snoise2":
		return NewOpNoise()
	case "turbulence":
		return NewTurbulence()
	case "ceil":
		return NewOpCeil()
	case "fbm":
		return NewOpFbm()
	case "floor":
		return NewOpFloor()
	case "negate":
		return NewOpNegate()
	case "square":
		return NewOpSquare()
	case "abs":
		return NewOpAbs()
	case "x":
		return NewOpX()
	case "y":
		return NewOpY()
	case "picture":
		return NewOpPict()
	default:
		panic("didnt understand what you mean" + s)
	}
}

func parse(tokens chan token, parent Node) Node {
	for {
		token, ok := <- tokens
		if !ok {
			panic("no more tokens")
		}

		switch token.typ {
		case openParam, closeParam:
			continue
		case operator:
			n := stringToNode(token.value)
			n.SetParent(parent)
			for i := range n.GetChildren() {
				n.GetChildren()[i] = parse(tokens, n)
			}
			return n
		case constant:
			n := NewOpConst()
			n.SetParent(parent)
			v, err := strconv.ParseFloat(token.value, 32)
			if err != nil {
				panic(err)
			}
			n.value = float32(v)
			return n
		}
	}
	return nil
}

func BeginLexing(s string) Node {
	l := &lexer{input: s, tokens: make(chan token, 100)}
	go l.run()
	return parse(l.tokens, nil)
}