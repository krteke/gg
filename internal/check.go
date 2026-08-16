package internal

import (
	"encoding/csv"
	"fmt"
	"io"
	"io/fs"
	"iter"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/dustin/go-humanize"
)

type CheckConfig struct {
	Root string
	Job  string
}

func (c *CheckConfig) Check() error {
	info, err := scanRoot(c.Root)
	if err != nil {
		return err
	}

	filesMap := toMap(info)

	var notFound []string
	var sizeUnmatched []string
	var total uint64

	for bucket, err := range readFileInfo(c.Job) {
		if err != nil {
			return err
		}

		bucketLen := len(bucket)
		total += uint64(bucketLen)
		for _, file := range bucket {
			path := file.Path
			size := file.Size

			value, ok := filesMap[path]
			if !ok {
				slog.Warn("file not found", "path", path)
				notFound = append(notFound, path)
			} else if value != size {
				slog.Warn("unmatched size", "expect", size, "found", value, "path at", path)
				sizeUnmatched = append(sizeUnmatched, path)
			}
		}
	}

	fmt.Printf("Checked:\t%d\nMissing:\t%d\nSize mismatch:\t%d\n", total, len(notFound), len(sizeUnmatched))

	return nil
}

func readFileInfo(job string) iter.Seq2[[]FileInfo, error] {
	return func(yield func([]FileInfo, error) bool) {
	O:
		for bucketIter, err := range readTSV(job) {
			if err != nil {
				if !yield(nil, err) {
					return
				}
				continue
			}

			bucket := []FileInfo{}
			for item, err := range bucketIter {
				if err != nil {
					if !yield(nil, err) {
						return
					}
					continue O
				}

				size, err := humanize.ParseBytes(item[0])
				if err != nil {
					if !yield(nil, err) {
						return
					}
					continue O
				}

				bucket = append(bucket, FileInfo{Size: size, Path: item[1]})
			}
			if !yield(bucket, nil) {
				return
			}
		}
	}
}

func readTSV(root string) iter.Seq2[iter.Seq2[[]string, error], error] {
	return func(yield func(iter.Seq2[[]string, error], error) bool) {
		_, err := os.Stat(root)
		if err != nil {
			error := Err(ErrWalkDir, "walk directory "+root, err)
			yield(nil, error)
			return
		}

		err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				slog.Error("walk dir", "err", err)
				return err
			}

			if d.IsDir() {
				return nil
			}

			if !yield(readTSVFile(path), nil) {
				return fs.SkipAll
			}

			return nil
		})
		if err != nil {
			yield(nil, err)
		}
	}
}

func readTSVFile(path string) iter.Seq2[[]string, error] {
	return func(yield func([]string, error) bool) {
		file, err := os.Open(path)
		if err != nil {
			error := Err(ErrReadFile, "read file "+path, err)
			yield(nil, error)
			return
		}
		defer file.Close()

		reader := csv.NewReader(file)
		reader.Comma = '\t'

		for {
			rec, err := reader.Read()
			if err == io.EOF {
				return
			}
			if err != nil {
				error := Err(ErrReadFile, "read TSV line", err)
				yield(nil, error)
				return
			}

			if !yield(rec, nil) {
				return
			}
		}
	}
}

func toMap(files []FileInfo) map[string]uint64 {
	filesMap := make(map[string]uint64, len(files))

	for _, file := range files {
		filesMap[file.Path] = file.Size
	}
	return filesMap
}
