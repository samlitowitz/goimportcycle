# Go Import Cycle
[![Go Report Card](https://goreportcard.com/badge/github.com/samlitowitz/goimportcycle)](https://goreportcard.com/report/github.com/samlitowitz/goimportcycle)

`goimportcycle` is a tool to visualize Go imports resolved to the package or file level.

# Installation
`go install github.com/samlitowitz/goimportcycle/cmd/goimportcycle@v1.0.6`

# Usage
```shell
goimportcycle -path examples/simple/ -dot imports.dot
dot -Tpng -o assets/example.png imports.dot
```

![Example import graph resolved to the file level](assets/examples/simple/file.png?raw=true "Example import graph resolved to the file level")

Red lines indicate files causing import cycles between packages. Packages involved in a cycle have their backgrounds colored red.

```shell
goimportcycle -path examples/simple/ -dot imports.dot -resolution package
dot -Tpng -o assets/example.png imports.dot
```
![Example import graph resolved to the package level](assets/examples/simple/package.png?raw=true "Example import graph resolved to the package level")

Red lines indicate import cycles between packages.
