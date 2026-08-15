package internal

import (
	"encoding/csv"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
)

type FileInfo struct {
	Path string
	Size uint64
}

type ScanConfig struct {
	Root   string
	Output string
	Max    string
}

func (config *ScanConfig) Scan() error {
	root := config.Root

	files, err := scanRoot(root)
	if err != nil {
		return err
	}

	max, err := humanize.ParseBytes(config.Max)
	if err != nil {
		error := Err(ErrParseUInt, "parse max bytes "+config.Max, err)
		return error
	}

	var buckets [][]FileInfo
	var size uint64 = 0
	var buffer []FileInfo

	for _, file := range files {
		testSize := size + file.Size

		if file.Size <= max && testSize <= max {
			size = testSize
			buffer = append(buffer, file)
		} else if file.Size > max {
			buckets = append(buckets, []FileInfo{file})
		} else {
			buckets = append(buckets, slices.Clone(buffer))
			buffer = buffer[:0]
			size = file.Size
			buffer = append(buffer, file)
		}
	}
	buckets = append(buckets, buffer)

	output := strings.TrimSpace(config.Output)
	if output == "" {
		output = "job-" + time.Now().Format("20060102")
	}

	err = os.MkdirAll(output, 0755)
	if err != nil {
		error := Err(ErrMkdir, "create output directory "+output, err)
		return error
	}

	for i, b := range buckets {
		file := path.Join(output, strconv.Itoa(i)+".lst")
		f, err := os.Create(file)
		if err != nil {
			error := Err(ErrCreateFile, "create file "+file, err)
			return error
		}
		defer f.Close()

		writer := csv.NewWriter(f)
		writer.Comma = '\t'
		defer writer.Flush()

		for _, info := range b {
			row := []string{
				strconv.FormatUint(info.Size, 10),
				info.Path,
			}

			if err := writer.Write(row); err != nil {
				error := Err(ErrWriteFile, fmt.Sprintf("write TSV at %q %s", file, row), err)
				return error
			}
		}
	}

	return nil
}

func scanRoot(root string) ([]FileInfo, error) {
	var files []FileInfo

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			slog.Error("walk dir", "err", err)
			error := Err(ErrWalkDir, "walk dir", err)
			return error
		}

		if d.IsDir() {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			slog.Error("read file info", "err", err)
			error := Err(ErrReadInfo, "read file info", err)
			return error
		}

		relPath, err := filepath.Rel(root, path)
		if err != nil {
			slog.Error("get relative path", "err", err)
			error := Err(ErrRelPath, "get relative path", err)
			return error
		}

		relPath = filepath.ToSlash(relPath)
		files = append(files, FileInfo{
			Path: relPath,
			Size: uint64(info.Size()),
		})

		return nil
	})

	return files, err
}
