package Pool

const (
	errCancelled = "ERROR: Work Unit Cancelled"
	errRecovery  = "ERROR: Work Unit failed due to a recoverable error: '%v'\n, Stack Trace:\n %s"
	errClosed    = "ERROR: Work Unit added/run after the pool had been closed or cancelled"
)

type ErrCancelled struct {
	s string
}

func (e *ErrCancelled) Error() string {
	return e.s
}

type ErrRecovery struct {
	s string
}

func (e *ErrRecovery) Error() string {
	return e.s
}

type ErrClosed struct {
	s string
}

func (e *ErrClosed) Error() string {
	return e.s
}
