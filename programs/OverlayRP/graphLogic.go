package main

import (
	"os"
	"time"

	"github.com/dominikbraun/graph"
	"github.com/dominikbraun/graph/draw"
)

func (rp *OverlayRP) addVertexToGraph(g graph.Graph[string, string], v string, neighbors []string, neighborsRTTs map[string]time.Duration) {
	g.AddVertex(v)
	for _, neighbor := range neighbors {
		err := g.AddEdge(v, neighbor, graph.EdgeWeight(int(neighborsRTTs[neighbor])))
		if err != nil {
			rp.logger.Println("Error adding edge", err)
		}
	}
}

func (rp *OverlayRP) removeVertexFromGraph(g graph.Graph[string, string], v string) {
	allEdges, _ := g.Edges()
	for _, edge := range allEdges {
		if edge.Source == v || edge.Target == v {
			g.RemoveEdge(edge.Source, edge.Target)
		}
	}
	g.RemoveVertex(v)
}

func (rp *OverlayRP) drawGraph(g graph.Graph[string, string]) {
	rp.dotLock.Lock()
	file, _ := os.OpenFile("graph.dot", os.O_CREATE|os.O_WRONLY, 0644)
	defer file.Close()
	draw.DOT(rp.fullTree, file)
	rp.dotLock.Unlock()
}