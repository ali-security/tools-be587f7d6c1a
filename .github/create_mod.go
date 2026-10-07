// Command create_mod packs a module directory into a module zip exactly as
// proxy.golang.org serves it. Used only by .github/workflows/seal-build.yml.
//
// Usage: create_mod <module-path> <version> <dir> <out.zip>
package main

import (
	"log"
	"os"

	"golang.org/x/mod/module"
	"golang.org/x/mod/zip"
)

func main() {
	if len(os.Args) != 5 {
		log.Fatalf("usage: %s <module-path> <version> <dir> <out.zip>", os.Args[0])
	}
	f, err := os.Create(os.Args[4])
	if err != nil {
		log.Fatal(err)
	}
	mv := module.Version{Path: os.Args[1], Version: os.Args[2]}
	if err := zip.CreateFromDir(f, mv, os.Args[3]); err != nil {
		log.Fatal(err)
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}
