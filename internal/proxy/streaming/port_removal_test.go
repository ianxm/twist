package streaming

import (
	"testing"
	"twist/internal/proxy/database"
)

func TestPortRemovalOnSectorVisitWithoutPort(t *testing.T) {
	db := database.NewDatabase()
	if err := db.CreateDatabase(":memory:"); err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.CloseDatabase()

	// Pre-populate a port in sector 100
	port := database.TPort{Name: "Stale Port", ClassIndex: 3}
	if err := db.SavePort(port, 100); err != nil {
		t.Fatalf("Failed to save port: %v", err)
	}

	// Verify port exists
	loaded, err := db.LoadPort(100)
	if err != nil {
		t.Fatalf("Failed to load port: %v", err)
	}
	if loaded.Name != "Stale Port" {
		t.Fatalf("Expected port 'Stale Port', got '%s'", loaded.Name)
	}

	parser := NewTWXParser(func() database.Database { return db }, nil)

	// Process a sector display for sector 100 WITHOUT a port line
	parser.ProcessString("Sector  : 100 in The Federation.\r")
	parser.ProcessString("Warps to Sector(s) :  101 - 102\r")
	// Command prompt triggers sectorCompleted
	parser.ProcessString("Command [TL=00:15:00]:[100] (?=Help)? : \r")

	// Verify port was deleted
	loaded, err = db.LoadPort(100)
	if err != nil {
		t.Fatalf("Failed to load port after deletion: %v", err)
	}
	if loaded.Name != "" {
		t.Errorf("Expected port to be deleted, but got '%s'", loaded.Name)
	}
}

func TestPortPreservedOnSectorVisitWithPort(t *testing.T) {
	db := database.NewDatabase()
	if err := db.CreateDatabase(":memory:"); err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.CloseDatabase()

	// Pre-populate a port in sector 200
	port := database.TPort{Name: "Existing Port", ClassIndex: 2}
	if err := db.SavePort(port, 200); err != nil {
		t.Fatalf("Failed to save port: %v", err)
	}

	parser := NewTWXParser(func() database.Database { return db }, nil)

	// Process a sector display for sector 200 WITH a port line
	parser.ProcessString("Sector  : 200 in The Federation.\r")
	parser.ProcessString("Ports   : Updated Port, Class 5 Port SBS\r")
	parser.ProcessString("Warps to Sector(s) :  201 - 202\r")
	parser.ProcessString("Command [TL=00:15:00]:[200] (?=Help)? : \r")

	// Verify port still exists (updated, not deleted)
	loaded, err := db.LoadPort(200)
	if err != nil {
		t.Fatalf("Failed to load port: %v", err)
	}
	if loaded.Name == "" {
		t.Errorf("Expected port to still exist, but it was deleted")
	}
}
