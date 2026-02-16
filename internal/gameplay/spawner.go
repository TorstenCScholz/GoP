package gameplay

import (
	"github.com/torsten/GoP/internal/entities"
	"github.com/torsten/GoP/internal/world"
)

// SpawnContext provides callbacks for entity spawning.
type SpawnContext struct {
	OnDeath       func()
	OnCheckpoint  func(id string, x, y float64)
	OnGoalReached func()
	Registry      *entities.TargetRegistry
}

// SpawnEntities creates entities from object data and returns them.
// Returns entities, triggers, and solid entities separately for the caller to add to the world.
func SpawnEntities(objects []world.ObjectData, ctx SpawnContext) ([]entities.Entity, []entities.Trigger, []entities.SolidEntity) {
	var entityList []entities.Entity
	var triggers []entities.Trigger
	var solidEnts []entities.SolidEntity

	// First pass: create all entities
	var switches []*entities.Switch

	for _, obj := range objects {
		switch obj.Type {
		case world.ObjectTypeHazard:
			hazard := entities.NewHazard(obj.X, obj.Y, obj.W, obj.H)
			hazard.OnDeath = ctx.OnDeath
			triggers = append(triggers, hazard)
			entityList = append(entityList, hazard)

		case world.ObjectTypeCheckpoint:
			id := obj.GetPropString("id", obj.Name)
			checkpoint := entities.NewCheckpoint(obj.X, obj.Y, obj.W, obj.H, id)
			checkpoint.OnActivate = ctx.OnCheckpoint
			triggers = append(triggers, checkpoint)
			entityList = append(entityList, checkpoint)

		case world.ObjectTypeGoal:
			goal := entities.NewGoal(obj.X, obj.Y, obj.W, obj.H)
			goal.OnComplete = ctx.OnGoalReached
			triggers = append(triggers, goal)
			entityList = append(entityList, goal)

		case world.ObjectTypeSwitch:
			// Check both "target" and "door_id" properties for flexibility
			targetID := obj.GetPropString("target", "")
			if targetID == "" {
				targetID = obj.GetPropString("door_id", "")
			}
			sw := entities.NewSwitch(obj.X, obj.Y, obj.W, obj.H, targetID)
			sw.SetToggleMode(obj.GetPropBool("toggle", true))
			sw.SetOnce(obj.GetPropBool("once", false))
			switches = append(switches, sw)
			triggers = append(triggers, sw)
			entityList = append(entityList, sw)

		case world.ObjectTypeDoor:
			id := obj.GetPropString("id", obj.Name)
			startOpen := obj.GetPropBool("startOpen", false)
			door := entities.NewDoor(obj.X, obj.Y, obj.W, obj.H, id)
			if startOpen {
				door.Open()
			}
			// Register door with registry if available
			if ctx.Registry != nil {
				ctx.Registry.Register(door)
			}
			solidEnts = append(solidEnts, door)
			entityList = append(entityList, door)
		}
	}

	// Second pass: link switches to registry
	for _, sw := range switches {
		if sw.GetTargetID() != "" && ctx.Registry != nil {
			sw.SetRegistry(ctx.Registry)
		}
	}

	return entityList, triggers, solidEnts
}
