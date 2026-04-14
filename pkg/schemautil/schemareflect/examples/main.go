package main

import (
	"context"
	"encoding/json"
	"go/ast"
	"log"
	"os"
	"reflect"
	"strings"

	"github.com/google/jsonschema-go/jsonschema"
	kubemetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/act3-ai/go-common/pkg/astutil"
	"github.com/act3-ai/go-common/pkg/schemautil/schemareflect"
	"github.com/act3-ai/go-common/pkg/schemautil/schemareflect/examples/cinema"
)

func main() {
	if err := mainE(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func mainE(ctx context.Context) error {
	info, err := astutil.LoadPackageInfo(ctx, []string{"./..."})
	if err != nil {
		return err
	}

	opts := &schemareflect.Options{}
	opts.PackageInfo = info
	opts.SetXOrder = true
	opts.Namer = func(t reflect.Type) string {
		switch t.PkgPath() {
		case "k8s.io/apimachinery/pkg/apis/meta/v1":
			return "kubemetav1." + t.Name()
		default:
			return schemareflect.DefaultNamer(t)
		}
	}
	// Update the comment formatter to remove Kubernetes style directive comments
	// like "// +optional"
	opts.CommentFormatter = func(comment *ast.CommentGroup) string {
		w := strings.Builder{}
		desc := comment.Text()
		for line := range strings.Lines(desc) {
			// Skip lines starting with "+"
			// Kubernetes types use this for code generation directives
			if strings.HasPrefix(line, "+") {
				continue
			}
			w.WriteString(line)
		}
		// Trim space around the entire output, there may have been newlines
		// before the "+" directives
		return strings.TrimSpace(w.String())
	}
	// Add a schema provider that provides the actual schema for the Kubernetes duration type
	opts.SchemaProviders = append(opts.SchemaProviders, func(t reflect.Type) *jsonschema.Schema {
		switch t {
		case typeMetaDuration:
			// Provide schema for
			return &jsonschema.Schema{
				Type:        "string",
				Description: `A duration string is a possibly signed sequence of decimal numbers, each with optional fraction and a unit suffix, such as "300ms", "-1.5h" or "2h45m". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".`,
			}
		default:
			return nil
		}
	})

	r := schemareflect.NewReflector(opts)

	schema, err := r.GenerateSchemaForType(reflect.TypeFor[cinema.Cinema]())
	if err != nil {
		return err
	}

	schema.Schema = "https://json-schema.org/draft/2020-12/schema"

	e := json.NewEncoder(os.Stdout)
	e.SetEscapeHTML(false)
	e.SetIndent("", "  ")
	return e.Encode(schema)
}

var (
	typeMetaDuration = reflect.TypeFor[kubemetav1.Duration]()
)
