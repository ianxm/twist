package analysis

import "twist/internal/api"

// FindNearbyPorts performs a BFS from startSector and returns up to 20 ports
// found, ordered by distance.
func FindNearbyPorts(sectors []api.SectorAnalysisView, startSector int) []api.NearbyPortInfo {
	warps := make(map[int][]int, len(sectors))
	portClass := make(map[int]int, len(sectors))
	for _, s := range sectors {
		warps[s.Number] = s.Warps
		portClass[s.Number] = s.PortClass
	}

	type entry struct {
		sector   int
		distance int
	}

	var ports []api.NearbyPortInfo
	visited := make(map[int]bool)
	queue := []entry{{sector: startSector, distance: 0}}

	for len(queue) > 0 && len(ports) < 20 {
		e := queue[0]
		queue = queue[1:]

		if visited[e.sector] {
			continue
		}
		visited[e.sector] = true

		if pc := portClass[e.sector]; pc > 0 && e.sector != startSector {
			ports = append(ports, api.NearbyPortInfo{Sector: e.sector, Class: pc, Distance: e.distance})
			if len(ports) >= 20 {
				break
			}
		}

		for _, w := range warps[e.sector] {
			if !visited[w] {
				queue = append(queue, entry{sector: w, distance: e.distance + 1})
			}
		}
	}

	return ports
}
