// Package app provides the main application structure and scene management.
package app

// Config holds application configuration settings.
type Config struct {
	WindowWidth  int
	WindowHeight int
	WindowTitle  string
	DebugMode    bool
}

// DefaultConfig returns a Config with sensible default values.
func DefaultConfig() *Config {
	return &Config{
		WindowWidth:  640,
		WindowHeight: 360,
		WindowTitle:  "Game",
		DebugMode:    false,
	}
}
