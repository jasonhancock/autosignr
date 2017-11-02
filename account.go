package autosignr

type Account interface {
	Init() error
	Check(instanceId string) bool
	Type() string
}
