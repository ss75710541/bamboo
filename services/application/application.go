package application

type Application struct {
	ID     string `param:"id" json:"id"`
	Weight int    `param:"weight" json:"weight"`
}

type Storage interface {
	All() ([]Application, error)
	Upsert(app Application) error
	Delete(appId string) error
}
