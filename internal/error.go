package internal

import (
	"errors"
	"fmt"
)

var (
	ErrOpenConfigFile = errors.New("open config file error")
	ErrDecodeConfig   = errors.New("decode config error")
	ErrWalkDir        = errors.New("walk dir error")
	ErrReadInfo       = errors.New("read file info error")
	ErrRelPath        = errors.New("get relative path error")
	ErrMkdir          = errors.New("create directory error")
	ErrCreateFile     = errors.New("create file error")
	ErrWriteFile      = errors.New("write file error")
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
