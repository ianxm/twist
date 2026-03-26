package analysis

import (
	"slices"
	"twist/internal/api"
)

const maxBubble = 5

// Internal node used during bubble detection
type node struct {
	number     int
	warps      []*node
	inBubble   bool
	bubbleSize int
}

// floodFill fills from start, excluding warps to gateway and already-in-bubble nodes.
// Returns the filled nodes, or nil if the fill exceeds maxBubble.
func floodFill(start *node, gateway *node) []*node {
	inFill := make(map[*node]bool)
	queue := []*node{start}

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		if inFill[cur] {
			continue
		}
		inFill[cur] = true
		for _, w := range cur.warps {
			if w == gateway || inFill[w] {
				continue
			}
			queue = append(queue, w)
		}
		if len(inFill)+len(queue) > maxBubble {
			return nil
		}
	}

	filled := make([]*node, 0, len(inFill))
	for n := range inFill {
		filled = append(filled, n)
	}
	return filled
}

// FindBubbles finds all bubbles in the sector data.
// A bubble is a small cluster of sectors (≤ maxBubble) reachable through a single gateway.
func FindBubbles(sectors []api.SectorAnalysisView) []api.BubbleInfo {
	// Build adjacency graph from flat sector data
	nodes := make(map[int]*node, len(sectors))
	for _, s := range sectors {
		nodes[s.Number] = &node{number: s.Number}
	}
	for _, s := range sectors {
		n := nodes[s.Number]
		for _, warpNum := range s.Warps {
			if target, ok := nodes[warpNum]; ok {
				n.warps = append(n.warps, target)
			}
		}
	}

	var bubbles []api.BubbleInfo

	// Visit all nodes
	toVisit := make([]*node, 0, len(nodes))
	for _, n := range nodes {
		toVisit = append(toVisit, n)
	}

	for len(toVisit) > 0 {
		current := toVisit[0]
		toVisit = toVisit[1:]

		if current.inBubble {
			continue
		}

		// Perimeter: current's neighbors that are not in a bubble
		var perimeter []*node
		for _, w := range current.warps {
			if !w.inBubble {
				perimeter = append(perimeter, w)
			}
		}

		for _, perimNode := range perimeter {
			filled := floodFill(current, perimNode)
			if filled == nil {
				continue
			}
			// Bubble found — perimNode is the gateway
			for _, n := range filled {
				if n.bubbleSize > 0 {
					n.bubbleSize = 0
					bubbles = slices.DeleteFunc(bubbles, func(b api.BubbleInfo) bool {
						return n.number == b.GatewaySector
					})
				}
				n.inBubble = true
			}
			perimNode.bubbleSize = len(filled)
			bubbles = append(bubbles, api.BubbleInfo{
				Size:          len(filled),
				GatewaySector: perimNode.number,
			})
		}
	}

	return bubbles
}
