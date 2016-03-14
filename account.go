package autosignr

type Account interface {
	Check(instanceId string) bool
	Type() string
}
