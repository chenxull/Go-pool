package basepool

import (
	"fmt"
	"sync"
)

type SimplePool struct {
	wg   sync.WaitGroup
	work chan func() //任务队列
}

//NewSimplePoll 创建协程池
func NewSimplePoll(workers int) *SimplePool {

	p := &SimplePool{
		wg:   sync.WaitGroup{},
		work: make(chan func()),
	}

	//同时并发数
	p.wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer func() {
				//捕获异常防止waitgroup阻塞
				if err := recover(); err != nil {
					fmt.Println(err)
					//每次结束都要释放锁
					p.wg.Done()
				}
			}()

			for fn := range p.work {
				fn()
			}
			p.wg.Done()
		}()
	}
	return p
}

// Add 增加任务
func (p *SimplePool) Add(fn func()) {
	p.work <- fn
}

//Run 运行任务
func (p *SimplePool) Run() {
	close(p.work)
	p.wg.Wait()
}
