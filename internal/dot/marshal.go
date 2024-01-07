package dot

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/samlitowitz/goimportcycle/internal/config"

	"github.com/samlitowitz/goimportcycle/internal"
)

func Marshal(cfg *config.Config, modulePath string, pkgs []*internal.Package) ([]byte, error) {
	buf := &bytes.Buffer{}

	writeHeader(buf, modulePath)
	switch cfg.Resolution {
	case config.FileResolution:
		writeNodeDefsForFileResolution(buf, cfg, pkgs)
		writeRelationshipsForFileResolution(buf, cfg, pkgs)
	case config.PackageResolution:
		writeNodeDefsForPackageResolution(buf, cfg, pkgs)
		writeRelationshipsForPackageResolution(buf, cfg, pkgs)
	}
	writeFooter(buf)

	return buf.Bytes(), nil
}

func writeHeader(buf *bytes.Buffer, modulePath string) {
	buf.WriteString(
		fmt.Sprintf(
			`
digraph {
	labelloc="t";
	label="%s";
	rankdir="TB";
	node [shape="rect"];
`,
			modulePath,
		),
	)
}

func writeFooter(buf *bytes.Buffer) {
	buf.WriteString(`
}
`,
	)
}

func pkgNodeName(pkg *internal.Package) string {
	return fmt.Sprintf(
		"pkg_%s",
		pkg.Name,
	)
}

func fileNodeName(file *internal.File) string {
	if file.Package == nil {
		return fmt.Sprintf(
			"file_%s",
			file.FileName,
		)
	}
	return fmt.Sprintf(
		"pkg_%s_file_%s",
		file.Package.Name,
		strings.TrimSuffix(file.FileName, ".go"),
	)
}
