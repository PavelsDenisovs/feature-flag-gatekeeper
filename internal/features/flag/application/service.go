package application

type Repository interface {}

type Service interface {}

type service struct {
	repo Repository
}

func NewService(repo Repository) *service {
	if repo == nil {
    panic("nil repository")
  }
	return &service{repo: repo}
}
