package binder

// BindMode is used to determine how to bind to
// a struct tag.
type BindMode uint8

const (
	// ModeIgnoreCase allows case insensitivity when
	// binding configuration keys to struct tag matches.
	ModeIgnoreCase BindMode = 1

	// ModeStrict requires case sensitivity when binding
	// configuration keys to struct tag matches.
	ModeStrict BindMode = 2
)

// DefaultBindMode is the default set of flags used
// for struct tag bind mode, and includes ModeIgnoreCase.
const DefaultBindMode = ModeIgnoreCase

func (po BindMode) has(other BindMode) bool {
	return po&other != 0
}
