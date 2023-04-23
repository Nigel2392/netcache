package cache

type errorType int

const (
	ErrNotError errorType = iota
	ErrItemNotFound
	ErrCacheAlreadyRunning
)

var errMap = map[errorType]string{
	ErrNotError:            "not a valid error",
	ErrItemNotFound:        "item not found",
	ErrCacheAlreadyRunning: "cache already running",
}

func (e errorType) Error() string {
	return errMap[e]
}

func (e errorType) Is(target error) bool {
	t, ok := target.(errorType)
	if !ok {
		return false
	}
	return t == e
}
