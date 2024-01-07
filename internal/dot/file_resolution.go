package dot

import (
	"bytes"
	"fmt"

	"github.com/samlitowitz/goimportcycle/internal"
	"github.com/samlitowitz/goimportcycle/internal/config"
)

func writeNodeDefsForFileResolution(buf *bytes.Buffer, cfg *config.Config, pkgs []*internal.Package) {
	clusterDefHeader := `
	subgraph cluster_%s {
		label="%s";
		style="filled";
		fontcolor="%s";
		fillcolor="%s";
`
	clusterDefFooter := `
	};
`
	nodeDef := `
		%s [label="%s", style="filled", fontcolor="%s", fillcolor="%s"];`

	for _, pkg := range pkgs {
		if pkg.IsStub {
			continue
		}
		if len(pkg.Files) == 0 {
			continue
		}
		pkgText := cfg.Palette.Base.PackageName
		pkgBackground := cfg.Palette.Base.PackageBackground
		if pkg.InImportCycle {
			pkgText = cfg.Palette.Cycle.PackageName
			pkgBackground = cfg.Palette.Cycle.PackageBackground
		}

		buf.WriteString(
			fmt.Sprintf(
				clusterDefHeader,
				pkgNodeName(pkg),
				pkg.ModuleRelativePath(),
				pkgText.Hex(),
				pkgBackground.Hex(),
			),
		)
		for _, file := range pkg.Files {
			if file.IsStub {
				continue
			}
			if len(file.Decls) == 0 {
				continue
			}
			fileText := cfg.Palette.Base.FileName
			fileBackground := cfg.Palette.Base.FileBackground
			if file.InImportCycle {
				fileText = cfg.Palette.Cycle.FileName
				fileBackground = cfg.Palette.Cycle.FileBackground
			}
			buf.WriteString(
				fmt.Sprintf(
					nodeDef,
					fileNodeName(file),
					file.FileName,
					fileText.Hex(),
					fileBackground.Hex(),
				),
			)
		}
		buf.WriteString(clusterDefFooter)
	}
}
