package entity

type Product struct {
	id int
}

func (p Product) GetId() int {
	return p.id
}