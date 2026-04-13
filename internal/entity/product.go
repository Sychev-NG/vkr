package entity

type ProductEntity struct {
	id int
}

func (p ProductEntity) GetId() int {
	return p.id
}