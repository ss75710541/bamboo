package application

type Weight struct {
	ID       string         `param:"id" json:"id"`
	Versions map[string]int `param:"versions" json:"versions"`
}

type Storage interface {
	All() ([]Weight, error)
	Upsert(weight Weight) error
	Delete(ID string) error
}
