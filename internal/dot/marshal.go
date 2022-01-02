package dot

import (
	"bytes"
	"fmt"
	"github.com/samlitowitz/goimportcycleviz/internal"
	"github.com/samlitowitz/goimportcycleviz/internal/ast/pkg"
	"path/filepath"
)

func Marshal(files []*internal.File) ([]byte, error) {
	m := &marshaler{
		edges:                 make(map[nodeID]map[nodeID]struct{}),
		nodes:                 make(map[nodeID]pkg.FilePath),
		nodesByFilePath:       make(map[pkg.FilePath]nodeID),
		nodesBySubgraph:       make(map[subgraphID]map[nodeID]struct{}),
		subgraphs:             make(map[subgraphID]pkg.ImportPath),
		subgraphsByImportPath: make(map[pkg.ImportPath]subgraphID),
	}
	for _, file := range files {
		m.processFile(file, make(map[*internal.File]struct{}))
	}

	buf := &bytes.Buffer{}

	_, err := buf.WriteString("digraph g {\n")
	if err != nil {
		return nil, err
	}

	_, err = buf.WriteString("\tnode [shape=\"rect\"];\n")
	if err != nil {
		return nil, err
	}

	for node, label := range m.nodes {
		_, err = buf.WriteString(fmt.Sprintf("\t%s [label=\"%s\"];\n", node, label))
		if err != nil {
			return nil, err
		}
	}

	for subgraph, nodes := range m.nodesBySubgraph {
		_, err = buf.WriteString(fmt.Sprintf("\tsubgraph %s {\n", subgraph))
		if err != nil {
			return nil, err
		}

		subgraphImportPath, ok := m.subgraphs[subgraph]
		if !ok {
			subgraphImportPath = "UNKNOWN"
		}

		_, err = buf.WriteString(fmt.Sprintf("\t\tlabel=\"%s\";\n", subgraphImportPath))
		for node := range nodes {
			_, err = buf.WriteString(fmt.Sprintf("\t\t%s;\n", node))
			if err != nil {
				return nil, err
			}
		}
		_, err = buf.WriteString("\t}\n")
		if err != nil {
			return nil, err
		}
	}

	for node, toNodes := range m.edges {
		for toNode := range toNodes {
			_, err = buf.WriteString(fmt.Sprintf("\t%s -> %s;\n", node, toNode))
			if err != nil {
				return nil, err
			}
		}
	}

	_, err = buf.WriteString("}\n\n")
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type nodeID string
type subgraphID string

type marshaler struct {
	edges map[nodeID]map[nodeID]struct{}

	// nodes
	nodeCount       int
	nodes           map[nodeID]pkg.FilePath
	nodesByFilePath map[pkg.FilePath]nodeID
	nodesBySubgraph map[subgraphID]map[nodeID]struct{}

	// subgraphs
	subgraphCount         int
	subgraphs             map[subgraphID]pkg.ImportPath
	subgraphsByImportPath map[pkg.ImportPath]subgraphID
}

func (m *marshaler) processFile(f *internal.File, visited map[*internal.File]struct{}) nodeID {
	if _, ok := visited[f]; ok {
		node, ok := m.nodesByFilePath[f.Name]
		if !ok {
			return ""
		}
		return node
	}
	visited[f] = struct{}{}

	// ensure file node exists
	node, ok := m.nodesByFilePath[f.Name]
	if !ok {
		node = m.nextNodeID()
		_, fileName := filepath.Split(string(f.Name))
		m.nodes[node] = pkg.FilePath(fileName)
		m.nodesByFilePath[f.Name] = node
	}

	// ensure package subgraph exists
	subgraph, ok := m.subgraphsByImportPath[f.Package]
	if !ok {
		subgraph = m.nextSubGraphID()
		m.subgraphs[subgraph] = f.Package
		m.subgraphsByImportPath[f.Package] = subgraph
	}

	// ensure file node belongs to package subgraph
	if _, ok := m.nodesBySubgraph[subgraph]; !ok {
		m.nodesBySubgraph[subgraph] = make(map[nodeID]struct{})
	}
	if _, ok := m.nodesBySubgraph[subgraph][node]; !ok {
		m.nodesBySubgraph[subgraph][node] = struct{}{}
	}

	// ensure file node can have edges
	if _, ok := m.edges[node]; !ok {
		m.edges[node] = make(map[nodeID]struct{})
	}

	for _, importedFile := range f.Imports {
		importedNode := m.processFile(importedFile, visited)
		if importedNode == "" {
			continue
		}
		m.edges[node][importedNode] = struct{}{}
	}
	return node
}

func (m *marshaler) nextNodeID() nodeID {
	next := nodeID(fmt.Sprintf("n%d", m.nodeCount))
	m.nodeCount++
	return next
}

func (m *marshaler) nextSubGraphID() subgraphID {
	next := subgraphID(fmt.Sprintf("cluster%d", m.subgraphCount))
	m.subgraphCount++
	return next
}
