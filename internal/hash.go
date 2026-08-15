package internal

import (
	"encoding/csv"
	"encoding/hex"
	"io"
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
	Retry  *int
}

func (c *HashConfig) Hash() error {
	buckets, err := readTsv(c.Job)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(c.Output, 0755); err != nil {
		error := Err(ErrMkdir, "create directory "+c.Output, err)
		return error
	}

	hasher := blake3.New()

	if c.Retry != nil {
		retry := *c.Retry
		c.processBucket(hasher, buckets[retry], retry)
	}

	for i, bucket := range buckets[c.At:] {
		c.processBucket(hasher, bucket, i)
	}
	return nil
}

func (c *HashConfig) processBucket(hasher *blake3.Hasher, bucket []FileInfo, index int) error {
	output := path.Join(c.Output, strconv.Itoa(index)+".hash.partial")
	outputFile, err := os.OpenFile(output, os.O_CREATE|os.O_TRUNC|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		error := Err(ErrOpenFile, "open file"+output, err)
		return error
	}
	defer outputFile.Close()

	writer := csv.NewWriter(outputFile)
	writer.Comma = '\t'
	defer writer.Flush()

	c.hashBucket(writer, hasher, bucket)

	new := strings.TrimSuffix(output, filepath.Ext(output))
	err = os.Rename(output, new)
	if err != nil {
		error := Err(ErrRename, "rename "+output, err)
		return error
	}

	return nil
}

func (c *HashConfig) hashBucket(writer *csv.Writer, hasher *blake3.Hasher, bucket []FileInfo) {
	for _, f := range bucket {
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
