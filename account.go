package autosignr

// Account defines the interface for checking cloud accounts
type Account interface {
	Init() error
	Check(instanceID string) bool
	Type() string
}
