package internal

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"iter"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/dustin/go-humanize"
)

type CheckConfig struct {
	Root   string
	Job    string
	Report string
}

func (c *CheckConfig) Check() error {
	hasReport := strings.TrimSpace(c.Report) != ""
	if hasReport {
		if err := os.MkdirAll(c.Report, 0755); err != nil {
			error := Err(ErrMkdir, "create directory "+c.Report, err)
			return error
		}
	}

	_, err := os.Stat(c.Root)
	if err != nil {
		error := Err(ErrWalkDir, "root not exist "+c.Root, err)
		return error
	}

	notFound := 0
	errors := 0
	sizeUnmatched := 0
	var total uint64

	index := 0
	for bucket, err := range readFileInfo(c.Job) {
		if err != nil {
			return err
		}

		bucketLen := len(bucket)
		total += uint64(bucketLen)

		if hasReport {
			hasWarn := false

			output := path.Join(c.Report, fmt.Sprintf("%04d", index)+".warn.partial")
			outputFile, err := os.OpenFile(output, os.O_CREATE|os.O_TRUNC|os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				error := Err(ErrOpenFile, "open file"+output, err)
				return error
			}
			defer outputFile.Close()

			writer := csv.NewWriter(outputFile)
			writer.Comma = '\t'
			defer writer.Flush()

			for _, file := range bucket {
				path := file.Path
				size := file.Size

				value, ok, error := findFile(c.Root, path)
				if error != nil {
					hasWarn = true
					errors++
					writeLine(writer, path, error.Error())
				} else if !ok {
					hasWarn = true
					slog.Warn("file not found", "path", path)
					notFound++
					writeLine(writer, path, "not found")
				} else if value != size {
					hasWarn = true
					slog.Warn("unmatched size", "expect", size, "found", value, "path at", path)
					sizeUnmatched++
					writeLine(writer, path, "size unmatched")
				}
			}
			writer.Flush()
			outputFile.Close()
			if hasWarn {
				new := strings.TrimSuffix(output, filepath.Ext(output))
				rename(output, new)
			} else {
				new := strings.TrimSuffix(output, filepath.Ext(output))
				new = strings.TrimSuffix(new, filepath.Ext(new))
				rename(output, new+".ok")
			}
		} else {
			for _, file := range bucket {
				path := file.Path
				size := file.Size

				value, ok, error := findFile(c.Root, path)
				if error != nil {
					errors++
				} else if !ok {
					slog.Warn("file not found", "path", path)
					notFound++
				} else if value != size {
					slog.Warn("unmatched size", "expect", size, "found", value, "path at", path)
					sizeUnmatched++
				}
			}
		}

		index++
	}

	fmt.Printf(
		"Checked:\t%d\nMissing:\t%d\nSize mismatch:\t%d\nErrors:\t%d\n",
		total, notFound, sizeUnmatched, errors)

	return nil
}

func rename(from string, to string) {
	if err := os.Rename(from, to); err != nil {
		slog.Error("failed to rename", "file", from, "err", err)
	}
}

func findFile(root, path string) (uint64, bool, error) {
	path = filepath.Join(root, path)

	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return 0, false, nil
		}
		return 0, false, err
	}

	return uint64(info.Size()), true, nil
}

func writeLine(writer *csv.Writer, path string, err string) {
	line := []string{
		path,
		err,
	}

	if err := writer.Write(line); err != nil {
		slog.Warn("write error line error", "err", err)
	}
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
