package Pool

import "sync/atomic"

type WorkUnit interface {
	Wait()
	Value() interface{}
	Error() error
	Cancel()
	IsCancelled() bool
}

//workUnit
type workUnit struct {
	value      interface{}
	err        error
	done       chan struct{}
	fn         WorkFunc
	cancelled  atomic.Value
	cancelling atomic.Value
	writing    atomic.Value
}

// Cancel cancels this specific unit of work, if not already committed to processing.
func (wu *workUnit) Cancel() {
	wu.cancelWithError(&ErrCancelled{s: errCancelled})
}

func (wu *workUnit) cancelWithError(err error) {
	//取消的过程中
	wu.cancelling.Store(struct{}{})

	//判断workunit是否正在写或者已经被取消
	if wu.writing.Load() == nil && wu.cancelled.Load() == nil {
		//将cancelled状态存储起来
		wu.cancelled.Store(struct{}{})
		wu.err = err
		close(wu.done)
	}
}

func (wu *workUnit) Wait() {
	<-wu.done
}

func (wu *workUnit) Value() interface{} {
	return wu.value
}

func (wu *workUnit) Error() error {
	return wu.err
}

func (wu *workUnit) IsCancelled() bool {
	// ensure that after this check we are committed as cannot be cancelled if not aalready
	wu.writing.Store(struct{}{})
	return wu.cancelled.Load() != nil
}
