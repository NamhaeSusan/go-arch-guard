package mocks

type FakeDB struct{ err error }

func (d *FakeDB) SetErr(err error) { d.err = err }
