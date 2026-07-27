// [[ when (modeIs "library" "cli-library") ]]
// Package [[.LibName]] is the public library surface for this module.
package [[.LibName]]

// Greet returns a short greeting for name.
func Greet(name string) string {
	if name == "" {
		return "hello"
	}
	return "hello, " + name
}
