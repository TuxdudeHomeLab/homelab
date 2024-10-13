package cmdexec

type Executor interface {
	Run(bin string, args ...string) (string, error)
}

func NewExecutor() Executor {
	return &executor{}
}
