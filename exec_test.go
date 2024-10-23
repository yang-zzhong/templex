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

func TestExec(t *testing.T) {
	input := []templex.Statement{{
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
	context := map[string]any{
		"tasks": []map[string]any{{
			"started_at": 1001001,
		}, {
			"started_at": 1001002,
		}},
	}
	var buf bytes.Buffer
	if err := templex.Exec(input, context, &buf); err != nil {
		t.Fatal(err)
	}
	if buf.String() != "调度在1001001开始调度在1001002开始" {
		t.Fatal("exec failed")
	}
	t.Log(buf.String())
}

func TestExecWithWhitespace(t *testing.T) {
	input := []templex.Statement{{
		Type: templex.StmtFor,
		Statements: []templex.Statement{
			{Type: templex.StmtRenderConst, Value: []byte("调度在")},
			{Type: templex.StmtRenderVar, Value: []byte("{{ .__value__.started_at }}")},
			{Type: templex.StmtRenderConst, Value: []byte("开始")},
		},
		Value: []byte("{{ #for .tasks }}"),
	}, {
		Type: templex.StmtRenderConst,
	}}
	context := map[string]any{
		"tasks": []map[string]any{{
			"started_at": 1001001,
		}, {
			"started_at": 1001002,
		}},
	}
	var buf bytes.Buffer
	if err := templex.Exec(input, context, &buf); err != nil {
		t.Fatal(err)
	}
	if buf.String() != "调度在1001001开始调度在1001002开始" {
		t.Fatal("exec failed")
	}
	t.Log(buf.String())
}

func TestExecUpperCaseFor(t *testing.T) {
	input := []templex.Statement{{
		Type: templex.StmtFor,
		Statements: []templex.Statement{
			{Type: templex.StmtRenderConst, Value: []byte("调度在")},
			{Type: templex.StmtRenderVar, Value: []byte("{{ .__value__.started_at }}")},
			{Type: templex.StmtRenderConst, Value: []byte("开始")},
		},
		Value: []byte("{{ #FOR .tasks }}"),
	}, {
		Type: templex.StmtRenderConst,
	}}
	context := map[string]any{
		"tasks": []map[string]any{{
			"started_at": 1001001,
		}, {
			"started_at": 1001002,
		}},
	}
	var buf bytes.Buffer
	if err := templex.Exec(input, context, &buf); err != nil {
		t.Fatal(err)
	}
	if buf.String() != "调度在1001001开始调度在1001002开始" {
		t.Fatal("exec failed")
	}
	t.Log(buf.String())
}
