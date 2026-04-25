package model

type Builder struct {
	name string
}

func (b *Builder) SetName(name string) *Builder {
	b.name = name
	return b
}
