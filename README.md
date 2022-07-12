# Go Import Cycle
[![Go Report Card](https://goreportcard.com/badge/github.com/samlitowitz/goimportcycle)](https://goreportcard.com/report/github.com/samlitowitz/goimportcycle)

`goimportcycle` is a tool to visualize Go imports resolved to the package or file level.

# Installation
`go get -u github.com/samlitowitz/goimportcycle/cmd/goimportcycle`

# Usage
```shell
goimportcycle -path examples/importcycle/ -dot imports.dot
dot -Tpng -o assets/example.png imports.dot
```

![Example import graph resolved to the file level](assets/example_file.png?raw=true "Example import graph resolved to the file level")

Red lines indicate files causing import cycles between packages. Packages involved in a cycle have their backgrounds colored red.

```shell
goimportcycle -path examples/importcycle/ -dot imports.dot -resolution package
dot -Tpng -o assets/example.png imports.dot
```
![Example import graph resolved to the package level](assets/example_package.png?raw=true "Example import graph resolved to the package level")

Red lines indicate import cycles between packages.

# Tasks that probably should get done
1. Make output graphs nicely organized (vague)
