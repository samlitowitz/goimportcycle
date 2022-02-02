package file

import (
	"bytes"
	"fmt"
	"github.com/samlitowitz/goimportcycle/internal"
	"github.com/samlitowitz/goimportcycle/internal/ast"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
	"gonum.org/v1/gonum/graph/topo"
	"strings"
)

func Marshal(cfg *internal.Config, modulePath string, pkgs []*ast.Package) ([]byte, error) {
	buf := &bytes.Buffer{}
	buf.WriteString("digraph {\n")
	buf.WriteString("\tlabelloc=\"t\";\n")
	buf.WriteString("\tlabel=\"" + modulePath + "\";\n")
	buf.WriteString("\trankdir=\"TB\";\n")
	buf.WriteString("\tnode [shape=\"rect\"];\n")

	b := NewBuilder()
	b.Build(pkgs)

	// detect cycles in package graph
	// track all fileEdges by originating package node
	pkgGraph := b.PackageGraph()
	cycles := topo.DirectedCyclesIn(pkgGraph)
	cyclesByOrigin := make(map[int64]map[int64]struct{})
	for _, cycle := range cycles {
		for i := 0; i < len(cycle)-1; i++ {
			from := cycle[i].ID()
			to := cycle[i+1].ID()
			if _, ok := cyclesByOrigin[from]; !ok {
				cyclesByOrigin[from] = make(map[int64]struct{})
			}
			cyclesByOrigin[from][to] = struct{}{}
		}
	}

	fileGraph := b.FileGraph()
	for fromPkgNodeID, toPkgNodeIDs := range cyclesByOrigin {
		fromPkgNode := pkgGraph.Node(fromPkgNodeID).(*packageNode)
		if fromPkgNode == nil {
			continue
		}
		fromPkgNode.SetInCycle(true)
		for _, fromFile := range fromPkgNode.Package().Files {
			fromFileNodeID, ok := b.GetFileNodeID(fromFile)
			if !ok {
				continue
			}
			toFileNodeIDByPkgNodeID := make(map[int64]int64)
			toFileNodes := fileGraph.From(fromFileNodeID)
			for toFileNodes.Next() != false {
				toFileNode := toFileNodes.Node().(*fileNode)
				toFilePkgNodeID, ok := b.GetPackageNodeID(toFileNode.File().Package)
				if !ok {
					continue
				}
				toFileNodeIDByPkgNodeID[toFilePkgNodeID] = toFileNode.ID()
			}
			for toPkgNodeID := range toPkgNodeIDs {
				toFileNodeID, ok := toFileNodeIDByPkgNodeID[toPkgNodeID]
				if !ok {
					continue
				}
				curEdge := fileGraph.Edge(fromFileNodeID, toFileNodeID).(*fileEdge)
				if curEdge == nil {
					continue
				}
				curEdge.SetInCycle(true)
				fileGraph.SetEdge(curEdge)

				// update file graph nodes with inCycle
				fromFileNode := fileGraph.Node(fromFileNodeID).(*fileNode)
				if fromFileNode == nil {
					continue
				}
				toFileNode := fileGraph.Node(toFileNodeID).(*fileNode)
				if toFileNode == nil {
					continue
				}
				fromFileNode.SetInCycle(true)
				toFileNode.SetInCycle(true)
			}
		}
	}

	// Create package subgraphs and their child file nodes
	fileIDsByPkgID := b.GetFileIDsByPackageIDs()
	for pkgNodeID, fileNodeIDs := range fileIDsByPkgID {
		// create subgraph
		pkgNode := pkgGraph.Node(pkgNodeID).(*packageNode)
		if pkgNode == nil {
			continue
		}

		// reduce clutter by not showing packages with no files
		if len(fileNodeIDs) == 0 {
			continue
		}

		buf.WriteString(fmt.Sprintf("\tsubgraph cluster%d {\n", pkgNodeID))
		buf.WriteString(fmt.Sprintf("\t\tlabel=\"%s\";\n", packageLabel(modulePath, pkgNode.Package())))
		buf.WriteString("\t\tstyle=\"filled\";\n")
		pkgFillColor := cfg.PackageFill.Hex()
		if pkgNode.GetInCycle() {
			pkgFillColor = cfg.CyclePackageFill.Hex()
		}
		buf.WriteString(fmt.Sprintf("\t\tfillcolor=\"%s\";\n", pkgFillColor))

		for fileNodeID := range fileNodeIDs {
			// create nodes
			fileNode := fileGraph.Node(fileNodeID).(*fileNode)
			if fileNode == nil {
				continue
			}
			fileFillColor := cfg.FileFill.Hex()
			if fileNode.GetInCycle() {
				fileFillColor = cfg.CycleFileFill.Hex()
			}
			buf.WriteString(
				fmt.Sprintf(
					"\t\tn%d [label=\"%s\", style=\"filled\", fillcolor=\"%s\"];\n",
					fileNode.ID(),
					fileLabel(modulePath, fileNode.File()),
					fileFillColor,
				),
			)
			exports := fileNode.File().Exports
			if len(exports) == 0 {
				continue
			}
			names := make([]string, 0, len(exports))
			for name := range exports {
				names = append(names, name)
			}
			buf.WriteString(fmt.Sprintf("\t\t// Exports: %s\n", strings.Join(names, ", ")))
		}

		buf.WriteString("\t}\n")
	}

	// draw edges between files!
	fileEdges := fileGraph.Edges()
	for fileEdges.Next() != false {
		curEdge := fileEdges.Edge().(*fileEdge)
		edgeColor := cfg.Line.Hex()
		if curEdge.GetInCycle() {
			edgeColor = cfg.CycleLine.Hex()
		}
		buf.WriteString(
			fmt.Sprintf(
				"\tn%d -> n%d [color=\"%s\"];\n",
				curEdge.From().ID(),
				curEdge.To().ID(),
				edgeColor,
			),
		)
		references := curEdge.References()
		if len(references) == 0 {
			continue
		}
		buf.WriteString(fmt.Sprintf("\t// References: %s\n", strings.Join(references, ", ")))
	}
	buf.WriteString("}\n")

	return buf.Bytes(), nil
}

