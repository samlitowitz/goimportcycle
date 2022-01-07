package file

import (
	"bytes"
	"fmt"
	"github.com/samlitowitz/goimportcycle/internal/ast"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
	"gonum.org/v1/gonum/graph/topo"
	"strings"
)

func Marshal(modulePath string, pkgs []*ast.Package) ([]byte, error) {
	buf := &bytes.Buffer{}
	buf.WriteString("digraph {\n")
	buf.WriteString("\tlabelloc=\"t\";\n")
	buf.WriteString("\tlabel=\"" + modulePath + "\";\n")
	buf.WriteString("\trankdir=\"TB\";\n")
	buf.WriteString("\tnode [shape=\"rect\"];\n")

	b := NewBuilder()
	b.Build(pkgs)

	nodes := b.Nodes()
	for nodes.Next() != false {
		curNode, ok := nodes.Node().(*node)
		if !ok {
			continue
		}
		buf.WriteString(fmt.Sprintf("\tn%d [label=\"%s\"];\n", curNode.ID(), fileLabel(modulePath, curNode.File())))
		exports := curNode.File().Exports
		if len(exports) == 0 {
			continue
		}
		names := make([]string, 0, len(exports))
		for name := range exports {
			names = append(names, name)
		}
		buf.WriteString(fmt.Sprintf("\t// Exports: %s\n", strings.Join(names, ", ")))
	}

	g := b.Graph()
	cycles := topo.DirectedCyclesIn(g)
	for _, cycle := range cycles {
		for i := 0; i < len(cycle)-1; i++ {
			curEdge := g.Edge(cycle[i].ID(), cycle[i+1].ID()).(*edge)
			if curEdge == nil {
				continue
			}
			curEdge.SetInCycle(true)
			g.SetEdge(curEdge)
		}
	}

	edges := b.Edges()
	for edges.Next() != false {
		curEdge := edges.Edge().(*edge)
		attrs := ""
		if curEdge.GetInCycle() {
			attrs = "[color=\"red\"]"
		}
		buf.WriteString(fmt.Sprintf("\tn%d -> n%d %s;\n", curEdge.From().ID(), curEdge.To().ID(), attrs))
		references := curEdge.References()
		if len(references) == 0 {
			continue
		}
		buf.WriteString(fmt.Sprintf("\t// References: %s\n", strings.Join(references, ", ")))
	}
	buf.WriteString("}\n")

	return buf.Bytes(), nil
}

type edge struct {
	from       *node
	to         *node
	references []string
	inCycle    bool
}

func (e *edge) From() graph.Node {
	return e.from
}

func (e *edge) To() graph.Node {
	return e.to
}

func (e *edge) ReversedEdge() graph.Edge {
	return &edge{
		from: e.to,
		to:   e.from,
	}
}

func (e *edge) References() []string {
	return e.references
}

func (e *edge) GetInCycle() bool {
	return e.inCycle
}

func (e *edge) SetInCycle(v bool) {
	e.inCycle = v
}

type node struct {
	id   int64
	file *ast.File
}

func (n *node) ID() int64 {
	return n.id
}

func (n *node) File() *ast.File {
	return n.file
}

type builder struct {
	next     int64
	idsByPkg map[*ast.File]int64
	g        *simple.DirectedGraph
}

func NewBuilder() *builder {
	return &builder{
		next:     1,
		idsByPkg: make(map[*ast.File]int64),
		g:        simple.NewDirectedGraph(),
	}
}

func (b *builder) Build(pkgs []*ast.Package) {
	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			b.build(file, make(map[*ast.File]int64))
		}
	}
}

func (b *builder) Graph() *simple.DirectedGraph {
	return b.g
}

func (b *builder) Edges() graph.Edges {
	return b.g.Edges()
}

func (b *builder) Nodes() graph.Nodes {
	return b.g.Nodes()
}

func (b *builder) build(file *ast.File, visited map[*ast.File]int64) *node {
	if curNodeID, ok := visited[file]; ok {
		return b.g.Node(curNodeID).(*node)
	}
	if curNodeID, ok := b.idsByPkg[file]; ok {
		return b.g.Node(curNodeID).(*node)
	}

	curNodeID := b.next
	b.next++
	curNode := &node{
		id:   curNodeID,
		file: file,
	}
	b.idsByPkg[file] = curNode.id
	b.g.AddNode(curNode)

	visited[file] = curNodeID

	for _, imprt := range file.Imports {
		for reference := range imprt.References {
			for _, referencedFile := range imprt.Package.Files {
				if _, ok := referencedFile.Exports[reference]; ok {
					importedNode := b.build(referencedFile, visited)

					importEdge := b.g.Edge(curNodeID, importedNode.ID())
					if importEdge == nil {
						b.g.SetEdge(&edge{
							from:       curNode,
							to:         importedNode,
							inCycle:    false,
							references: []string{imprt.Package.Name + "." + reference},
						})
						break
					}
					ourImportEdge, ok := importEdge.(*edge)
					if !ok {
						break
					}
					ourImportEdge.references = append(ourImportEdge.references, imprt.Package.Name + "." + reference)
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
