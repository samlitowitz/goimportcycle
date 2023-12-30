package source_file

import (
	"io/fs"
	"path/filepath"
)

type Emitter struct {
	includeNonGoFiles bool

	out chan<- string
}

func NewEmitter(opts ...Option) (*Emitter, <-chan string) {
	output := make(chan string)
	emitter := &Emitter{out: output}

	for _, opt := range opts {
		opt.apply(emitter)
	}

	return emitter, output
}

func (emitter *Emitter) WalkDirFunc(path string, d fs.DirEntry, err error) error {
	// any error stops all processing
	if err != nil {
		return err
	}

	// skip non-go files
	if !emitter.includeNonGoFiles && filepath.Ext(path) != ".go" {
		return nil
	}

	emitter.out <- path

	return nil
}

func (emitter *Emitter) Close() {
	close(emitter.out)
}

// https://uptrace.dev/blog/golang-functional-options.html

type Option interface {
	apply(emitter *Emitter)
}

type option func(emitter *Emitter)

func (fn option) apply(emitter *Emitter) {
	fn(emitter)
}
