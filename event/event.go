package event

// RepoEvent stores relevant information about a repository when an event is received
type RepoEvent struct {
	Author string
	Branch string
	Commit string
}
