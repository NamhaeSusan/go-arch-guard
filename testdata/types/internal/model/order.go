package model

type Order struct {
	name   string
	amount int
}

func (o *Order) SetName(name string) { o.name = name }

func (o *Order) Set(name string) { o.name = name }

func (o *Order) SetAmount(amount int) { o.amount = amount }

func (o *Order) setInternal(x int) { o.amount = x }

func (o Order) SetNothing(x int) {}

func (o *Order) SetReturningSelf(name string) *Order {
	o.name = name
	return o
}
