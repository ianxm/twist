package database

import (
	"fmt"
	"twist/internal/api"
)

// GetAllSectorAnalysisData returns lightweight sector data for all explored sectors,
// including warp connections and port class (via LEFT JOIN on ports).
func (d *SQLiteDatabase) GetAllSectorAnalysisData() ([]api.SectorAnalysisView, error) {
	if !d.dbOpen {
		return nil, fmt.Errorf("database not open")
	}

	query := `SELECT s.sector_index,
	                  s.warp1, s.warp2, s.warp3, s.warp4, s.warp5, s.warp6,
	                  COALESCE(p.class_index, 0)
	           FROM sectors s
	           LEFT JOIN ports p ON s.sector_index = p.sector_index
	           WHERE s.explored > 0`

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query sector analysis data: %w", err)
	}
	defer rows.Close()

	var result []api.SectorAnalysisView
	for rows.Next() {
		var sectorIndex int
		var warps [6]int
		var portClass int
		if err := rows.Scan(&sectorIndex,
			&warps[0], &warps[1], &warps[2], &warps[3], &warps[4], &warps[5],
			&portClass); err != nil {
			return nil, fmt.Errorf("failed to scan sector analysis data: %w", err)
		}
		var warpList []int
		for _, w := range warps {
			if w > 0 {
				warpList = append(warpList, w)
			}
		}
		result = append(result, api.SectorAnalysisView{
			Number:    sectorIndex,
			Warps:     warpList,
			PortClass: portClass,
		})
	}
	return result, rows.Err()
}