type packageEdge struct {
	from *packageNode
	to   *packageNode
}

func (e *packageEdge) From() graph.Node {
	return e.from
}

func (e *packageEdge) To() graph.Node {
	return e.to
}

func (e *packageEdge) ReversedEdge() graph.Edge {
	return &packageEdge{
		from: e.to,
		to:   e.from,
	}
}

type packageNode struct {
	id      int64
	pkg     *ast.Package
	inCycle bool
}

func (n *packageNode) ID() int64 {
	return n.id
}

func (n *packageNode) Package() *ast.Package {
	return n.pkg
}

func (n *packageNode) GetInCycle() bool {
	return n.inCycle
}

func (n *packageNode) SetInCycle(v bool) {
	n.inCycle = v
}

type fileEdge struct {
	from       *fileNode
	to         *fileNode
	references []string
	inCycle    bool
}

func (e *fileEdge) From() graph.Node {
	return e.from
}

func (e *fileEdge) To() graph.Node {
	return e.to
}

func (e *fileEdge) ReversedEdge() graph.Edge {
	return &fileEdge{
		from: e.to,
		to:   e.from,
	}
}

func (e *fileEdge) References() []string {
	return e.references
}

func (e *fileEdge) GetInCycle() bool {
	return e.inCycle
}

func (e *fileEdge) SetInCycle(v bool) {
	e.inCycle = v
}

type fileNode struct {
	id      int64
	file    *ast.File
	inCycle bool
}

func (n *fileNode) ID() int64 {
	return n.id
}

func (n *fileNode) File() *ast.File {
	return n.file
}

func (n *fileNode) GetInCycle() bool {
	return n.inCycle
}

func (n *fileNode) SetInCycle(v bool) {
	n.inCycle = v
}

type builder struct {
	next int64

	idsByPkg map[*ast.Package]int64
	pkgGraph *simple.DirectedGraph

	idsByFile map[*ast.File]int64
	fileGraph *simple.DirectedGraph

	fileIDsByPkgID map[int64]map[int64]struct{}
}

func NewBuilder() *builder {
	return &builder{
		next:           1,
		idsByPkg:       make(map[*ast.Package]int64),
		pkgGraph:       simple.NewDirectedGraph(),
		idsByFile:      make(map[*ast.File]int64),
		fileGraph:      simple.NewDirectedGraph(),
		fileIDsByPkgID: make(map[int64]map[int64]struct{}),
	}
}

