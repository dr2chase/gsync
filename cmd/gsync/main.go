// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"github.com/dr2chase/gsync"
	"io/fs"
	"os"
	"strconv"
	"strings"
)

func usage(cmd, extra string) {
		if extra != "" {
			fmt.Fprintf(os.Stderr,"%s\n", extra)
		}
		fmt.Fprintf(os.Stderr,
			`USED LIKE THIS: %s [-v / -f / -m octal-XYZ ] srcDir dstDir

recursively copies the contents of srcDir to dstDir
unless a destination file already exists.

-f override permissions in destination, if possible.
-v verbose

It is a stupid, local, implementation of "rsync -a srcDir/ dstDir"
`, cmd)
		os.Exit(1)

}

func main() {
	cmd := os.Args[0]
	args := os.Args[1:]
	force := false
	verbose := false
	mode := fs.FileMode(0)

	for ; len(args) > 0 && len(args[0]) > 0 && args[0][0] == '-'; args = args[1:] {
		whine := true
		if strings.Contains(args[0], "v") {
			verbose = true
			whine = false
		}
		if strings.Contains(args[0], "f") {
			force = true
			whine = false
		}
		if strings.Contains(args[0], "m") {
			args = args[1:]
			if len(args) == 0 {
				usage(cmd, "-m needs a value")
			}
			m, err := strconv.ParseInt(args[0], 8, 64)
			if err != nil {
				usage(cmd, "-m needs an octal value")
			}
			mode = fs.FileMode(m)
			whine = false
			fmt.Printf("Saw mode %o\n", mode)
		}
		if whine {
			usage(cmd, "unrecognized option " + args[0])
		}
	}

	if len(args) != 2 {
		usage(cmd, "needs source AND destination")
	}

	err := gsync.CopyDir(args[0], args[1], mode, force, verbose)
	if err != nil {
		fmt.Println("Error copying directory:", err)
	}
}
