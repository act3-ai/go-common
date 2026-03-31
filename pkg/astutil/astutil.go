package astutil

import (
	"cmp"
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"iter"

	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/go/packages"
)

// PackageInfo stores Go package information.
type PackageInfo struct {
	Pkgs []*packages.Package
	Fset *token.FileSet
	// Inspectors []*inspector.Inspector
	inspectors map[string]*inspector.Inspector // per-package inspectors
}

// LoadPackageInfo loads all packages matching any of the given patterns and
// returns a PackageInfo ready for comment and AST queries.
//
// The patterns follow the same syntax as `go list`
// (e.g. "./...", "std", "github.com/foo/bar/...").
func LoadPackageInfo(ctx context.Context, patterns ...string) (*PackageInfo, error) {
	// Create new FileSet
	fset := token.NewFileSet()

	// Create packages config that will load comments.
	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedSyntax |
			packages.NeedTypes |
			packages.NeedTypesInfo,
		Context: ctx,
		Fset:    fset,
		ParseFile: func(fset *token.FileSet, filename string, src []byte) (*ast.File, error) {
			return parser.ParseFile(fset, filename, src, parser.ParseComments)
		},
	}

	// Load the packages matching the patterns
	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		return nil, fmt.Errorf("loading packages: %w", err)
	}

	// Initialize inspectors for each set of files
	inspectors := make([]*inspector.Inspector, 0, len(pkgs))
	for _, pkg := range pkgs {
		if len(pkg.Errors) > 0 {
			return nil, fmt.Errorf("package %s: %v", pkg.PkgPath, pkg.Errors[0])
		}
		inspectors = append(inspectors, inspector.New(pkg.Syntax))
	}

	return &PackageInfo{
		Pkgs:       pkgs,
		Fset:       fset,
		inspectors: make(map[string]*inspector.Inspector, len(pkgs)),
	}, nil
}

// TypeComment produces the ast.CommentGroup for the
func (info *PackageInfo) TypeComment(pkgPath, typeName string) (*ast.CommentGroup, error) {
	typeCursor, _, err := info.getType(pkgPath, typeName)
	if err != nil {
		return nil, err
	}
	return getTypeComment(typeCursor), nil
}

func (info *PackageInfo) FieldComment(pkgPath, typeName, fieldName string) (*ast.CommentGroup, error) {
	typeCursor, _, err := info.getType(pkgPath, typeName)
	if err != nil {
		return nil, err
	}

	for _, field := range allFieldsForTypeSpec(typeCursor) {
		for _, ident := range field.Names {
			if ident.Name == fieldName {
				return getFieldComment(field), nil
			}
		}
	}

	return nil, fmt.Errorf("field %s.%s.%s not found", pkgPath, typeName, fieldName)
}

func (info *PackageInfo) getPackage(pkgPath string) (inspector.Cursor, *packages.Package, error) {
	for _, pkg := range info.Pkgs {
		if pkg.PkgPath == pkgPath {
			if _, ok := info.inspectors[pkgPath]; !ok {
				// Create inspector if not yet initialized
				info.inspectors[pkgPath] = inspector.New(pkg.Syntax)
			}
			return info.inspectors[pkgPath].Root(), pkg, nil
		}
	}
	return inspector.Cursor{}, nil, fmt.Errorf("package %s not found", pkgPath)
}

func (info *PackageInfo) getType(pkgPath, typeName string) (inspector.Cursor, *ast.TypeSpec, error) {
	pkgCursor, _, err := info.getPackage(pkgPath)
	if err != nil {
		return inspector.Cursor{}, nil, err
	}

	// Search all TypeSpecs for matching name
	for typeCursor, typeSpec := range allTypeSpecs(pkgCursor) {
		// Return cursor if found
		if typeSpec.Name.Name == typeName {
			return typeCursor, typeSpec, nil
		}
	}

	return inspector.Cursor{}, nil, fmt.Errorf("type %s.%s not found", pkgPath, typeName)
}

type ExtractedComment struct {
	PkgPath      string            `json:"pkgPath"`
	TypeName     string            `json:"typeName"`
	FieldName    *string           `json:"fieldName,omitzero"`
	Comment      *string           `json:"comment"`
	CommentGroup *ast.CommentGroup `json:"-"`
}

func (info *PackageInfo) AllComments() []ExtractedComment {
	result := []ExtractedComment{}
	for _, pkg := range info.Pkgs {
		pkgCursor, _, _ := info.getPackage(pkg.PkgPath)
		for typeCursor, typeSpec := range allTypeSpecs(pkgCursor) {
			typeComment := getTypeComment(typeCursor)

			// Add type comment
			result = append(result, ExtractedComment{
				PkgPath:      pkg.PkgPath,
				TypeName:     typeSpec.Name.Name,
				FieldName:    nil,
				Comment:      commentGroupText(typeComment),
				CommentGroup: typeComment,
			})
			// Add each field's comment
			for _, field := range allFieldsForTypeSpec(typeCursor) {
				// Add comment to map under all defined names
				fieldComment := getFieldComment(field)
				for _, fieldName := range field.Names {
					result = append(result, ExtractedComment{
						PkgPath:      pkg.PkgPath,
						TypeName:     typeSpec.Name.Name,
						FieldName:    &fieldName.Name,
						Comment:      commentGroupText(fieldComment),
						CommentGroup: fieldComment,
					})
				}
			}
		}
	}
	return result
}

