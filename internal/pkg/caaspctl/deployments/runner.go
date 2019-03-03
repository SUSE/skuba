package deployments

type Runner interface {
	Run(Target) error
}

type State struct {
	DoRun func(t Target) error
}

func (s State) Run(t Target) error {
	return s.DoRun(t)
}
