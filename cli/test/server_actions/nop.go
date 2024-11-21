package server_actions

type ServerActionNop struct {
	ServerActionTest
	EmptyCleanup
}

func (t *ServerActionNop) Skip() (bool, string) {
	return true, "skip nop action"
}
func (t ServerActionNop) Start() error { return nil }
