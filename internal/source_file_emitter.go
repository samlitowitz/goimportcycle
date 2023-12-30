package internal

import (
	"io/fs"
	"path/filepath"
)

type SourceFileEmitter struct {
	includeNonGoFiles bool

	output chan<- string
}

func NewSourceFileEmitter(opts ...Option) (*SourceFileEmitter, <-chan string) {
	output := make(chan string)
	emitter := &SourceFileEmitter{output: output}

	for _, opt := range opts {
		opt.apply(emitter)
	}

	return emitter, output
}

func (emitter *SourceFileEmitter) WalkDirFunc(path string, d fs.DirEntry, err error) error {
	// any error stops all processing
	if err != nil {
		return err
	}

	// skip non-go files
	if !emitter.includeNonGoFiles && filepath.Ext(path) != ".go" {
		return nil
	}

	emitter.output <- path

	return nil
}

func (emitter *SourceFileEmitter) Close() {
	close(emitter.output)
}

// https://uptrace.dev/blog/golang-functional-options.html

type Option interface {
	apply(emitter *SourceFileEmitter)
}

type option func(emitter *SourceFileEmitter)

func (fn option) apply(emitter *SourceFileEmitter) {
	fn(emitter)
}
