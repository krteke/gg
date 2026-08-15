package internal

import (
	"encoding/csv"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strconv"
)

type CheckConfig struct {
	Root string
	Job  string
}

func (c *CheckConfig) Check() error {
	buckets, err := readTsv(c.Job)
	if err != nil {
		return err
	}

	info, err := scanRoot(c.Root)
	if err != nil {
		return err
	}

	filesMap := toMap(info)

	var notFound []string
	var sizeUnmatched []string
	var total uint64

	for _, bucket := range buckets {
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

func readTsv(job string) ([][]FileInfo, error) {
	buckets := [][]FileInfo{}

	err := filepath.WalkDir(job, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			slog.Error("walk dir", "err", err)
			error := Err(ErrWalkDir, "walk dir", err)
			return error
		}

		if d.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			error := Err(ErrReadFile, "read file "+path, err)
			return error
		}
		defer file.Close()

		reader := csv.NewReader(file)
		reader.Comma = '\t'

		buffer := []FileInfo{}
		for {
			rec, err := reader.Read()
			if err == io.EOF {
				buckets = append(buckets, slices.Clone(buffer))
				break
			}
			if err != nil {
				error := Err(ErrReadFile, "read TSV line", err)
				return error
			}

			size, err := strconv.ParseUint(rec[0], 10, 64)
			if err != nil {
				error := Err(ErrParseUInt, "parse size "+rec[0], err)
				return error
			}
			path := rec[1]
			buffer = append(buffer, FileInfo{Size: size, Path: path})
		}

		return nil
	})

	return buckets, err
}

func toMap(files []FileInfo) map[string]uint64 {
	filesMap := make(map[string]uint64, len(files))

	for _, file := range files {
		filesMap[file.Path] = file.Size
	}
	return filesMap
}
