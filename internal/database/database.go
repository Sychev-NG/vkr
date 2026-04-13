package database

type DB interface{
	Query(q  string) error
	QueryRow(q  string) error
	Exec(q  string) error
}