func (b *builder) Build(pkgs []*ast.Package) {
	for _, pkg := range pkgs {
		pkgNode := b.buildPkg(pkg, make(map[*ast.Package]int64))
		if _, ok := b.fileIDsByPkgID[pkgNode.ID()]; !ok {
			b.fileIDsByPkgID[pkgNode.ID()] = make(map[int64]struct{})
		}
		for _, file := range pkg.Files {
			fileNode := b.buildFile(file, make(map[*ast.File]int64))
			b.fileIDsByPkgID[pkgNode.ID()][fileNode.ID()] = struct{}{}
		}
	}
}

func (b *builder) GetFileIDsByPackageIDs() map[int64]map[int64]struct{} {
	return b.fileIDsByPkgID
}

func (b *builder) GetFileNodeID(f *ast.File) (int64, bool) {
	id, ok := b.idsByFile[f]
	if ok {
		return id, true
	}
	return -1, false
}

func (b *builder) GetPackageNodeID(pkg *ast.Package) (int64, bool) {
	id, ok := b.idsByPkg[pkg]
	if ok {
		return id, true
	}
	return -1, false
}

func (b *builder) PackageGraph() *simple.DirectedGraph {
	return b.pkgGraph
}

func (b *builder) FileGraph() *simple.DirectedGraph {
	return b.fileGraph
}

func (b *builder) buildPkg(pkg *ast.Package, visited map[*ast.Package]int64) *packageNode {
	if curNodeID, ok := visited[pkg]; ok {
		return b.pkgGraph.Node(curNodeID).(*packageNode)
	}
	if curNodeID, ok := b.idsByPkg[pkg]; ok {
		return b.pkgGraph.Node(curNodeID).(*packageNode)
	}

	curNodeID := b.next
	b.next++
	curNode := &packageNode{
		id:  curNodeID,
		pkg: pkg,
	}
	b.idsByPkg[pkg] = curNode.id
	b.pkgGraph.AddNode(curNode)

	visited[pkg] = curNodeID

	for _, file := range pkg.Files {
		for _, fileImport := range file.Imports {
			importedPkgNode := b.buildPkg(fileImport.Package, visited)
			if importedPkgNode == nil {
				continue
			}
			b.pkgGraph.SetEdge(&packageEdge{
				from: curNode,
				to:   importedPkgNode,
			})
		}
	}

	delete(visited, pkg)
	return curNode
}

func (b *builder) buildFile(file *ast.File, visited map[*ast.File]int64) *fileNode {
	if curNodeID, ok := visited[file]; ok {
		return b.fileGraph.Node(curNodeID).(*fileNode)
	}
	if curNodeID, ok := b.idsByFile[file]; ok {
		return b.fileGraph.Node(curNodeID).(*fileNode)
	}

	curNodeID := b.next
	b.next++
	curNode := &fileNode{
		id:   curNodeID,
		file: file,
	}
	b.idsByFile[file] = curNode.id
	b.fileGraph.AddNode(curNode)

	visited[file] = curNodeID

	for _, imprt := range file.Imports {
		for reference := range imprt.References {
			for _, referencedFile := range imprt.Package.Files {
				if _, ok := referencedFile.Exports[reference]; ok {
					importedNode := b.buildFile(referencedFile, visited)

					importEdge := b.fileGraph.Edge(curNodeID, importedNode.ID())
					if importEdge == nil {
						b.fileGraph.SetEdge(&fileEdge{
							from:       curNode,
							to:         importedNode,
							inCycle:    false,
							references: []string{imprt.Package.Name + "." + reference},
						})
						break
					}
					ourImportEdge, ok := importEdge.(*fileEdge)
					if !ok {
						break
					}
					ourImportEdge.references = append(ourImportEdge.references, imprt.Package.Name+"."+reference)
					break
				}
			}
		}
	}

	delete(visited, file)
	return curNode
}

func fileLabel(modulePath string, file *ast.File) string {
	i := strings.LastIndex(file.Path, modulePath)
	if i == -1 {
		return file.Path
	}
	return file.Path[i+len(modulePath):]
}

func packageLabel(modulePath string, pkg *ast.Package) string {
	if pkg.Name == "main" {
		return "main"
	}
	i := strings.LastIndex(pkg.Path, modulePath)
	if i == -1 {
		return pkg.Path
	}
	return pkg.Path[i+len(modulePath):]
}
