package deployments

type Runner interface {
	Run(t Target, data interface{}) error
}

type State struct {
	DoRun func(t Target, data interface{}) error
}

func (s State) Run(t Target, data interface{}) error {
	return s.DoRun(t, data)
}
