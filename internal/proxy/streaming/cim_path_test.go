package streaming

import (
	"testing"
	"twist/internal/proxy/database"
)

func TestCIMPathParsing(t *testing.T) {
	db := database.NewDatabase()
	if err := db.CreateDatabase(":memory:"); err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.CloseDatabase()

	parser := NewTWXParser(func() database.Database { return db }, nil)

	// Simulate CIM path output
	lines := []string{
		": ",
		"FM > 17282",
		"  TO > 24751",
		"17282 > (15925) > (29639) > (21050) > (2286) > (4288) > (20837) > (24678) >",
		" (24691) > (2924) > (5505) > (24703) > (19626) > (24360) > 18861 > (12243) >",
		" (23930) > 10029 > 27870 > 5222 > (952) > (24751)",
		": ",
	}

	for _, line := range lines {
		parser.ProcessString(line + "\r")
	}

	// Verify warp connections were saved
	// 17282 should warp to 15925
	sector, err := db.LoadSector(17282)
	if err != nil {
		t.Fatalf("Failed to load sector 17282: %v", err)
	}
	found := false
	for _, w := range sector.Warp {
		if w == 15925 {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Sector 17282 should have warp to 15925, got warps: %v", sector.Warp)
	}

	// 15925 should warp to 29639
	sector, err = db.LoadSector(15925)
	if err != nil {
		t.Fatalf("Failed to load sector 15925: %v", err)
	}
	found = false
	for _, w := range sector.Warp {
		if w == 29639 {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Sector 15925 should have warp to 29639, got warps: %v", sector.Warp)
	}

	// 952 should warp to 24751 (last pair)
	sector, err = db.LoadSector(952)
	if err != nil {
		t.Fatalf("Failed to load sector 952: %v", err)
	}
	found = false
	for _, w := range sector.Warp {
		if w == 24751 {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Sector 952 should have warp to 24751, got warps: %v", sector.Warp)
	}

	// Unexplored sectors should be marked EtCalc
	if sector.Explored != database.EtCalc {
		t.Errorf("Sector 952 should be EtCalc, got %d", sector.Explored)
	}
}
