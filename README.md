# Go Import Cycle
[![Go Report Card](https://goreportcard.com/badge/github.com/samlitowitz/goimportcycle)](https://goreportcard.com/report/github.com/samlitowitz/goimportcycle)

`goimportcycle` is a tool to visualize Go imports resolved to the file level.

# Installation
`go get -u github.com/samlitowitz/goimportcycle/cmd/goimportcycle`

# Usage
```shell
cd examples
goimportcycle -path examples/importcycle/ -dot imports.dot
dot -Tpng -o assets/example.png imports.dot
```

![Example `dot` output for the above example](assets/example.png?raw=true "Example import graph")

# Tasks that should get done
1. Display `"main"` package correctly

   ![Package `"main"` shown correctly](assets/tasks/display_main_package_correctly.png?raw=true)

2. Color edges involved in an import cycle differently from edges not involved in an import cycle

   ![Color import cycles](assets/tasks/color_import_cycles.png?raw=true)

3. Nest packages to reduce label length

   ![Nested packages](assets/tasks/nested_packages.png?raw=true)
