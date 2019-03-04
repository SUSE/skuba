package deployments

const (
	MasterRole = iota
	WorkerRole = iota
)

type Role int

type JoinConfiguration struct {
	Role Role
}
