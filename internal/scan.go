package internal

import (
	"fmt"
	"io/fs"
	"log/slog"
	"path/filepath"
	"slices"
)

type FileInfo struct {
	Path string
	Size int64
}

func Scan(config ScanConfig) error {
	var files []FileInfo

	root := config.Root
	cfg, err := LoadConfig(config.ConfigPath)
	if err != nil {
		return err
	}

	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
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
			Size: info.Size(),
		})

		return nil
	})

	if err != nil {
		return err
	}

	max := cfg.Bucket.Max

	var buckets [][]FileInfo
	var size int64 = 0
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

	for i, b := range buckets {
		var size int64
		for _, f := range b {
			size += f.Size
		}

		fmt.Printf("bucket: %d, size: %.2f Gib\n", i, float64(size)/1024.0/1024.0/1024.0)
	}

	return nil
}
