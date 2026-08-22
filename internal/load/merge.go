package load

import (
	"fmt"

	"niraluli/internal/ast"
)

// MergeFiles concatenates same-package files into one File (imports + decls).
func MergeFiles(files []*ast.File) (*ast.File, []error) {
	if len(files) == 0 {
		return nil, []error{fmt.Errorf("no source files")}
	}
	var errs []error
	var pkg *ast.PackageClause
	var imports []*ast.ImportSpec
	var decls []ast.Decl
	path := ""
	for i, f := range files {
		if f == nil {
			errs = append(errs, fmt.Errorf("nil file at index %d", i))
			continue
		}
		if f.Package == nil || f.Package.Name == nil {
			errs = append(errs, fmt.Errorf("missing package clause"))
			continue
		}
		if pkg == nil {
			pkg = f.Package
		} else if f.Package.Name.Name != pkg.Name.Name {
			errs = append(errs, fmt.Errorf("%d:%d: package name mismatch: got %s, want %s",
				f.Package.Name.Pos().Line, f.Package.Name.Pos().Col, f.Package.Name.Name, pkg.Name.Name))
		}
		if path == "" {
			path = f.Path
		}
		imports = append(imports, f.Imports...)
		decls = append(decls, f.Decls...)
	}
	if len(errs) != 0 {
		return nil, errs
	}
	if pkg == nil {
		return nil, []error{fmt.Errorf("missing package clause")}
	}
	return &ast.File{Path: path, Package: pkg, Imports: imports, Decls: decls}, nil
}
