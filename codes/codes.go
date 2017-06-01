package codes

const (
	// Nil indicates an unknown error. Should you encounter a Nil status, fret not! It will be reported to the server admin.
	Nil = iota
	// NotFound indicates that the requested resource could not be found, or the user is not allowed to view it.
	NotFound
	BadRequest
)