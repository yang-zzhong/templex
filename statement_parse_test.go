// Copyright (c) 2024 Yang,Zhong
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package templex_test

import (
	"bytes"
	"testing"

	"github.com/yang-zzhong/templex"
)

func stmtEqual(stmt1, stmt2 templex.Statement) bool {
	if stmt1.Type != stmt2.Type {
		return false
	}
	if !bytes.Equal(stmt1.Value, stmt2.Value) {
		return false
	}
	if len(stmt1.Statements) != len(stmt2.Statements) {
		return false
	}
	for i := 0; i < len(stmt1.Statements); i++ {
		if !stmtEqual(stmt1.Statements[i], stmt2.Statements[i]) {
			return false
		}
	}
	return true
}

func TestParseStatement(t *testing.T) {
	input := []templex.Token{
		{Type: templex.TokenLoopBegin, Raw: []byte("{{#for .tasks}}")},
		{Type: templex.TokenConst, Raw: []byte("调度在")},
		{Type: templex.TokenVar, Raw: []byte("{{.__value__.started_at}}")},
		{Type: templex.TokenConst, Raw: []byte("开始")},
		{Type: templex.TokenLoopEnd, Raw: []byte("{{#end}}")},
		{Type: templex.TokenConst, Raw: []byte("")},
	}
	stmts, err := templex.Parse(input)
	if err != nil {
		t.Fatal(err)
	}
	should := []templex.Statement{{
		Type: templex.StmtFor,
		Statements: []templex.Statement{
			{Type: templex.StmtRenderConst, Value: []byte("调度在")},
			{Type: templex.StmtRenderVar, Value: []byte("{{.__value__.started_at}}")},
			{Type: templex.StmtRenderConst, Value: []byte("开始")},
		},
		Value: []byte("{{#for .tasks}}"),
	}, {
		Type: templex.StmtRenderConst,
	}}

	for i := 0; i < len(stmts); i++ {
		if !stmtEqual(stmts[i], should[i]) {
			t.Fatalf("stmts[%d] != should[%d]", i, i)
		}
	}
}
