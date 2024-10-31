// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gsync

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// copyFile copies a single file from src to dst, overwriting dst if it exists.
// if chmod is true, the source protections are copied to the dest even if it
// already existed.
func copyFile(src, dst string, chmod bool, mode fs.FileMode) (e error) {
	var sourceFile, destFile *os.File
	sourceFile, e = os.Open(src)
	if e != nil {
		return 
	}
	defer sourceFile.Close()

	var fi fs.FileInfo
	fi, e = os.Stat(src)
	if e != nil {
		return 
	}
	if mode == 0 {
		mode = fi.Mode()
	}

	destFile, e = os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, mode)
	if e != nil {
		return 
	}
	defer destFile.Close()
	if chmod {
		defer func () {
			if e == nil {
				e = os.Chmod(dst, mode)
			}
		}()
	}

	_, e = io.Copy(destFile, sourceFile)

	return 
}

// copyDir recursively copies all files from srcDir to dstDir,
// ignoring files that already exist in dstDir.
//
// force causes an attempt to chmod a not-writeable destination file.
//
// the default file mode is copied from the source,
// and the default directory mode is 0755.  If mode != 0 ,
// this overrides the default choices for file and directory modes.
func CopyDir(srcDir, dstDir string, mode fs.FileMode, force, verbose bool) error {
	dirMode := fs.FileMode(0755)
	if mode != 0 {
		// this won't work unless we can RWX the directory ourselves
		dirMode = mode | 0700
	}

	return filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Construct the destination path
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dstDir, relPath)

		// Handle directories
		if d.IsDir() {
			if verbose {
				fmt.Printf("Making directory %s\n", dstPath)
			}
			if err := os.MkdirAll(dstPath, dirMode); err != nil {
				return err
			}
			return nil
		}

		chmodAfter := false

		// Check if the file already exists in the destination, and is wrong size or out of date.
		if dfi, err := os.Stat(dstPath); err == nil {
			sfi, err := os.Stat(path)
			if err != nil {
				return err
			}
			// if the sizes are equal and the source is not modified after the destination,
			// then there is no need to copy.
			if sfi.Size() == dfi.Size() && !sfi.ModTime().After(dfi.ModTime()) {
				if verbose {
					fmt.Printf("Skipping %s (already exists w/ same size and not-older mod time)\n", dstPath)
				}
				return nil
			}
			// Must be able to read and write the file, somehow.
			if dfi.Mode()&0222 == 0 || dfi.Mode()&0444 == 0 {
				if force {
					// Add read/write permission for the owner, if possible
					newMode := dfi.Mode() | 0600
					// Change the file mode
					err = os.Chmod(dstPath, newMode)
					if err != nil {
						return err
					}
				} else {
					return fmt.Errorf("%s needs to be updated but is not writeable. Specifying -f will override, if permitted", dstPath)
				}
			}
			// Since the file already exists, the create in copyfile will not set its mode.
			chmodAfter = true
		}

		// Copy the file
		if verbose {
			fmt.Printf("Copying %s to %s\n", path, dstPath)
		}

		return copyFile(path, dstPath, chmodAfter, mode)
	})
}
