// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cmd_test

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"golang.org/x/tools/go/packages/packagestest"
	"golang.org/x/tools/internal/lsp/cmd"
	"golang.org/x/tools/internal/lsp/tests"
	"golang.org/x/tools/internal/testenv"
	"golang.org/x/tools/internal/tool"
)

func TestMain(m *testing.M) {
	testenv.ExitIfSmallMachine()
	os.Exit(m.Run())
}

type runner struct {
	exporter packagestest.Exporter
	data     *tests.Data
	ctx      context.Context
}

func TestCommandLine(t *testing.T) {
	packagestest.TestAll(t, testCommandLine)
}

func testCommandLine(t *testing.T, exporter packagestest.Exporter) {
	data := tests.Load(t, exporter, "../testdata")
	defer data.Exported.Cleanup()

	r := &runner{
		exporter: exporter,
		data:     data,
		ctx:      tests.Context(t),
	}
	tests.Run(t, r, data)
}

// TestServeImplicitAllInterfaces checks that the server refuses to
// implicitly listen on all network interfaces: the -port flag (which bound
// ":port") is gone, and -listen rejects addresses with an empty host.
func TestServeImplicitAllInterfaces(t *testing.T) {
	const argsEnv = "GOPLS_TEST_SERVE_ARGS"
	if args := os.Getenv(argsEnv); args != "" {
		// Child process: run gopls with the given arguments. tool.Main exits
		// the process if the command fails; on success it would keep serving.
		app := cmd.New("gopls-test", "", nil)
		tool.Main(context.Background(), app, strings.Fields(args))
		os.Exit(0)
	}
	for _, test := range []struct {
		args string
		want string
	}{
		// -port=N used to listen on ":N"; a nonzero value is needed to reach
		// that code path, but the flag itself must now be rejected.
		{"serve -port=12345", "flag provided but not defined: -port"},
		{"-port=12345", "flag provided but not defined: -port"},
		{"serve -listen=:0", "-listen=:0 implicitly binds all network interfaces"},
		{"-listen=:0", "-listen=:0 implicitly binds all network interfaces"},
	} {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		c := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestServeImplicitAllInterfaces$")
		c.Env = append(os.Environ(), argsEnv+"="+test.args)
		out, err := c.CombinedOutput()
		timedOut := ctx.Err() != nil
		cancel()
		if timedOut {
			t.Errorf("gopls %s: still serving after timeout, want it to refuse to listen on all network interfaces\n%s", test.args, out)
			continue
		}
		if err == nil {
			t.Errorf("gopls %s: succeeded, want failure\n%s", test.args, out)
			continue
		}
		if !strings.Contains(string(out), test.want) {
			t.Errorf("gopls %s: output does not contain %q:\n%s", test.args, test.want, out)
		}
	}
}

func (r *runner) Completion(t *testing.T, data tests.Completions, snippets tests.CompletionSnippets, items tests.CompletionItems) {
	//TODO: add command line completions tests when it works
}

func (r *runner) FoldingRange(t *testing.T, data tests.FoldingRanges) {
	//TODO: add command line folding range tests when it works
}

func (r *runner) Highlight(t *testing.T, data tests.Highlights) {
	//TODO: add command line highlight tests when it works
}

func (r *runner) Reference(t *testing.T, data tests.References) {
	//TODO: add command line references tests when it works
}

func (r *runner) Rename(t *testing.T, data tests.Renames) {
	//TODO: add command line rename tests when it works
}

func (r *runner) PrepareRename(t *testing.T, data tests.PrepareRenames) {
	//TODO: add command line prepare rename tests when it works
}

func (r *runner) Symbol(t *testing.T, data tests.Symbols) {
	//TODO: add command line symbol tests when it works
}

func (r *runner) SignatureHelp(t *testing.T, data tests.Signatures) {
	//TODO: add command line signature tests when it works
}

func (r *runner) Link(t *testing.T, data tests.Links) {
	//TODO: add command line link tests when it works
}

func (r *runner) Import(t *testing.T, data tests.Imports) {
	//TODO: add command line imports tests when it works
}

func (r *runner) SuggestedFix(t *testing.T, data tests.SuggestedFixes) {
	//TODO: add suggested fix tests when it works
}

func captureStdOut(t testing.TB, f func()) string {
	r, out, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	defer func() {
		os.Stdout = old
		out.Close()
		r.Close()
	}()
	os.Stdout = out
	f()
	out.Close()
	data, err := ioutil.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// normalizePaths replaces all paths present in s with just the fragment portion
// this is used to make golden files not depend on the temporary paths of the files
func normalizePaths(data *tests.Data, s string) string {
	type entry struct {
		path     string
		index    int
		fragment string
	}
	match := make([]entry, 0, len(data.Exported.Modules))
	// collect the initial state of all the matchers
	for _, m := range data.Exported.Modules {
		for fragment := range m.Files {
			filename := data.Exported.File(m.Name, fragment)
			index := strings.Index(s, filename)
			if index >= 0 {
				match = append(match, entry{filename, index, fragment})
			}
			if slash := filepath.ToSlash(filename); slash != filename {
				index := strings.Index(s, slash)
				if index >= 0 {
					match = append(match, entry{slash, index, fragment})
				}
			}
			quoted := strconv.Quote(filename)
			if escaped := quoted[1 : len(quoted)-1]; escaped != filename {
				index := strings.Index(s, escaped)
				if index >= 0 {
					match = append(match, entry{escaped, index, fragment})
				}
			}
		}
	}
	// result should be the same or shorter than the input
	buf := bytes.NewBuffer(make([]byte, 0, len(s)))
	last := 0
	for {
		// find the nearest path match to the start of the buffer
		next := -1
		nearest := len(s)
		for i, c := range match {
			if c.index >= 0 && nearest > c.index {
				nearest = c.index
				next = i
			}
		}
		// if there are no matches, we copy the rest of the string and are done
		if next < 0 {
			buf.WriteString(s[last:])
			return buf.String()
		}
		// we have a match
		n := &match[next]
		// copy up to the start of the match
		buf.WriteString(s[last:n.index])
		// skip over the filename
		last = n.index + len(n.path)
		// add in the fragment instead
		buf.WriteString(n.fragment)
		// see what the next match for this path is
		n.index = strings.Index(s[last:], n.path)
		if n.index >= 0 {
			n.index += last
		}
	}
}
