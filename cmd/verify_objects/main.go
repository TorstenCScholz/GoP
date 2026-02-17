package main

import (
	"fmt"
	"os"

	"github.com/torsten/GoP/internal/world"
)

func main() {
	// Read the level file
	data, err := os.ReadFile("assets/levels/level_01.json")
	if err != nil {
		fmt.Printf("Error reading level file: %v\n", err)
		os.Exit(1)
	}

	// Parse objects
	objects, err := world.ParseObjects(data)
	if err != nil {
		fmt.Printf("Error parsing objects: %v\n", err)
		os.Exit(1)
	}

	// Print summary
	fmt.Printf("Successfully parsed %d objects:\n\n", len(objects))

	// Count by type
	counts := make(map[world.ObjectType]int)
	for _, obj := range objects {
		counts[obj.Type]++
	}

	for typ, count := range counts {
		fmt.Printf("  %s: %d\n", typ, count)
	}

	// Verify we have all expected object types
	expected := map[world.ObjectType]int{
		world.ObjectTypeSpawn:      1,
		world.ObjectTypePlatform:   5,
		world.ObjectTypeSwitch:     5,
		world.ObjectTypeDoor:       4,
		world.ObjectTypeHazard:     4,
		world.ObjectTypeCheckpoint: 4,
		world.ObjectTypeGoal:       1,
	}

	fmt.Println("\nVerification:")
	allMatch := true
	for typ, expectedCount := range expected {
		actualCount := counts[typ]
		if actualCount == expectedCount {
			fmt.Printf("  [OK] %s: expected %d, got %d\n", typ, expectedCount, actualCount)
		} else {
			fmt.Printf("  [FAIL] %s: expected %d, got %d\n", typ, expectedCount, actualCount)
			allMatch = false
		}
	}

	if allMatch {
		fmt.Println("\n[SUCCESS] All objects parsed correctly! The visibility fix is working.")
	} else {
		fmt.Println("\n[FAILURE] Some objects were not parsed correctly.")
		os.Exit(1)
	}
}
