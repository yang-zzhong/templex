// Copyright (c) 2024 Yang,Zhong
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package templex

import (
	"io"
	"reflect"
	"strconv"
	"strings"

	"github.com/nao1215/nameconv"
	"github.com/spf13/cast"
)

func Exec(stmt []Statement, context map[string]any, w io.Writer) (err error) {
	for _, stmt := range stmt {
		if err = exec(stmt, context, w); err != nil {
			return
		}
	}
	return
}

func exec(stmt Statement, context map[string]any, w io.Writer) (err error) {
	switch stmt.Type {
	case StmtRenderConst:
		_, err = w.Write(stmt.Value)
		return
	case StmtRenderVar:
		key := strings.ReplaceAll(strings.ReplaceAll(string(stmt.Value), "}}", ""), "{{", "")
		val, found := searchValue(reflect.ValueOf(context), strings.Trim(key, " "))
		if !found {
			return
		}
		_, err = w.Write([]byte(cast.ToString(val.Interface())))
		return
	case StmtFor:
		key := strings.ReplaceAll(strings.ReplaceAll(string(stmt.Value), "{{#for", ""), "}}", "")
		val, found := searchValue(reflect.ValueOf(context), strings.Trim(key, " "))
		if !found {
			return
		}
		return execStmtFor(val, context, stmt.Statements, w)
	}
	return
}

func execStmtFor(val reflect.Value, context map[string]any, stmts []Statement, w io.Writer) (err error) {
	switch val.Kind() {
	case reflect.Interface:
		inter := val.Interface()
		if v, ok := inter.(map[string]any); ok {
			return execStmtFor(reflect.ValueOf(v), context, stmts, w)
		} else if v, ok := inter.([]any); ok {
			return execStmtFor(reflect.ValueOf(v), context, stmts, w)
		} else if v, ok := inter.([]map[string]any); ok {
			return execStmtFor(reflect.ValueOf(v), context, stmts, w)
		}
	case reflect.Slice:
		for i := 0; i < val.Len(); i++ {
			context["__value__"] = val.Index(i).Interface()
			context["__key__"] = i
			for _, stmt := range stmts {
				err = exec(stmt, context, w)
				if err != nil {
					return
				}
			}
		}
	case reflect.Map:
		keys := val.MapKeys()
		for _, ke := range keys {
			context["__value__"] = val.MapIndex(ke).Interface()
			context["__key__"] = ke.Interface()
			for _, stmt := range stmts {
				err = exec(stmt, context, w)
				if err != nil {
					return
				}
			}
		}
	}
	return
}

func searchValue(args reflect.Value, key string) (reflect.Value, bool) {
	keys := strings.Split(key, ".")
	for i, k := range keys {
		if k == "" {
			continue
		}
		val, ok := getValue(args, k)
		if !ok {
			return reflect.Value{}, false
		}
		if i == len(keys)-1 {
			return val, true
		}
		args = val
	}
	return reflect.Value{}, false
}

func getValue(args reflect.Value, k string) (reflect.Value, bool) {
	switch args.Type().Kind() {
	case reflect.Struct:
		field := args.FieldByName(nameconv.ToPascalCase(k))
		if field.IsValid() {
			return field, true
		}
		fields := args.Type().NumField()
		for i := 0; i < fields; i++ {
			field := args.Type().Field(i)
			tag := field.Tag.Get("json")
			if tag == "-" {
				continue
			}
			ts := strings.Split(tag, ",")
			if ts[0] == k {
				return args.Field(i), true
			}
		}
		return reflect.Value{}, false
	case reflect.Map:
		keys := args.MapKeys()
		for _, ke := range keys {
			if ke.String() == k {
				return args.MapIndex(ke), true
			}
		}
		return reflect.Value{}, false
	case reflect.Slice:
		index := 0
		if k == "first" {
			index = 0
		} else if k == "last" {
			index = args.Len() - 1
		} else if i, err := strconv.Atoi(k); err == nil {
			index = i
		} else {
			return reflect.Value{}, false
		}
		if index > args.Len() {
			return reflect.Value{}, false
		}
		return args.Index(index), true
	case reflect.Interface:
		inter := args.Interface()
		if v, ok := inter.(map[string]any); ok {
			return getValue(reflect.ValueOf(v), k)
		} else if v, ok := inter.([]any); ok {
			return getValue(reflect.ValueOf(v), k)
		} else if v, ok := inter.([]map[string]any); ok {
			return getValue(reflect.ValueOf(v), k)
		}
		return reflect.Value{}, false
	default:
		return reflect.Value{}, false
	}
}
