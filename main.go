package main

import (
	"fmt"
	"io"
	"os"
	"crypto/rand"
	"path/filepath"
)

type zeroReader struct{}

func (zeroReader) Read(b []byte) (int, error) {
	for i := range b { b[i] = 0 }
	return len(b), nil
}

type xorReader []io.Reader

func (r xorReader) Read(b []byte) (int, error) {
	for i := range b { b[i] = 0 }
	// we read one file and expect to read as much in other files
	n, err := r[0].Read(b)
	if err != nil {
		return 0, err
	}
	b2 := make([]byte, n)
	for i := 1; i < len(r); i++ {
		if _, err := io.ReadFull(r[i], b2); err != nil {
			return 0, err
		}
		for i := range b2 {
			b[i] = b[i] ^ b2[i]
		}
	}
	return n, nil
}

type randWriter []io.Writer

func (r randWriter) Write(b []byte) (int, error) {
	b2 := make([]byte, len(b))
	copy(b2, b)
	b3 := make([]byte, len(b))
	for i := 0; i < len(r)-1; i++ {
		if _, err := io.ReadFull(rand.Reader,b3); err != nil {
			return 0, err
		}
		for j := 0; j < len(b); j++ {
			b2[j] = b2[j] ^ b3[j]
		}
		if _, err := r[i].Write(b3); err != nil {
			return 0, err
		}
	}
	return r[len(r)-1].Write(b2)
}

func ShowHelp() {
	fmt.Fprintf(os.Stderr, `Usage: %s [-i DIR | -o DIR]...
-i DIR    Add an input directory or file
-o DIR    Add an output directory or file
`, os.Args[0])
}

type DirInfo struct {
	Subdirs map[string]struct{}
	Files map[string]int64
}

func (d *DirInfo) Scan(base string, prefix string) error {
	path := filepath.Join(base, prefix)
	dir, err := os.Open(path)
	if err != nil {
		return err
	}
	defer dir.Close()
	stat, err := dir.Stat()
	if err != nil {
		return err
	}
	if stat.IsDir() {
		d.Subdirs[prefix] = struct{}{}
		names, err := dir.Readdirnames(0)
		if err != nil {
			return err
		}
		for _, name := range names {
			if err := d.Scan(base, filepath.Join(prefix, name)); err != nil {
				return err
			}
		}
	} else {
		maxSize := d.Files[prefix]
		curSize := stat.Size()
		if curSize > maxSize {
			d.Files[prefix] = curSize
		}
	}
	return nil
}

func main() {
	var InputDirs []string
	var OutputDirs []string
	args := os.Args
	for i := 1; i < len(args); {
		switch args[i] {
		case "-i":
			InputDirs = append(InputDirs, args[i+1])
			i += 2
		case "-o":
			OutputDirs = append(OutputDirs, args[i+1])
			i += 2
		default:
			ShowHelp()
			return
		}
	}
	if len(InputDirs) == 0 {
		fmt.Fprintln(os.Stderr, "No input directories specified")
		return
	}
	if len(OutputDirs) == 0 {
		fmt.Fprintln(os.Stderr, "No output directories specified")
		return
	}
	info := &DirInfo{Subdirs: make(map[string]struct{}), Files: make(map[string]int64)}
	for _, dir := range InputDirs {
		if err := info.Scan(dir, ""); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to scan input directory: %s\n", err)
			return
		}
	}

	// Make directory structure
	for path := range info.Subdirs {
		for _, outDir := range OutputDirs {
			curPath := filepath.Join(outDir, path)
			if err := os.MkdirAll(curPath, 0775); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to make directory: %s\n", err)
				return
			}
		}
	}

	// Create files
	for path, size := range info.Files {
		if err := XorFile(path, size, InputDirs, OutputDirs); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to xor file: %s\n", err)
			return
		}
	}
}

func XorFile(path string, size int64, InputDirs []string, OutputDirs []string) error {
	var reader xorReader
	for _, inpDir := range InputDirs {
		file, err := os.Open(filepath.Join(inpDir, path))
		if err != nil {
			return err
		}
		defer file.Close()
		reader = append(reader, io.LimitReader(io.MultiReader(file, zeroReader{}), size))
	}
	var writer randWriter
	for _, outDir := range OutputDirs {
		file, err := os.Create(filepath.Join(outDir, path))
		if err != nil {
			return err
		}
		defer file.Close()
		writer = append(writer, file)
	}
	_, err := io.Copy(writer, reader)
	return err
}
