package pkg

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

			// update file graph nodes with inCycle
			fromFileNode := g.Node(cycle[i].ID()).(*node)
			if fromFileNode == nil {
				continue
			}
			toFileNode := g.Node(cycle[i+1].ID()).(*node)
			if toFileNode == nil {
				continue
			}
			fromFileNode.SetInCycle(true)
			toFileNode.SetInCycle(true)
		}
	}

	nodes := b.Nodes()
	for nodes.Next() != false {
		curNode, ok := nodes.Node().(*node)
		if !ok {
			continue
		}
		fillColor := cfg.PackageFill.Hex()
		if curNode.GetInCycle() {
			fillColor = cfg.CyclePackageFill.Hex()
		}

		buf.WriteString(fmt.Sprintf("\tn%d [label=\"%s\", style=\"filled\", fillcolor=\"%s\"];\n", curNode.ID(), packageLabel(modulePath, curNode.Package()), fillColor))
	}

	edges := b.Edges()
	for edges.Next() != false {
		curEdge := edges.Edge().(*edge)
		buf.WriteString(fmt.Sprintf("\tn%d -> n%d [color=\"%s\"];\n", curEdge.From().ID(), curEdge.To().ID(), cfg.Line.Hex()))
	}

	// create legend
	// file in cycle
	// pkg in cycle
	// edge in cycle
	buf.WriteString("}\n")

	return buf.Bytes(), nil
}

type edge struct {
	from    *node
	to      *node
	inCycle bool
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

func (e *edge) GetInCycle() bool {
	return e.inCycle
}

func (e *edge) SetInCycle(v bool) {
	e.inCycle = v
}

type node struct {
	id      int64
	pkg     *ast.Package
	inCycle bool
}

func (n *node) ID() int64 {
	return n.id
}

func (n *node) Package() *ast.Package {
	return n.pkg
}

func (n *node) GetInCycle() bool {
	return n.inCycle
}

func (n *node) SetInCycle(inCycle bool) {
	n.inCycle = inCycle
}

type builder struct {
	next     int64
	idsByPkg map[*ast.Package]int64
	g        *simple.DirectedGraph
}

func NewBuilder() *builder {
	return &builder{
		next:     1,
		idsByPkg: make(map[*ast.Package]int64),
		g:        simple.NewDirectedGraph(),
	}
}

func (b *builder) Build(pkgs []*ast.Package) {
	for _, pkg := range pkgs {
		b.build(pkg, make(map[*ast.Package]int64))
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

func (b *builder) build(pkg *ast.Package, visited map[*ast.Package]int64) *node {
	if curNodeID, ok := visited[pkg]; ok {
		return b.g.Node(curNodeID).(*node)
	}
	if curNodeID, ok := b.idsByPkg[pkg]; ok {
		return b.g.Node(curNodeID).(*node)
	}

	curNodeID := b.next
	b.next++
	curNode := &node{
		id:  curNodeID,
		pkg: pkg,
	}
	b.idsByPkg[pkg] = curNode.id
	b.g.AddNode(curNode)

	visited[pkg] = curNodeID

	for _, file := range pkg.Files {
		for _, fileImport := range file.Imports {
			importedPkgNode := b.build(fileImport.Package, visited)
			if importedPkgNode == nil {
				continue
			}
			b.g.SetEdge(&edge{
				from:    curNode,
				to:      importedPkgNode,
				inCycle: false,
			})
		}
	}

	delete(visited, pkg)
	return curNode
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
