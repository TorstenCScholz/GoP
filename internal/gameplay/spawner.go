package gameplay

import (
	"fmt"

	"github.com/torsten/GoP/internal/entities"
	"github.com/torsten/GoP/internal/physics"
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
// Returns entities, triggers, solid entities, and kinematics separately for the caller to add to the world.
func SpawnEntities(objects []world.ObjectData, ctx SpawnContext) ([]entities.Entity, []entities.Trigger, []entities.SolidEntity, []physics.Kinematic) {
	var entityList []entities.Entity
	var triggers []entities.Trigger
	var solidEnts []entities.SolidEntity
	var kinematics []physics.Kinematic

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

		case world.ObjectTypePlatform:
			// Parse platform properties
			id := obj.GetPropString("id", obj.Name)
			if id == "" {
				id = fmt.Sprintf("platform_%d", obj.ID)
			}

			// endX and endY are relative offsets from the start position
			endXOffset := obj.GetPropFloat("endX", 0)
			endYOffset := obj.GetPropFloat("endY", 0)
			speed := obj.GetPropFloat("speed", 60)
			waitTime := obj.GetPropFloat("waitTime", 0.5)
			pushPlayer := obj.GetPropBool("pushPlayer", false)

			// Calculate absolute end position
			endX := obj.X + endXOffset
			endY := obj.Y + endYOffset

			platform := entities.NewMovingPlatform(id, obj.X, obj.Y, obj.W, obj.H, endX, endY, speed)
			platform.SetWaitTime(waitTime)
			platform.SetPushPlayer(pushPlayer)

			// Platform is both a solid entity and a kinematic
			solidEnts = append(solidEnts, platform)
			kinematics = append(kinematics, platform)
			entityList = append(entityList, platform)
		}
	}

	// Second pass: link switches to registry
	for _, sw := range switches {
		if sw.GetTargetID() != "" && ctx.Registry != nil {
			sw.SetRegistry(ctx.Registry)
		}
	}

	return entityList, triggers, solidEnts, kinematics
}
