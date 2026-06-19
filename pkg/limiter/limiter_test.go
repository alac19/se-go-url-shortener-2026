package limiter

import (
	"sync"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

func TestAllow(t *testing.T) {
	t.Run("同一IP连续请求触发限流", func(t *testing.T) {
		lm := NewLimiterMap(rate.Every(50*time.Millisecond), 2) // 桶容量 2，每 50 ms 加 1 令牌

		ip := "192.168.1.1"
		var results []bool

		// 连续请求3次
		for i := 0; i < 3; i++ {
			results = append(results, lm.Allow(ip))
		}

		// 验证限流效果：前2次应通过，第3次被限流
		if results[0] != false {
			t.Errorf("第一次请求应该不被限流 (false), 实际得到 %v", results[0])
		}
		if results[1] != false {
			t.Errorf("第二次请求应该不被限流 (false), 实际得到 %v", results[1])
		}
		if results[2] != true {
			t.Errorf("第三次请求应该被限流 (true), 实际得到 %v", results[2])
		}
	})

	t.Run("不同IP限流器独立", func(t *testing.T) {
		lm := NewLimiterMap(rate.Every(50*time.Millisecond), 1)

		ip1 := "192.168.1.1"
		ip2 := "192.168.1.2"

		// 先让 ip1 用掉令牌
		lm.Allow(ip1)
		if got := lm.Allow(ip1); got != true {
			t.Errorf("ip1 第二次请求应被限流 (true), 实际得到 %v", got)
		}
		if got := lm.Allow(ip2); got != false {
			t.Errorf("ip2 第一次请求应不被限流 (false), 实际得到 %v", got)
		}
	})

	t.Run("等待后令牌恢复", func(t *testing.T) {
		lm := NewLimiterMap(rate.Every(50*time.Millisecond), 1)

		ip := "192.168.1.1"

		// 第一次请求，消耗令牌
		if got := lm.Allow(ip); got != false {
			t.Fatalf("第一次请求应该成功 (false), 实际得到 %v", got)
		}
		// 第二次请求立即进行，应该失败
		if got := lm.Allow(ip); got != true {
			t.Fatalf("第二次请求应该失败 (true), 实际得到 %v", got)
		}

		// 等待足够长时间让令牌恢复
		time.Sleep(60 * time.Millisecond)

		// 恢复后再次请求，应该成功
		if got := lm.Allow(ip); got != false {
			t.Errorf("等待后令牌应恢复, 请求应成功 (false), 实际得到 %v", got)
		}
	})

	t.Run("并发安全", func(t *testing.T) {
		lm := NewLimiterMap(rate.Every(12*time.Second), 5)

		var wg sync.WaitGroup

		for i := 0; i < 100; i++ {
			wg.Add(1)

			go func() {
				defer wg.Done()
				lm.Allow("1.2.3.4")
			}()
		}

		wg.Wait()
	})
}
