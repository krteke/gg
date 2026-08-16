package internal

import (
	"encoding/csv"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type CompareConfig struct {
	Job    string
	Source string
	Target string
	Output string
}

func requireHashFile(path string) error {
	_, err := os.Stat(path)
	if err != nil {
		return Err(ErrOpenFile, "find hash file "+path, err)
	}

	return nil
}

func (c *CompareConfig) Compare() error {
	buckets, err := c.findCompareBuckets()
	if err != nil {
		return err
	}

	for _, bucket := range buckets {
		if err := requireHashFile(bucket.source); err != nil {
			return err
		}
		if err := requireHashFile(bucket.target); err != nil {
			return err
		}
	}

	if err := os.MkdirAll(c.Output, 0755); err != nil {
		return Err(ErrMkdir, "create directory "+c.Output, err)
	}

	if err := c.compareBuckets(buckets); err != nil {
		return err
	}

	return nil
}

func (c *CompareConfig) findCompareBuckets() ([]compareBucketPaths, error) {
	jobRoot := c.Job
	sourceRoot := c.Source
	targetRoot := c.Target

	var buckets []compareBucketPaths

	_, err := os.Stat(jobRoot)
	if err != nil {
		return nil, Err(ErrWalkDir, "walk job directory "+jobRoot, err)
	}

	err = filepath.WalkDir(jobRoot, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		path = filepath.Base(path)
		name := strings.TrimSuffix(path, filepath.Ext(path))
		hash := name + ".hash"
		buckets = append(buckets, compareBucketPaths{
			source: filepath.Join(sourceRoot, hash),
			target: filepath.Join(targetRoot, hash),
			output: name,
		})

		return nil
	})
	if err != nil {
		return nil, Err(ErrWalkDir, "walk job directory "+jobRoot, err)
	}

	return buckets, nil
}

func (c *CompareConfig) compareBuckets(buckets []compareBucketPaths) error {
	for _, bucket := range buckets {
		output := filepath.Join(c.Output, bucket.output)
		sourcePath := bucket.source
		targetPath := bucket.target

		source, err := os.Open(sourcePath)
		if err != nil {
			return Err(ErrOpenFile, "open hash file "+sourcePath, err)
		}
		defer source.Close()

		target, err := os.Open(targetPath)
		if err != nil {
			return Err(ErrOpenFile, "open hash file "+targetPath, err)
		}
		defer target.Close()

		diffs, err := compareHashStreams(
			newHashRecordReader(source, sourcePath),
			newHashRecordReader(target, targetPath),
		)
		if err != nil {
			return err
		}

		final := output + ".ok"
		stale := output + ".diff"
		if len(diffs) != 0 {
			final, stale = stale, final
		}

		if err := writeCompareResult(final, diffs); err != nil {
			return err
		}

	}
	return nil
}

func writeCompareResult(path string, diffs [][]string) error {
	result, err := os.Create(path)
	if err != nil {
		return Err(ErrCreateFile, "create file "+path, err)
	}
	defer result.Close()

	writer := csv.NewWriter(result)
	writer.Comma = '\t'
	for _, diff := range diffs {
		if err := writer.Write(diff); err != nil {
			return Err(ErrWriteFile, "write compare result "+path, err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return Err(ErrFlush, "flush "+path, err)
	}
	if err := result.Close(); err != nil {
		return Err(ErrFlush, "close "+path, err)
	}

	return nil
}

type compareBucketPaths struct {
	source string
	target string
	output string
}

type hashRecord struct {
	status string
	value  string
	size   string
	path   string
}

type hashRecordReader struct {
	reader *csv.Reader
	path   string
}

func newHashRecordReader(input io.Reader, path string) *hashRecordReader {
	reader := csv.NewReader(input)
	reader.Comma = '\t'
	reader.FieldsPerRecord = 4

	return &hashRecordReader{reader: reader, path: path}
}

func (r *hashRecordReader) read() (*hashRecord, error) {
	row, err := r.reader.Read()
	if err == io.EOF {
		return nil, nil
	}
	if err != nil {
		return nil, Err(ErrReadFile, "read hash file "+r.path, err)
	}

	return &hashRecord{
		status: row[0],
		value:  row[1],
		size:   row[2],
		path:   row[3],
	}, nil
}

func compareHashStreams(source, target *hashRecordReader) ([][]string, error) {
	var diffs [][]string

	sourceRecord, err := source.read()
	if err != nil {
		return nil, err
	}
	targetRecord, err := target.read()
	if err != nil {
		return nil, err
	}

	for sourceRecord != nil && targetRecord != nil {
		reason := ""
		switch {
		case sourceRecord.status == "E" || targetRecord.status == "E":
			reason = "ERROR"
		case sourceRecord.status == "N" || targetRecord.status == "N":
			reason = "MISSING"
		case sourceRecord.size != targetRecord.size:
			reason = "SIZE"
		case sourceRecord.value != targetRecord.value:
			reason = "HASH"
		}

		if reason != "" {
			diffs = append(diffs, []string{reason, sourceRecord.path})
		}

		sourceRecord, err = source.read()
		if err != nil {
			return nil, err
		}
		targetRecord, err = target.read()
		if err != nil {
			return nil, err
		}
	}

	return diffs, nil
}
