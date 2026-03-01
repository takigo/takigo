// Package config provides the runtime configuration table for widgets,
// supporting cget/configure operations. It ports the Tk_OptionSpec system
// from tk/generic/tkConfig.c to a Go-native approach.
package config

// OptionType identifies the kind of option value.
type OptionType int

const (
	TypeString OptionType = iota
	TypeInt
	TypeFloat
	TypeBool
	TypeColor
	TypeRelief
	TypeAnchor
	TypeJustify
	TypePixels
	TypeFont
	TypeCursor
	TypeBitmap
	TypeBorder
	TypeCommand
)

// Spec describes a single configuration option.
type Spec struct {
	Name         string     // option name, e.g. "background"
	Alias        string     // short name, e.g. "bg"
	Type         OptionType // expected value type
	DefaultValue string     // default as string
}
