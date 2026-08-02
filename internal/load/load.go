// Package load expands and parses Aram source paths (files or directories).
package load

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"aram/internal/ast"
	"aram/internal/parse"
)

// Expand turns CLI args into a sorted list of .aram file paths.
// A directory argument expands to all *.aram files directly inside it
// (non-recursive).
func Expand(args []string) ([]string, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("no source paths")
	}
	seen := map[string]bool{}
	var out []string
	for _, a := range args {
		info, err := os.Stat(a)
		if err != nil {
			return nil, err
		}
		if info.IsDir() {
			entries, err := os.ReadDir(a)
			if err != nil {
				return nil, err
			}
			var dirFiles []string
			for _, e := range entries {
				if e.IsDir() {
					continue
				}
				name := e.Name()
				if strings.HasSuffix(name, ".aram") {
					dirFiles = append(dirFiles, filepath.Join(a, name))
				}
			}
			if len(dirFiles) == 0 {
				return nil, fmt.Errorf("no .aram files in %s", a)
			}
			sort.Strings(dirFiles)
			for _, p := range dirFiles {
				if !seen[p] {
					seen[p] = true
					out = append(out, p)
				}
			}
			continue
		}
		if !strings.HasSuffix(a, ".aram") {
			return nil, fmt.Errorf("not an .aram file: %s", a)
		}
		if !seen[a] {
			seen[a] = true
			out = append(out, a)
		}
	}
	sort.Strings(out)
	return out, nil
}

// ParsePaths reads and parses each path. Returns files (possibly partial) and errors.
func ParsePaths(paths []string) ([]*ast.File, []error) {
	var files []*ast.File
	var errs []error
	for _, path := range paths {
		src, err := os.ReadFile(path)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		f, perrs := parse.ParseFile(path, string(src))
		errs = append(errs, perrs...)
		if f != nil {
			files = append(files, f)
		}
	}
	return files, errs
}

// DefaultOutName chooses build/<name> for emit/build when -o is omitted.
func DefaultOutName(paths []string) string {
	if len(paths) == 0 {
		return "out"
	}
	if len(paths) == 1 {
		return strings.TrimSuffix(filepath.Base(paths[0]), filepath.Ext(paths[0]))
	}
	dir := filepath.Dir(paths[0])
	for _, p := range paths[1:] {
		if filepath.Dir(p) != dir {
			return strings.TrimSuffix(filepath.Base(paths[0]), filepath.Ext(paths[0]))
		}
	}
	base := filepath.Base(dir)
	if base == "" || base == "." || base == string(filepath.Separator) {
		return strings.TrimSuffix(filepath.Base(paths[0]), filepath.Ext(paths[0]))
	}
	return base
}
