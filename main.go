package main

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/alecthomas/kingpin"
	"github.com/fatih/structtag"
	"github.com/pkg/errors"
)

var (
	app      = kingpin.New("goprotoyamltag", "Go tool to add YAML tag into structs generated from Protocol Buffers.")
	filename = app.Flag("filename", "Target filename.").Short('f').Required().String()
	write    = app.Flag("write", "Write result to (source) file instead of stdout.").Short('w').Default("false").Bool()

	ignorePrefix = "XXX_"
)

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))
	if err := realMain(*filename); err != nil {
		log.Fatal(err)
	}
}

func realMain(filepath string) error {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filepath, nil, parser.AllErrors)
	if err != nil {
		return errors.Wrap(err, "failed to parse")
	}

	var inspectErr error
	ast.Inspect(f, func(n ast.Node) bool {
		if inspectErr != nil {
			return false
		}
		switch x := n.(type) {
		case *ast.TypeSpec:
			if !x.Name.IsExported() {
				return false
			}
			if s, ok := x.Type.(*ast.StructType); ok {
				for _, field := range s.Fields.List {
					inspectErr = add(field)
				}
			}
		}
		return true
	})
	if inspectErr != nil {
		return inspectErr
	}

	var w io.Writer = os.Stdout
	if *write {
		file, err := os.OpenFile(filepath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
		if err != nil {
			return errors.Wrap(err, "failed to write result")
		}
		w = file
		defer file.Close()
	}
	if err := format.Node(w, fset, f); err != nil {
		return errors.Wrap(err, "failed to write result")
	}

	return nil
}

func add(field *ast.Field) error {
	var name string
	if len(field.Names) == 0 {
		// embedded struct field
		i, ok := field.Type.(*ast.Ident)
		if !ok {
			return nil
		}
		name = i.Name
	} else {
		name = field.Names[0].Name
	}

	if field.Tag == nil {
		field.Tag = &ast.BasicLit{
			Kind: token.STRING,
		}
	}
	tagVal := field.Tag.Value
	if tagVal != "" {
		v, err := strconv.Unquote(field.Tag.Value)
		if err != nil {
			return err
		}
		tagVal = v
	}
	tags, err := structtag.Parse(tagVal)
	if err != nil {
		return err
	}

	if strings.HasPrefix(name, ignorePrefix) {
		tags.Set(&structtag.Tag{
			Key:  "yaml",
			Name: "-",
		})
	} else {
		tags.Set(&structtag.Tag{
			Key:     "yaml",
			Name:    camelCase(name),
			Options: []string{"omitempty"},
		})
	}
	field.Tag.Value = fmt.Sprintf("`%s`", tags.String())

	return nil
}

func camelCase(str string) string {
	if str == "" {
		return str
	}
	str = strings.ToLower(string(str[0])) + str[1:]
	return str
}
