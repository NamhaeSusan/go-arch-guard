package repo

type User interface {
	FindByID(id int) error
}
