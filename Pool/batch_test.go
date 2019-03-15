package Pool

import (
	"fmt"
	"testing"
	"time"
)

func SendMail(int int) WorkFunc {
	fn := func(wu WorkUnit) (interface{}, error) {
		// sleep 1s 模拟发邮件过程
		time.Sleep(time.Second * 1)
		// 模拟异常任务需要取消
		if int == 17 {
			wu.Cancel()
		}
		if wu.IsCancelled() {
			return false, nil
		}
		fmt.Println("send to", int)
		return true, nil
	}
	return fn
}

func TestBatchWork(t *testing.T) {
	// 初始化groutine数量为20的pool
	p := NewLimited(20)
	defer p.Close()
	batch := p.Batch()
	// 设置一个批量任务的过期超时时间
	a := time.After(10 * time.Second)
	go func() {
		for i := 0; i < 100; i++ {
			batch.Queue(SendMail(i))
		}
		batch.QueueComplete()
	}()
	// 因为 batch.Results 中要close results channel 所以不能将其放在LOOP中执行
	r := batch.Results()
LOOP:
	for {
		select {
		case <-a:
			// 登台超时通知
			fmt.Println("recived timeout")
			break LOOP

		case email, ok := <-r:
			// 读取结果集
			if ok {
				if err := email.Error(); err != nil {
					fmt.Println("err", err.Error())
				}
				fmt.Println(email.Value())
			} else {
				fmt.Println("finish")
				break LOOP
			}
		}
	}
}
