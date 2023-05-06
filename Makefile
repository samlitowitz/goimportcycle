.PHONY: clean

build:
	go build -o ${GOPATH}/bin/goimportcycle cmd/goimportcycle/main.go

example: build
	${GOPATH}/bin/goimportcycle -path examples/importcycle/ -dot imports.dot

clean:
	rm ${GOPATH}/bin/goimportcycle