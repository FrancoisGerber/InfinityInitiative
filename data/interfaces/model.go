package interfaces

type Model interface {
	Migrate() error
	Create() (int64, error)
	Update() (int64, error)
	Delete() (int64, error)
}
