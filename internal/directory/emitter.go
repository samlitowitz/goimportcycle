package directory

import (
	"io/fs"
)

type Emitter struct {
	out chan<- string
}

func NewEmitter() (*Emitter, <-chan string) {
	output := make(chan string)
	emitter := &Emitter{out: output}

	return emitter, output
}

func (emitter *Emitter) WalkDirFunc(path string, d fs.DirEntry, err error) error {
	// any error stops all processing
	if err != nil {
		return err
	}

	// skip individual files
	if !d.IsDir() {
		return nil
	}

	// emit directory path
	emitter.out <- path
	return nil
}

func (emitter *Emitter) Close() {
	close(emitter.out)
}
