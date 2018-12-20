package main

import (
	"go/ast"
	"go/token"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestAdd(t *testing.T) {
	tests := map[string]struct {
		field    *ast.Field
		expected *ast.Field
	}{
		"embedded struct": {
			field: &ast.Field{
				Type: &ast.Ident{
					Name: "EmbeddedStruct",
				},
			},
			expected: &ast.Field{
				Type: &ast.Ident{
					Name: "EmbeddedStruct",
				},
				Tag: &ast.BasicLit{
					Kind:  token.STRING,
					Value: "`yaml:\"embeddedStruct,omitempty\"`",
				},
			},
		},
		"no tag": {
			field: &ast.Field{
				Names: []*ast.Ident{&ast.Ident{Name: "FieldName"}},
			},
			expected: &ast.Field{
				Names: []*ast.Ident{&ast.Ident{Name: "FieldName"}},
				Tag: &ast.BasicLit{
					Kind:  token.STRING,
					Value: "`yaml:\"fieldName,omitempty\"`",
				},
			},
		},
		"tag exists": {
			field: &ast.Field{
				Names: []*ast.Ident{&ast.Ident{Name: "FieldName"}},
				Tag: &ast.BasicLit{
					Kind:  token.STRING,
					Value: "`json:\"field_name,omitempty\"`",
				},
			},
			expected: &ast.Field{
				Names: []*ast.Ident{&ast.Ident{Name: "FieldName"}},
				Tag: &ast.BasicLit{
					Kind:  token.STRING,
					Value: "`json:\"field_name,omitempty\" yaml:\"fieldName,omitempty\"`",
				},
			},
		},
	}
	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			if err := add(test.field); err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if diff := cmp.Diff(test.expected, test.field); diff != "" {
				t.Errorf("result differs: (-want +got)\n%s", diff)
			}
		})
	}
}

func TestCamelCase(t *testing.T) {
	tests := []struct {
		str      string
		expected string
	}{
		{
			str:      "",
			expected: "",
		},
		{
			str:      "a",
			expected: "a",
		},
		{
			str:      "A",
			expected: "a",
		},
		{
			str:      "AAAAA",
			expected: "aAAAA",
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.str, func(t *testing.T) {
			if got, expected := camelCase(test.str), test.expected; got != expected {
				t.Errorf(`expected "%s" but got "%s"`, expected, got)
			}
		})
	}
}
