package builder

type Builder struct{ name string }

func New() *Builder { return &Builder{} }

// NOT flagged — returns *Builder (receiver type)
func (b *Builder) SetName(name string) *Builder { b.name = name; return b }
