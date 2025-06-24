package loadshed

import (
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

// stat holds the latest CPU and memory usage values.
type stat struct {
	cpu int64 // store as hundredths of a percent, e.g., 82 = 82%
	mem int64
}

type Config struct {
	CPUThreshold float64
	MemThreshold float64
	Interval     time.Duration
}

// periodically update CPU and memory usage in the background
func startStatUpdater(s *stat, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			cpuPercents, err := cpu.Percent(0, false)
			var cpuVal int64
			if err == nil && len(cpuPercents) > 0 {
				cpuVal = int64(cpuPercents[0])
			}
			memStat, err := mem.VirtualMemory()
			var memVal int64
			if err == nil {
				memVal = int64(memStat.UsedPercent)
			}
			atomic.StoreInt64(&s.cpu, cpuVal)
			atomic.StoreInt64(&s.mem, memVal)
			// log.Printf("CPU usage: %.2f, Memory usage: %.2f", float64(cpuVal)/100, float64(memVal)/100)
		}
	}()
}

// Returns a Fiber middleware that sheds load if CPU or memory usage exceeds the given thresholds.
func NewLoadSheddingMiddleware(config Config) (fiber.Handler, error) {
	s := &stat{}
	startStatUpdater(s, config.Interval)

	if config.CPUThreshold < 0 || config.CPUThreshold > 100 {
		config.CPUThreshold = 0
		return nil, fmt.Errorf("CPU threshold must be between 0 and 100")
	}
	if config.MemThreshold < 0 || config.MemThreshold > 100 {
		config.MemThreshold = 0
		return nil, fmt.Errorf("memory threshold must be between 0 and 100")
	}

	cpuThresholdInt := int64(config.CPUThreshold * 100)
	memThresholdInt := int64(config.MemThreshold * 100)

	return func(c *fiber.Ctx) error {
		cpuUsage := atomic.LoadInt64(&s.cpu)
		memUsage := atomic.LoadInt64(&s.mem)

		if cpuUsage > cpuThresholdInt {
			log.Printf("Request rejected due to high CPU usage: %d, %d", cpuUsage, cpuThresholdInt)
			return c.Status(fiber.StatusServiceUnavailable).SendString("Service unavailable")
		}
		if memUsage > memThresholdInt {
			log.Printf("Request rejected due to high memory usage: %d, %d", memUsage, memThresholdInt)
			return c.Status(fiber.StatusServiceUnavailable).SendString("Service unavailable")
		}
		return c.Next()
	}, nil
}
