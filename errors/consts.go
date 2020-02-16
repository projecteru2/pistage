package errors

var (
	// ErrInvalidType .
	ErrInvalidType = New("invalid data type")
	// ErrInvalidValue .
	ErrInvalidValue = New("invalid value")

	// ErrNoSuchContainer .
	ErrNoSuchContainer = New("no such container")

	// ErrKeyExists .
	ErrKeyExists = New("key is exists")
	// ErrKeyNotExists .
	ErrKeyNotExists = New("key is not exists")
	// ErrKeyBadVersion .
	ErrKeyBadVersion = New("bad version")

	// ErrNotImplemented indicates the function isn't implemented.
	ErrNotImplemented = New("does not implemented")
)
