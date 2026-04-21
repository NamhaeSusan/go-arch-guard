package model

type Order struct {
	name   string
	amount int
}

// flagged: classic setter
func (o *Order) SetName(name string) { o.name = name }

// flagged: literal Set
func (o *Order) Set(name string) { o.name = name }

// flagged: SetAmount
func (o *Order) SetAmount(amount int) { o.amount = amount }

// NOT flagged: unexported
func (o *Order) setInternal(x int) { o.amount = x }

// NOT flagged: value receiver
func (o Order) SetNothing(x int) {}

// NOT flagged: fluent builder (returns receiver type)
func (o *Order) SetReturningSelf(name string) *Order { o.name = name; return o }
