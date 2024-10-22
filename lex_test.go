package templex_test

import (
	"bytes"
	"testing"

	"github.com/yang-zzhong/templex"
)

func TestLexSimple(t *testing.T) {
	bs := "调度在{{.tasks.first.started_at}}开始"
	should := []templex.Token{
		{Type: templex.TokenConst, Raw: []byte("调度在")},
		{Type: templex.TokenVar, Raw: []byte("{{.tasks.first.started_at}}")},
		{Type: templex.TokenConst, Raw: []byte("开始")},
	}
	tokens, err := templex.Lex(bytes.NewReader([]byte(bs)))
	if err != nil {
		t.Fatal(err)
	}
	if len(tokens) != len(should) {
		t.Fatalf("len(tokens) != len(should)")
	}
	for i, s := range should {
		if s.Type != tokens[i].Type {
			t.Fatalf("s.Type != tokens[i].Type")
		}
		if !bytes.Equal(s.Raw, tokens[i].Raw) {
			t.Fatalf("s.Value != tokens[i].Value")
		}
	}
}

func TestLexLoop(t *testing.T) {
	bs := "{{#for .tasks}}调度在{{.items.started_at}}开始{{#end}}"
	should := []templex.Token{
		{Type: templex.TokenLoopBegin, Raw: []byte("{{#for .tasks}}")},
		{Type: templex.TokenConst, Raw: []byte("调度在")},
		{Type: templex.TokenVar, Raw: []byte("{{.items.started_at}}")},
		{Type: templex.TokenConst, Raw: []byte("开始")},
		{Type: templex.TokenLoopEnd, Raw: []byte("{{#end}}")},
		{Type: templex.TokenConst, Raw: []byte("")},
	}
	tokens, err := templex.Lex(bytes.NewReader([]byte(bs)))
	if err != nil {
		t.Fatal(err)
	}
	if len(tokens) != len(should) {
		t.Fatalf("len(tokens) != len(should)")
	}
	for i, s := range should {
		if s.Type != tokens[i].Type {
			t.Fatalf("s.Type != tokens[i].Type")
		}
		if !bytes.Equal(s.Raw, tokens[i].Raw) {
			t.Fatalf("s.Value != tokens[i].Value")
		}
	}
}
