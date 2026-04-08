package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"reflect"

	"github.com/act3-ai/go-common/pkg/astutil"
	"github.com/act3-ai/go-common/pkg/schemautil/schemagen"
	"github.com/act3-ai/go-common/pkg/schemautil/schemagen/examples/cinema"
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

	gen := schemagen.NewGenerator()
	gen.PackageInfo = info
	gen.SetXOrder = true
	gen.Namer = func(t reflect.Type) string {
		switch t.PkgPath() {
		case "k8s.io/apimachinery/pkg/apis/meta/v1":
			return "kubemetav1." + t.Name()
		default:
			return schemagen.DefaultNamer(t)
		}
	}

	schema, err := gen.GenerateSchemaForType(reflect.TypeFor[cinema.Cinema]())
	if err != nil {
		return err
	}

	schema.Schema = "https://json-schema.org/draft/2020-12/schema"

	e := json.NewEncoder(os.Stdout)
	e.SetEscapeHTML(false)
	e.SetIndent("", "  ")
	return e.Encode(schema)
}
