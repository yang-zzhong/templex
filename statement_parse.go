// Copyright (c) 2024 Yang,Zhong
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package templex

import "errors"

const (
	StmtRenderConst = iota
	StmtRenderVar
	StmtFor
)

type Statement struct {
	Type       int
	Value      []byte
	Statements []Statement
}

var (
	ErrUnexpectedEnd = errors.New("unexpected end")
)

func Parse(tokens []Token) (statements []Statement, err error) {
	statements, _, err = parse(tokens)
	return
}

func parse(tokens []Token, expect ...int) (statements []Statement, consumed int, err error) {
	if len(tokens) == 0 {
		if len(expect) > 0 {
			return nil, 0, ErrUnexpectedEnd
		}
		return
	}
	if len(expect) > 0 && tokens[0].Type == expect[0] {
		consumed += 1
		return
	}
	token := tokens[0]
	switch token.Type {
	case TokenConst:
		statements = append(statements, Statement{Type: StmtRenderConst, Value: token.Raw})
		consumed += 1
	case TokenVar:
		statements = append(statements, Statement{Type: StmtRenderVar, Value: token.Raw})
		consumed += 1
	case TokenLoopBegin:
		var forStmts []Statement
		var csd int
		consumed += 1
		if forStmts, csd, err = parse(tokens[1:], TokenLoopEnd); err != nil {
			return
		}
		consumed += csd
		statements = append(statements, Statement{Type: StmtFor, Value: token.Raw, Statements: forStmts})
	}
	var stmts []Statement
	var csd int
	if stmts, csd, err = parse(tokens[consumed:], expect...); err != nil {
		return
	}
	consumed += csd
	statements = append(statements, stmts...)
	return
}
