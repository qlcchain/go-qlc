package node

type TaskStatus int

const (
	_ TaskStatus = iota
	Success
	Failed
)

//Task action and status
type Task interface {
	Init() error
	Start() error
	Stop() error
	State() TaskStatus
}
