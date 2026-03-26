package analysis

import "twist/internal/api"

const maxPairDistance = 10

var validPairs = map[[2]int]bool{
	{1, 2}: true, {1, 3}: true, {1, 4}: true,
	{2, 3}: true, {2, 5}: true,
	{3, 6}: true,
	{4, 5}: true, {4, 6}: true,
	{5, 6}: true,
}

func isPair(a, b int) bool {
	if a > b {
		a, b = b, a
	}
	return validPairs[[2]int{a, b}]
}

type visitEntry struct {
	sector   int
	distance int
}

// FindPairs finds paired trading ports reachable from startSector via BFS,
// up to maxPairDistance hops away.
func FindPairs(sectors []api.SectorAnalysisView, startSector int) []api.PairInfo {
	warps := make(map[int][]int, len(sectors))
	portClass := make(map[int]int, len(sectors))
	for _, s := range sectors {
		warps[s.Number] = s.Warps
		portClass[s.Number] = s.PortClass
	}

	var pairs []api.PairInfo
	isDone := make(map[int]bool)
	toVisit := []visitEntry{{sector: startSector, distance: 0}}

	for len(toVisit) > 0 {
		entry := toVisit[0]
		toVisit = toVisit[1:]
		s := entry.sector

		if isDone[s] {
			continue
		}
		isDone[s] = true

		sClass := portClass[s]
		hasValidPort := sClass >= 1 && sClass <= 6

		for _, w := range warps[s] {
			if isDone[w] {
				continue
			}
			wClass := portClass[w]
			if hasValidPort && wClass >= 1 && wClass <= 6 && isPair(sClass, wClass) {
				pairs = append(pairs, api.PairInfo{
					Sector1:  s,
					Class1:   sClass,
					Sector2:  w,
					Class2:   wClass,
					Distance: entry.distance,
				})
			}
			if entry.distance+1 <= maxPairDistance {
				toVisit = append(toVisit, visitEntry{sector: w, distance: entry.distance + 1})
			}
		}
	}

	return pairs
}
