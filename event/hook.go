package event

type Hook struct {
	Type    string
	Path    string
	Timeout int
	Cmd     []string
}