func commentGroupText(cg *ast.CommentGroup) *string {
	if cg == nil {
		return nil
	}
	text := cg.Text()
	return &text
}

func allTypeSpecs(pkgCursor inspector.Cursor) iter.Seq2[inspector.Cursor, *ast.TypeSpec] {
	return func(yield func(inspector.Cursor, *ast.TypeSpec) bool) {
		for typeCursor := range pkgCursor.Preorder((*ast.TypeSpec)(nil)) {
			if !yield(typeCursor, typeCursor.Node().(*ast.TypeSpec)) {
				return
			}
		}
	}
}

func allFieldsForTypeSpec(typeCursor inspector.Cursor) iter.Seq2[inspector.Cursor, *ast.Field] {
	return func(yield func(inspector.Cursor, *ast.Field) bool) {
		for fieldCursor := range typeCursor.Preorder((*ast.Field)(nil)) {
			// Must be a field on a struct
			if !isFieldOnStruct(fieldCursor) {
				continue
			}
			if !yield(fieldCursor, fieldCursor.Node().(*ast.Field)) {
				return
			}
		}
	}
}

func isFieldOnStruct(fieldCursor inspector.Cursor) bool {
	// Direct parent is FieldList, grandparent is StructType.
	grandparent := fieldCursor.Parent().Parent()
	_, ok := grandparent.Node().(*ast.StructType)
	return ok
}

func getTypeComment(typeCursor inspector.Cursor) *ast.CommentGroup {
	if genDecl, ok := typeCursor.Parent().Node().(*ast.GenDecl); ok && genDecl != nil {
		return genDecl.Doc
	}
	return nil
}

func getFieldComment(field *ast.Field) *ast.CommentGroup {
	switch {
	case field == nil:
		return nil
	case field.Doc != nil:
		return field.Doc
	case field.Comment != nil:
		return field.Comment
	default:
		return nil
	}
}

// TypeSpecNodes calls the function f for all TypeSpec nodes found under root.
func TypeSpecNodes(root ast.Node, stack []ast.Node, f func(typeSpec *ast.TypeSpec, stack []ast.Node) bool) {
	ast.PreorderStack(root, stack, func(n ast.Node, stack []ast.Node) bool {
		typeSpec, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}
		return f(typeSpec, stack)
	})
}

// GetTypeSpecComment inspects the stack at an *ast.TypeSpec node to locate the comment
// for the type spec, which is defined on the parent declaration group.
func GetTypeSpecComment(stack []ast.Node) *ast.CommentGroup {
	if len(stack) < 1 {
		return nil
	}
	genDecl, ok := stack[len(stack)-1].(*ast.GenDecl)
	if !ok {
		return nil
	}
	return genDecl.Doc
}

// ExtractComments extracts type and struct field comments from the provided packages.
func ExtractComments(pkgs []*packages.Package) []ExtractedComment {
	result := []ExtractedComment{}
	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			TypeSpecNodes(file, nil, func(typeSpec *ast.TypeSpec, stack []ast.Node) bool {
				// Add type comment
				typeComment := GetTypeSpecComment(stack)
				result = append(result, ExtractedComment{
					PkgPath:      pkg.PkgPath,
					TypeName:     typeSpec.Name.Name,
					FieldName:    nil,
					Comment:      commentGroupText(typeComment),
					CommentGroup: typeComment,
				})

				StructFieldNodes(typeSpec, stack, func(field *ast.Field, stack []ast.Node) bool {
					// Add field comment
					fieldComment := cmp.Or(field.Doc, field.Comment)
					for _, fieldName := range field.Names {
						result = append(result, ExtractedComment{
							PkgPath:      pkg.PkgPath,
							TypeName:     typeSpec.Name.Name,
							FieldName:    &fieldName.Name,
							Comment:      commentGroupText(fieldComment),
							CommentGroup: fieldComment,
						})
					}
					return true
				})

				return true
			})
		}
	}
	return result
}

// StructFieldNodes calls the function f for all struct field nodes found under root.
func StructFieldNodes(root ast.Node, stack []ast.Node, f func(field *ast.Field, stack []ast.Node) bool) {
	ast.PreorderStack(root, stack, func(n ast.Node, stack []ast.Node) bool {
		field, ok := n.(*ast.Field)
		if !ok {
			return true
		}
		if stackIsFieldOnStruct(stack) {
			return f(field, stack)
		}
		return true
	})
}

// Direct parent is FieldList, grandparent is StructType.
func stackIsFieldOnStruct(stack []ast.Node) bool {
	if len(stack) < 2 {
		return false
	}
	if _, ok := stack[len(stack)-1].(*ast.FieldList); !ok {
		return false
	}
	if _, ok := stack[len(stack)-2].(*ast.StructType); !ok {
		return false
	}
	return true
}
