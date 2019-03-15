package Pool

import "sync"

type Batch interface {
	Queue(fn WorkFunc)
	QueueComplete()
	Cancel()
	Results() <-chan WorkUnit
	WaitAll()
}

//Batch contains all information for a batch run of WorkUnits
type batch struct {
	pool    Pool
	m       sync.Mutex
	units   []WorkUnit // 工作单元的slice， 这个主要用在不设并发限制的场景，这里忽略
	results chan WorkUnit
	done    chan struct{} // 结果集,执行完后的workUnit会更新其value,error,可以从结果集channel中读取
	closed  bool
	wg      *sync.WaitGroup //内置的计数器，用于阻塞
}

func newBatch(p Pool) Batch {
	return &batch{
		pool:    p,
		units:   make([]WorkUnit, 0, 4),
		results: make(chan WorkUnit),
		done:    make(chan struct{}),
		wg:      new(sync.WaitGroup),
	}
}

//将协程池中要运行的工作排队，一旦所有的工作都已排队，一定要调用QueueComplete()
func (b *batch) Queue(fn WorkFunc) {
	b.m.Lock()

	if b.closed {
		b.m.Unlock()
		return
	}

	//插入工作
	wu := b.pool.Queue(fn)

	b.units = append(b.units, wu) //为取消工作单元留下引用
	b.wg.Add(1)
	b.m.Unlock()

	go func(b *batch, wu WorkUnit) {
		wu.Wait()
		b.results <- wu
		b.wg.Done()
	}(b, wu)
}

//让batch知道不在有工作单元排队，这样一旦所有的工作完成，它就可以关闭result通道
func (b *batch) QueueComplete() {
	b.m.Lock()
	b.closed = true
	close(b.done)
	b.m.Unlock()
}

// 取消属于此batch的工作单元
func (b *batch) Cancel() {
	b.QueueComplete()

	b.m.Lock()

	for i := len(b.units) - 1; i >= 0; i-- {
		b.units[i].Cancel()
	}
	b.m.Unlock()
}

func (b *batch) Results() <-chan WorkUnit {

	go func(b *batch) {
		<-b.done
		b.m.Lock()
		b.wg.Wait()
		b.m.Unlock()
		close(b.results)
	}(b)

	return b.results
}

func (b *batch) WaitAll() {

	for range b.Results() {
	}
}
