.PHONY: clean

build:
	go build -o ${GOPATH}/bin/goimportcycle cmd/goimportcycle/main.go

examples: build
	./scripts/examples.sh

clean:
	rm ${GOPATH}/bin/goimportcycle