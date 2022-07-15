package main

import (
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cheggaaa/pb/v3"
)

var (
	srcDir, dstDir string
	runAfter       time.Duration
	ignoreDotFiles bool
)

func init() {
	flag.StringVar(&srcDir, "src", "", "source directory")
	flag.StringVar(&dstDir, "dst", "", "destination directory")
	flag.DurationVar(&runAfter, "", 24*time.Hour, "copy files after this duration")
	flag.BoolVar(&ignoreDotFiles, "ignore-dot-files", true, "ignore dot files")

	flag.Parse()
}

func copyFile(src, dst string, size int64) error {
	fmt.Printf("Copying %s\n", src)

	s, err := os.Open(src)

	if err != nil {
		return err
	}
	defer s.Close()

	bar := pb.Full.Start64(size)
	defer bar.Finish()

	d, err := os.Create(filepath.Join(dst, filepath.Base(src)))
	if err != nil {
		return err
	}
	defer d.Close()

	_, err = io.Copy(d, bar.NewProxyReader(s))
	if err != nil {
		return err
	}

	return nil
}

func main() {
	filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if path == srcDir {
			return nil
		}
		if d.IsDir() {
			return fs.SkipDir
		}

		if ignoreDotFiles && strings.HasPrefix(filepath.Base(path), ".") {
			return nil
		}

		info, err := d.Info()

		if err != nil {
			return err
		}

		if time.Since(info.ModTime()) < runAfter {
			return nil
		}

		return copyFile(path, dstDir, info.Size())
	})
}
