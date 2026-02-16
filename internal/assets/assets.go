// Package assets provides asset loading utilities for the game.
// Assets are loaded from disk at runtime.
package assets

import (
	"io/fs"
	"os"
)

// AssetsDir is the default directory for game assets.
const AssetsDir = "assets"

// FS returns the filesystem for the assets directory.
// This uses the OS filesystem for runtime loading.
func FS() fs.FS {
	return os.DirFS(AssetsDir)
}

// SubFS returns a sub-filesystem rooted at the given path.
// This is useful for isolating specific asset directories like "sprites".
func SubFS(name string) (fs.FS, error) {
	return fs.Sub(FS(), name)
}
