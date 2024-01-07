.PHONY: clean

build:
	go build -o ${GOPATH}/bin/goimportcycle cmd/goimportcycle/main.go

examples: build
	./scripts/generate-examples.sh

clean:
	rm -f ${GOPATH}/bin/goimportcycle
