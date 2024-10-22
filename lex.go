package templex

import (
	"bytes"
	"errors"
	"io"
)

const (
	TokenConst = iota
	TokenVar
	TokenLoopBegin
	TokenLoopEnd
)

type Token struct {
	Type int
	Raw  []byte
}

const (
	stateConst = iota
	stateVarStart
	stateVarMaybe
	stateVarBegin
	stateVar
	stateVarEndBegin
	stateLoopBegin
	stateLoopF
	stateLoopFo
	stateLoopFor
	stateLoopE
	stateLoopEn
	stateLoopEnd
	stateLoopEndBegin
	stateLoopSeekVar
	stateLoopVarBegin
	stateLoopVar
	stateLoopVarEnd
	stateVarEnd
)

func Lex(r io.Reader) ([]Token, error) {
	p := tokenlexer{}
	tokens := []Token{}
	for {
		var token Token
		err, end := p.Token(r, &token)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
		if end {
			return tokens, nil
		}
	}
}

// 变量 ^\{\{\.[a-zA-Z]+[\.0-9a-zA-Z\_]+\}\}
//
//	调度在{{.tasks.first.started_at}}开始, {{.tasks.last.ended_at}}结束，共 {{.task_count}} 共调度{{.schedule_inst_count}}个实例，成功{{.success_count}}个，失败{{.failed_count}}个
//	 其中
// {{#for .tasks}}
//
//	{{.item.include_inst_kind}}: {{.item.schedule_inst_count}}个, 成功{{.item.success_count}}个,失败{{.item.failed_count}}个，取消{{.item.cancel_count}}
//
// {{#end}}

// ['调度在', {{.tasks.first.started_at}},',', {{}.tasks.last_ended_at}]

type tokenlexer struct {
	state int
	stash []byte
	maybe []byte
	unget *Token
}

func (p *tokenlexer) Token(r io.Reader, token *Token) (err error, end bool) {
	if p.unget != nil {
		*token = *p.unget
		p.unget = nil
		return nil, false
	}
	bs := make([]byte, 1)
	var l int
	for {
		l, err = r.Read(bs)
		if err != nil {
			if errors.Is(err, io.EOF) {
				p.state = stateConst
				p.stash = append(p.stash, p.maybe...)
				*token = Token{Type: TokenConst, Raw: p.stash}
				p.stash = nil
				p.maybe = nil
				return nil, true
			}
			return
		}
		if l == 0 {
			continue
		}
		switch p.state {
		case stateConst:
			p.handleStart(bs[0])
		case stateVarStart:
			p.handleVarStart(bs[0])
		case stateVarMaybe:
			p.handleVarMaybe(bs[0])
		case stateVarBegin:
			p.handleVarBegin(bs[0])
		case stateVar:
			p.handleVar(bs[0])
		case stateVarEndBegin:
			tokens, get := p.handleVarEndBegin(bs[0])
			if !get {
				continue
			}
			*token = tokens[0]
			if len(tokens) > 1 {
				p.unget = &tokens[1]
			}
			return nil, false
		case stateLoopBegin:
			p.handleLoopBegin(bs[0])
		case stateLoopF:
			p.handleLoop(bs[0], 'o', stateLoopFo)
		case stateLoopFo:
			p.handleLoop(bs[0], 'r', stateLoopFor)
		case stateLoopE:
			p.handleLoop(bs[0], 'n', stateLoopEn)
		case stateLoopEn:
			p.handleLoop(bs[0], 'd', stateLoopEnd)
		case stateLoopEnd:
			p.handleLoopEnd(bs[0])
		case stateLoopEndBegin:
			tokens, get := p.handleLoopEndBegin(bs[0])
			if !get {
				continue
			}
			*token = tokens[0]
			if len(tokens) > 1 {
				p.unget = &tokens[1]
			}
			return nil, false
		case stateLoopFor:
			p.handleLoopFor(bs[0])
		case stateLoopSeekVar:
			p.handleLoopSeekVar(bs[0])
		case stateLoopVarBegin:
			p.handleLoopVarBegin(bs[0])
		case stateLoopVar:
			p.handleLoopVar(bs[0])
		case stateLoopVarEnd:
			tokens, get := p.handleLoopVarEnd(bs[0])
			if !get {
				continue
			}
			*token = tokens[0]
			if len(tokens) > 1 {
				p.unget = &tokens[1]
			}
			return nil, false
		}
	}
}

func (p *tokenlexer) resetToConst(b byte) {
	p.state = stateConst
	p.stash = append(p.stash, p.maybe...)
	p.stash = append(p.stash, b)
	p.maybe = nil
}

func (p *tokenlexer) handleLoopBegin(b byte) {
	switch b {
	case 'f', 'F':
		p.maybe = append(p.maybe, b)
		p.state = stateLoopF
	case 'e', 'E':
		p.maybe = append(p.maybe, b)
		p.state = stateLoopE
	default:
		p.resetToConst(b)
	}
}

func (p *tokenlexer) handleLoop(b byte, expect byte, state int) {
	switch b {
	case expect, bytes.ToUpper([]byte{expect})[0]:
		p.maybe = append(p.maybe, b)
		p.state = state
	default:
		p.resetToConst(b)
	}
}

