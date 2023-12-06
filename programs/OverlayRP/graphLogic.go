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
	file, _ := os.OpenFile("graph.dot", os.O_TRUNC|os.O_WRONLY, 0644)
	defer file.Close()
	draw.DOT(rp.fullTree, file)
	rp.dotLock.Unlock()
}

func (rp *OverlayRP) copyGraph(g graph.Graph[string, string]) graph.Graph[string, string] {
	graph_copy := graph.New[string, string](graph.StringHash, graph.Weighted())
	rp.nodesLock.RLock()
	graph_copy.AddVerticesFrom(g)
	graph_copy.AddEdgesFrom(g)
	rp.nodesLock.RUnlock()
	return graph_copy
}

func (rp *OverlayRP) generatePathsForStream(as *availableStream) {
	as.mstreelock.Lock()
	graph_copy := rp.copyGraph(rp.overlayTree)
	rp.nodesLock.RLock()
	requesters := as.requesters
	for _, requester := range requesters {
		graph_copy.AddVertex(requester.node.Addr)
		err := graph_copy.AddEdge(requester.node.Addr, requester.node.neighbor, graph.EdgeWeight(int(requester.node.rtts)))
		if err != nil {
			rp.logger.Println("Error adding edge", err)
		}
	}
	rp.nodesLock.RUnlock()
	as.mstree, _ = graph.MinimumSpanningTree[string,string](graph_copy)
	rp.drawGraph(as.mstree)
	for _, requester := range requesters {
		path, err := graph.ShortestPath[string, string](as.mstree, rp.overlay.LocalAddr, requester.node.Addr)
		if err != nil {
			rp.logger.Println("Error generating path", err)
		}
		requester.graphPath = path
	}
	as.mstreelock.Unlock()
}