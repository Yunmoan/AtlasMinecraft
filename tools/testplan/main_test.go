package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTransitiveReferenceAndBinaryRequirements(t *testing.T) {
	directory := t.TempDir()
	source := `package sample
func reference() { os.ReadFile(filepath.Join("..", "..", "testdata", "worldgen", "tree_feature_1.21.1.txt")) }
func middle() { reference() }
func TestReference(t *testing.T) { middle() }
func TestBasic(t *testing.T) { println("basic") }
func binary() { os.Getenv("ATLAS_SERVER_EXE") }
func TestBinary(t *testing.T) { binary() }
`
	if err := os.WriteFile(filepath.Join(directory, "sample_test.go"), []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	tests, err := packageTests(directory, "./internal/example")
	if err != nil || len(tests) != 3 {
		t.Fatalf("%v: %v", tests, err)
	}
	if tests[0].Suite != "basic" || tests[1].Suite != "smoke" || tests[2].Suite != "reference" {
		t.Fatal(tests)
	}
	if len(tests[2].Inputs) != 1 || tests[2].Inputs[0] != "testdata/worldgen/tree_feature_1.21.1.txt" {
		t.Fatal(tests[2])
	}
}

func TestPackageLocalFixtureStaysInBasicSuite(t *testing.T) {
	directory := t.TempDir()
	source := `package sample
func TestLocal(t *testing.T) { os.ReadFile("testdata/fixture.bin") }
`
	if err := os.WriteFile(filepath.Join(directory, "sample_test.go"), []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	tests, err := packageTests(directory, "./internal/example")
	if err != nil || len(tests) != 1 || tests[0].Suite != "basic" || tests[0].Inputs[0] != "internal/example/testdata/fixture.bin" {
		t.Fatalf("%v: %v", tests, err)
	}
}
