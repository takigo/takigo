package text

// MarkGravity determines which side of an insert a mark stays on.
type MarkGravity int

const (
	GravityLeft  MarkGravity = iota // mark stays before inserted text
	GravityRight                    // mark stays after inserted text (default for "insert")
)

// Mark is a named position in the document that is adjusted on edits.
type Mark struct {
	Name    string
	Pos     Index
	Gravity MarkGravity
}
