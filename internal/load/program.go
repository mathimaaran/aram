package load

import (
	"fmt"
	"path/filepath"

	"aram/internal/ast"
)

// Package is one Aram package (directory of .aram files).
type Package struct {
	Name  string
	Dir   string
	Files []*ast.File
}

// Program is an entry package plus imported packages (topo-sorted in All).
type Program struct {
	Entry *Package
	All   []*Package // dependencies first, entry last
}

func bindImport(im *ast.ImportSpec, pkgName string) {
	im.Pkg = pkgName
	if im.Alias != nil {
		im.Name = im.Alias.Name
	} else {
		im.Name = pkgName
	}
}

// LoadProgram expands entry args, parses the entry package, and follows கொணர் imports.
func LoadProgram(args []string) (*Program, []error) {
	entryPaths, err := Expand(args)
	if err != nil {
		return nil, []error{err}
	}
	entryFiles, perrs := ParsePaths(entryPaths)
	if len(perrs) != 0 {
		return nil, perrs
	}
	entryMerged, merrs := MergeFiles(entryFiles)
	if len(merrs) != 0 {
		return nil, merrs
	}
	entryDir := commonDir(entryPaths)
	entry := &Package{
		Name:  entryMerged.Package.Name.Name,
		Dir:   entryDir,
		Files: entryFiles,
	}

	byName := map[string]*Package{entry.Name: entry}
	byDir := map[string]*Package{filepath.Clean(entryDir): entry}
	var errs []error
	var order []string

	visiting := map[string]bool{}
	var stack []string
	var visit func(pkg *Package)
	visit = func(pkg *Package) {
		if visiting[pkg.Name] {
			cycle := append(append([]string{}, stack...), pkg.Name)
			errs = append(errs, fmt.Errorf("import cycle: %s", joinArrow(cycle)))
			return
		}
		if contains(order, pkg.Name) {
			return
		}
		visiting[pkg.Name] = true
		stack = append(stack, pkg.Name)

		type impJob struct {
			spec *ast.ImportSpec
			dir  string
		}
		seenDir := map[string]bool{}
		var jobs []impJob
		for _, f := range pkg.Files {
			base := filepath.Dir(f.Path)
			for _, im := range f.Imports {
				rel := ""
				if im.Path != nil {
					rel = im.Path.Value
				}
				if rel == "" {
					errs = append(errs, fmt.Errorf("%s: empty import path", f.Path))
					continue
				}
				dir := filepath.Clean(filepath.Join(base, rel))
				if dep := byDir[dir]; dep != nil {
					bindImport(im, dep.Name)
				}
				if seenDir[dir] {
					continue
				}
				seenDir[dir] = true
				jobs = append(jobs, impJob{spec: im, dir: dir})
			}
		}
		for _, job := range jobs {
			if dep, ok := byDir[job.dir]; ok {
				for _, f := range pkg.Files {
					base := filepath.Dir(f.Path)
					for _, im := range f.Imports {
						if im.Path != nil && filepath.Clean(filepath.Join(base, im.Path.Value)) == job.dir {
							bindImport(im, dep.Name)
						}
					}
				}
				visit(dep)
				continue
			}
			paths, err := Expand([]string{job.dir})
			if err != nil {
				errs = append(errs, fmt.Errorf("கொணர் %q: %v", job.dir, err))
				continue
			}
			files, perrs := ParsePaths(paths)
			errs = append(errs, perrs...)
			if len(perrs) != 0 {
				continue
			}
			merged, merrs := MergeFiles(files)
			errs = append(errs, merrs...)
			if len(merrs) != 0 || merged == nil {
				continue
			}
			name := merged.Package.Name.Name
			if other, ok := byName[name]; ok {
				errs = append(errs, fmt.Errorf("package name %s from %s conflicts with %s", name, job.dir, other.Dir))
				continue
			}
			dep := &Package{Name: name, Dir: job.dir, Files: files}
			byName[name] = dep
			byDir[job.dir] = dep
			for _, f := range pkg.Files {
				base := filepath.Dir(f.Path)
				for _, im := range f.Imports {
					if im.Path != nil && filepath.Clean(filepath.Join(base, im.Path.Value)) == job.dir {
						bindImport(im, name)
					}
				}
			}
			visit(dep)
		}
		stack = stack[:len(stack)-1]
		visiting[pkg.Name] = false
		order = append(order, pkg.Name)
	}
	visit(entry)
	if len(errs) != 0 {
		return nil, errs
	}
	var all []*Package
	for _, name := range order {
		all = append(all, byName[name])
	}
	return &Program{Entry: entry, All: all}, nil
}

// MergedFiles returns dependency-ordered merged package files with ImportSpec.Name/Pkg set.
func (p *Program) MergedFiles() ([]*ast.File, []error) {
	var out []*ast.File
	var errs []error
	for _, pkg := range p.All {
		m, merrs := MergeFiles(pkg.Files)
		errs = append(errs, merrs...)
		if m != nil {
			out = append(out, m)
		}
	}
	if len(errs) != 0 {
		return out, errs
	}
	return out, nil
}

func commonDir(paths []string) string {
	if len(paths) == 0 {
		return "."
	}
	dir := filepath.Dir(paths[0])
	for _, p := range paths[1:] {
		if filepath.Dir(p) != dir {
			return dir
		}
	}
	return dir
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

func joinArrow(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	out := parts[0]
	for i := 1; i < len(parts); i++ {
		out += " → " + parts[i]
	}
	return out
}