func (p *tokenlexer) handleLoopFor(b byte) {
	if p.isWhitespace(b) {
		p.state = stateLoopSeekVar
		p.maybe = append(p.maybe, b)
		return
	}
	p.resetToConst(b)
}

func (p *tokenlexer) handleLoopEnd(b byte) {
	if p.isWhitespace(b) {
		p.maybe = append(p.maybe, b)
		return
	}
	if b == '}' {
		p.maybe = append(p.maybe, b)
		p.state = stateLoopEndBegin
		return
	}
	p.resetToConst(b)
}

func (p *tokenlexer) handleLoopEndBegin(b byte) ([]Token, bool) {
	if b == '}' {
		ret := []Token{}
		if len(p.stash) > 2 {
			ret = append(ret, Token{Type: TokenConst, Raw: p.stash})
			p.stash = []byte{}
		}
		ret = append(ret, Token{Type: TokenLoopEnd, Raw: append(p.maybe, b)})
		p.maybe = []byte{}
		p.state = stateConst
		return ret, true
	}
	p.stash = append(p.stash, p.maybe...)
	p.stash = append(p.stash, b)
	p.maybe = []byte{}
	p.state = stateConst
	return nil, false
}

func (p *tokenlexer) handleLoopSeekVar(b byte) {
	if p.isWhitespace(b) {
		p.maybe = append(p.maybe, b)
		return
	}
	if b == '.' {
		p.maybe = append(p.maybe, b)
		p.state = stateLoopVarBegin
		return
	}
	p.resetToConst(b)
}

func (p *tokenlexer) handleLoopVarBegin(b byte) {
	if b == '_' || b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z' {
		p.state = stateLoopVar
		p.maybe = append(p.maybe, b)
		return
	}
	p.resetToConst(b)
}

func (p *tokenlexer) handleLoopVar(b byte) {
	if b == '.' {
		p.state = stateLoopVarBegin
		p.maybe = append(p.maybe, b)
		return
	}
	if b == '}' {
		p.state = stateLoopVarEnd
		p.maybe = append(p.maybe, b)
		return
	}
	if b == '_' || b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z' || b >= '0' && b <= '9' {
		p.state = stateLoopVar
		p.maybe = append(p.maybe, b)
		return
	}
	p.resetToConst(b)
}

func (p *tokenlexer) handleLoopVarEnd(b byte) ([]Token, bool) {
	if b == '}' {
		ret := []Token{}
		if len(p.stash) > 2 {
			ret = append(ret, Token{Type: TokenConst, Raw: p.stash})
			p.stash = []byte{}
		}
		ret = append(ret, Token{Type: TokenLoopBegin, Raw: append(p.maybe, b)})
		p.maybe = []byte{}
		p.state = stateConst
		return ret, true
	}
	p.stash = append(p.stash, p.maybe...)
	p.stash = append(p.stash, b)
	p.maybe = []byte{}
	p.state = stateConst
	return nil, false
}

func (p *tokenlexer) handleStart(b byte) {
	switch b {
	case '{':
		p.maybe = append(p.maybe, b)
		p.state = stateVarStart
		return
	}
	p.stash = append(p.stash, b)
}

func (p *tokenlexer) handleVarStart(b byte) {
	switch b {
	case '{':
		p.state = stateVarMaybe
		p.maybe = append(p.maybe, b)
		return
	}
	p.stash = append(p.stash, b)
}

func (p *tokenlexer) handleVarMaybe(b byte) {
	switch b {
	case '.':
		p.maybe = append(p.maybe, '.')
		p.state = stateVarBegin
	case '#':
		p.maybe = append(p.maybe, '#')
		p.state = stateLoopBegin
	default:
		p.resetToConst(b)
	}
}

func (p *tokenlexer) handleVarBegin(b byte) {
	if b >= 'a' && b <= 'z' || b >= 'A' || b <= 'Z' || b >= '0' && b <= '9' || b == '_' {
		p.state = stateVar
		p.maybe = append(p.maybe, b)
		return
	}
	p.resetToConst(b)
}

func (p *tokenlexer) handleVar(b byte) {
	if b == '.' {
		p.maybe = append(p.maybe, b)
		p.state = stateVarBegin
	} else if b == '}' {
		p.maybe = append(p.maybe, b)
		p.state = stateVarEndBegin
	} else if b >= 'a' && b <= 'z' || b >= 'A' || b <= 'Z' || b >= '0' && b <= '9' || b == '_' {
		p.maybe = append(p.maybe, b)
	} else {
		p.resetToConst(b)
	}
}

func (p *tokenlexer) handleVarEndBegin(b byte) ([]Token, bool) {
	if b == '}' {
		ret := []Token{}
		if len(p.stash) > 0 {
			ret = append(ret, Token{Type: TokenConst, Raw: p.stash})
			p.stash = []byte{}
		}
		ret = append(ret, Token{Type: TokenVar, Raw: append(p.maybe, b)})
		p.maybe = []byte{}
		p.state = stateConst
		return ret, true
	}
	p.stash = append(p.stash, p.maybe...)
	p.stash = append(p.stash, b)
	p.maybe = []byte{}
	p.state = stateConst
	return nil, false
}

func (p *tokenlexer) isWhitespace(b byte) bool {
	return b == ' '
}
