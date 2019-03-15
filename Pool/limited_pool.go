package Pool

import (
	"fmt"
	"math"
	"runtime"
	"sync"
)

var _ Pool = new(limitedPool)

//协程池的数据结构
type limitedPool struct {
	workers uint           //并发量
	work    chan *workUnit //任务channel
	cancel  chan struct{}
	closed  bool
	m       sync.RWMutex
}

func NewLimited(workers uint) Pool {
	if workers == 0 {
		panic("invalid workers ‘0’")
	}

	p := &limitedPool{
		workers: workers,
	}
	p.initialize()
	return p

}

func (p *limitedPool) initialize() {
	p.work = make(chan *workUnit, p.workers*2)
	p.cancel = make(chan struct{})
	p.closed = false
	for i := 0; i < int(p.workers); i++ {
		//初始化并发单元
		p.newWorker(p.work, p.cancel)
	}
}

//将工作和取消channels传递给newworker以避免任何存在于p.work 读和写之间潜在的竞争
func (p *limitedPool) newWorker(work chan *workUnit, cancel chan struct{}) {
	go func(p *limitedPool) {
		var wu *workUnit

		//捕获异常，结束掉异常的工作单元，并将其再次作为新的任务启动
		defer func(p *limitedPool) {
			if err := recover(); err != nil {
				trace := make([]byte, 1<<16)
				n := runtime.Stack(trace, true)
				s := fmt.Sprintf(errRecovery, err, string(trace[:int(math.Min(float64(n), float64(7000)))]))

				iwu := wu
				iwu.err = &ErrRecovery{s: s}
				close(iwu.done)

				//启动新的worker去替代原来的
				p.newWorker(p.work, p.cancel)
			}
		}(p)

		var value interface{}
		var err error

		for {
			select {
			// workChannel中读取任务
			case wu = <-work:
				// 防止channel 被关闭后读取到零值
				if wu == nil {
					continue
				}
				//在执行任务前判断任务是否被取消
				if wu.cancelled.Load() == nil {
					value, err = wu.fn(wu)
					wu.writing.Store(struct{}{})

					// 任务执行完在写入结果时需要再次检查工作单元是否被取消，防止产生竞争条件
					if wu.cancelled.Load() == nil && wu.cancelling.Load() == nil {
						wu.value, wu.err = value, err
						close(wu.done)
					}
				}

			case <-cancel:
				return
			}
		}
	}(p)
}

//Queue queues the work to be run, and starts processing immediately
func (p *limitedPool) Queue(fn WorkFunc) WorkUnit {
	w := &workUnit{
		done: make(chan struct{}),
		fn:   fn,
	}
	go func() {
		p.m.RLock()
		if p.closed {
			w.err = &ErrClosed{s: errClosed}
			if w.cancelled.Load() == nil {
				close(w.done)
			}
			p.m.RUnlock()
			return
		}
		p.work <- w
		p.m.RUnlock()
	}()
	return w
}

func (p *limitedPool) Reset() {

	p.m.Lock()

	if !p.closed {
		p.m.Unlock()
		return
	}

	// cancelled the pool, not closed it, pool will be usable after calling initialize().
	p.initialize()
	p.m.Unlock()
}
func (p *limitedPool) closeWithError(err error) {

	p.m.Lock()

	if !p.closed {
		close(p.cancel)
		close(p.work)
		p.closed = true
	}

	for wu := range p.work {
		wu.cancelWithError(err)
	}

	p.m.Unlock()
}

func (p *limitedPool) Cancel() {
	err := &ErrCancelled{s: errCancelled}
	p.closeWithError(err)
}

func (p *limitedPool) Close() {
	err := &ErrClosed{s: errClosed}
	p.closeWithError(err)
}

func (p *limitedPool) Batch() Batch {
	return newBatch(p)
}
