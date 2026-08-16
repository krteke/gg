package internal

import (
	"errors"
	"fmt"
)

var (
	ErrOpenFile   = errors.New("open file error")
	ErrWalkDir    = errors.New("walk dir error")
	ErrReadInfo   = errors.New("read file info error")
	ErrRelPath    = errors.New("get relative path error")
	ErrMkdir      = errors.New("create directory error")
	ErrCreateFile = errors.New("create file error")
	ErrWriteFile  = errors.New("write file error")
	ErrReadFile   = errors.New("read file error")
	ErrParseUInt  = errors.New("parse uint error")
	ErrRename     = errors.New("file rename error")
	ErrOutofRange = errors.New("index out of range error")
	ErrFlush      = errors.New("flush error")
)

type CliError struct {
	Kind error
	Msg  string
	Err  error
}

func (e *CliError) Error() string {
	return fmt.Sprintf("%s: %v", e.Msg, e.Err)
}

func (e *CliError) Unwrap() error {
	return e.Err
}

func (e *CliError) Is(target error) bool {
	return target == e.Kind
}

func Err(k error, msg string, e error) *CliError {
	return &CliError{
		Kind: k,
		Msg:  msg,
		Err:  e,
	}
}
