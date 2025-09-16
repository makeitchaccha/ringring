package rule

type Manager struct {
	Repository Repository
}

func NewManager(repository Repository) Manager {
	return Manager{
		Repository: repository,
	}
}


