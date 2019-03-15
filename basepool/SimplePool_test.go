package basepool

import (
	"fmt"
	"testing"
	"time"
)

func TestSimplePool(t *testing.T) {
	//设置并发数
	p := NewSimplePoll(2000)
	//将任务添加到协程池
	for i := 0; i < 1000000; i++ {
		p.Add(parseTask(i))
	}
	p.Run()
}

//模拟任务
func parseTask(i int) func() {
	return func() {
		time.Sleep(time.Second * 1)
		fmt.Println("finish parse", i)
	}
}
