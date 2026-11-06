// Command testplan discovers tests and their transitive reference-file inputs.
package main

import (
	"encoding/json"
	"flag"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type test struct {
	Package string   `json:"package"`
	Name    string   `json:"name"`
	Suite   string   `json:"suite"`
	Inputs  []string `json:"inputs,omitempty"`
	Env     []string `json:"env,omitempty"`
}

type function struct {
	name, file string
	calls      map[string]bool
	inputs     map[string]bool
	env        map[string]bool
	binary     bool
}

func inputs(node ast.Node, directory, name string) (map[string]bool, map[string]bool, bool) {
	files, env := map[string]bool{}, map[string]bool{}
	var values []string
	ast.Inspect(node, func(n ast.Node) bool {
		if literal, ok := n.(*ast.BasicLit); ok && literal.Kind == token.STRING {
			if value, err := strconv.Unquote(literal.Value); err == nil {
				values = append(values, value)
			}
		}
		return true
	})
	domain := "worldgen"
	for _, value := range values {
		if strings.Contains(value, "testdata/protocol") || value == "protocol" {
			domain = "protocol"
		}
	}
	binary := false
	local := func(value string) string {
		return filepath.ToSlash(filepath.Clean(filepath.Join(strings.TrimPrefix(name, "./"), value)))
	}
	for _, value := range values {
		if index := strings.Index(value, "testdata/"); index >= 0 {
			files[local(value)] = true
		} else if !strings.HasPrefix(value, "_") && !strings.ContainsAny(value, `/\`) && strings.Contains(value, "1.21.1") &&
			(strings.HasSuffix(value, ".txt") || strings.HasSuffix(value, ".gz") || strings.HasSuffix(value, ".json")) {
			if _, err := os.Stat(filepath.Join(directory, value)); err == nil {
				files[local(value)] = true
			} else {
				files["testdata/"+domain+"/"+value] = true
			}
		}
		if value == "ATLAS_SERVER_EXE" || value == "ATLAS_TEST_EXE" {
			binary = true
		}
		if value == "ATLAS_SERVER_JAR" || value == "ATLAS_VANILLA_BIOME_REFERENCE_WORLD" {
			env[value] = true
		}
	}
	// filepath.Join forms must be resolved as a whole, including a filename
	// supplied separately from the testdata directory.
	ast.Inspect(node, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || selector.Sel.Name != "Join" {
			return true
		}
		var parts []string
		for _, argument := range call.Args {
			literal, ok := argument.(*ast.BasicLit)
			if !ok || literal.Kind != token.STRING {
				return true
			}
			value, _ := strconv.Unquote(literal.Value)
			parts = append(parts, value)
		}
		joined := filepath.ToSlash(filepath.Join(parts...))
		if index := strings.Index(joined, "testdata/"); index >= 0 {
			files[local(joined)] = true
		}
		return true
	})
	return files, env, binary
}

func keys(values map[string]bool) []string {
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func packageTests(directory, name string) ([]test, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}
	functions := map[string]*function{}
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		file, err := parser.ParseFile(token.NewFileSet(), filepath.Join(directory, entry.Name()), nil, 0)
		if err != nil {
			return nil, err
		}
		for _, declaration := range file.Decls {
			fn, ok := declaration.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			files, env, binary := inputs(fn.Body, directory, name)
			f := &function{fn.Name.Name, entry.Name(), map[string]bool{}, files, env, binary}
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				if call, ok := n.(*ast.CallExpr); ok {
					called := ""
					if id, ok := call.Fun.(*ast.Ident); ok {
						f.calls[id.Name] = true
						called = id.Name
					} else if selector, ok := call.Fun.(*ast.SelectorExpr); ok {
						called = selector.Sel.Name
						if owner, ok := selector.X.(*ast.Ident); ok && owner.Name == "world" && called == "Open" {
							f.calls["@integration"] = true
						}
					}
					if strings.HasPrefix(called, "Generate") ||
						(strings.HasPrefix(called, "New") && strings.HasSuffix(called, "Generator")) ||
						called == "CreateNormalWorld" || called == "CreateFlatWorld" || called == "OpenDimension" ||
						(name == "./internal/world" && called == "Open") {
						f.calls["@integration"] = true
					}
				}
				return true
			})
			functions[f.name] = f
		}
	}
	var tests []test
	for _, f := range functions {
		if !strings.HasPrefix(f.name, "Test") {
			continue
		}
		seen, files, env := map[string]bool{}, map[string]bool{}, map[string]bool{}
		binary := false
		integration := false
		var visit func(string)
		visit = func(name string) {
			if seen[name] || functions[name] == nil {
				return
			}
			seen[name] = true
			fn := functions[name]
			for value := range fn.inputs {
				files[value] = true
			}
			for value := range fn.env {
				env[value] = true
			}
			binary = binary || fn.binary
			integration = integration || fn.calls["@integration"]
			for call := range fn.calls {
				visit(call)
			}
		}
		visit(f.name)
		suite := "basic"
		if integration {
			suite = "integration"
		}
		externalFiles := false
		for path := range files {
			externalFiles = externalFiles || strings.HasPrefix(path, "testdata/")
		}
		if externalFiles || len(env) > 0 {
			suite = "reference"
		}
		if binary {
			suite = "smoke"
			if strings.Contains(f.file, "performance") {
				suite = "performance"
			}
		}
		tests = append(tests, test{name, f.name, suite, keys(files), keys(env)})
	}
	sort.Slice(tests, func(i, j int) bool { return tests[i].Name < tests[j].Name })
	return tests, nil
}

func main() {
	root := flag.String("root", ".", "repository root")
	flag.Parse()
	var tests []test
	for _, directory := range []string{"internal", "cmd"} {
		err := filepath.WalkDir(filepath.Join(*root, directory), func(path string, entry os.DirEntry, err error) error {
			if err != nil || !entry.IsDir() {
				return err
			}
			relative, _ := filepath.Rel(*root, path)
			found, err := packageTests(path, "./"+filepath.ToSlash(relative))
			tests = append(tests, found...)
			return err
		})
		if err != nil {
			panic(err)
		}
	}
	if err := json.NewEncoder(os.Stdout).Encode(tests); err != nil {
		panic(err)
	}
}
