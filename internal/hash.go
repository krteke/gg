package internal

import (
	"encoding/csv"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"iter"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/zeebo/blake3"
)

type HashConfig struct {
	Root   string
	Job    string
	Output string
	At     uint32
	Retry  *uint32
}

func At[T any](seq iter.Seq2[T, error], index uint32) (T, error) {
	var zero T

	var i uint32
	for v, err := range seq {
		if i == index {
			if err != nil {
				return zero, err
			}

			return v, nil
		}
		i++
	}

	return zero, Err(ErrOutofRange, fmt.Sprintf("index %d len %d", index, i), ErrOutofRange)
}

func Skip[T any](seq iter.Seq2[T, error], n uint32) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		var i uint32

		for v, err := range seq {
			if errors.Is(err, ErrWalkDir) {
				yield(v, err)
				return
			}

			if i < n {
				i++
				continue
			}

			if !yield(v, err) {
				return
			}
		}
	}
}

func (c *HashConfig) Hash() error {
	if err := os.MkdirAll(c.Output, 0755); err != nil {
		error := Err(ErrMkdir, "create directory "+c.Output, err)
		return error
	}

	if c.Retry != nil {
		retry := *c.Retry
		bucket, err := At(readFileInfo(c.Job), retry)
		if err != nil {
			return err
		}

		if err := c.processBucket(bucket, int(retry)); err != nil {
			return err
		}
	}

	index := c.At
	for bucket, err := range Skip(readFileInfo(c.Job), c.At) {
		if err != nil {
			return err
		}

		if err := c.processBucket(bucket, int(index)); err != nil {
			return err
		}

		index++
	}
	return nil
}

func (c *HashConfig) processBucket(bucket []FileInfo, index int) error {
	output := path.Join(c.Output, fmt.Sprintf("%04d", index)+".hash.partial")
	outputFile, err := os.OpenFile(output, os.O_CREATE|os.O_TRUNC|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		error := Err(ErrOpenFile, "open file"+output, err)
		return error
	}
	defer outputFile.Close()

	writer := csv.NewWriter(outputFile)
	writer.Comma = '\t'

	c.hashBucket(writer, bucket)

	new := strings.TrimSuffix(output, filepath.Ext(output))
	writer.Flush()
	if err := writer.Error(); err != nil {
		return Err(ErrFlush, "flush "+output, err)
	}
	err = os.Rename(output, new)
	if err != nil {
		error := Err(ErrRename, "rename "+output, err)
		return error
	}

	return nil
}

func (c *HashConfig) hashBucket(writer *csv.Writer, bucket []FileInfo) {
	for _, f := range bucket {
		hasher := blake3.New()

		p := path.Join(c.Root, f.Path)

		file, err := os.Open(p)
		if err != nil {
			writeErrLine(writer, err, f)
			continue
		}
		defer file.Close()

		if _, err := io.Copy(hasher, file); err != nil {
			writeErrLine(writer, err, f)
			continue
		}

		sum := hex.EncodeToString(hasher.Sum(nil))

		if err := writeSuccessLine(writer, sum, f); err != nil {
			writeErrLine(writer, err, f)
		}
		file.Close()
	}
}

func writeErrLine(writer *csv.Writer, err error, info FileInfo) {
	line := []string{
		"E",
		err.Error(),
		strconv.FormatUint(info.Size, 10),
		info.Path,
	}

	if err := writer.Write(line); err != nil {
		slog.Warn("write error line error", "err", err)
	}
}

func writeSuccessLine(writer *csv.Writer, hash string, info FileInfo) error {
	line := []string{
		"H",
		hash,
		strconv.FormatUint(info.Size, 10),
		info.Path,
	}

	return writer.Write(line)
}
