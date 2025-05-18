package simulator

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jasonKoogler/cpu-sim/internal/config"
)

type Statistics struct {
	TotalCycles             int64
	InstructionsExecuted    int64
	IPC                     float64 // Instructions Per Cycle
	CacheHitRate            float64
	CoreUtilization         []float64
	MemoryAccessLatency     float64 // Average memory access latency
	InterconnectUtilization float64
}

type simulator struct {
	config     *config.Config
	clock      int64
	running    atomic.Bool
	wg         sync.WaitGroup
	stopChan   chan struct{}
	stats      Statistics
	statsMutex sync.RWMutex
}

func New(cfg *config.Config) (*simulator, error) {
	if cfg == nil {
		return nil, fmt.Errorf("nil configuration provided")
	}

	sim := &simulator{
		config:   cfg,
		clock:    0,
		stopChan: make(chan struct{}),
		stats: Statistics{
			CoreUtilization: make([]float64, cfg.NumCores),
		},
	}

	return sim, nil
}

func (s *simulator) Run(cycles int64) error {
	if cycles <= 0 {
		return fmt.Errorf("cycle count must be greater than 0")
	}

	// Atomically check and set running flag
	if !s.running.CompareAndSwap(false, true) {
		return fmt.Errorf("simulation is already running")
	}

	startTime := time.Now()

	for i := int64(0); i < cycles; i++ {
		select {
		case <-s.stopChan:
			s.running.Store(false)
			return nil
		default:
			atomic.AddInt64(&s.clock, 1)
			s.simulateOneCycle()
		}
	}

	s.running.Store(false)
	duration := time.Since(startTime)

	s.statsMutex.Lock()
	s.stats.TotalCycles = atomic.LoadInt64(&s.clock)
	s.statsMutex.Unlock()

	fmt.Printf("Simulated %d cycles in %v (%.2f cycles/second)\n)", cycles, duration, float64(cycles)/duration.Seconds())
	fmt.Printf("\nSimulation Summary:\n")
	fmt.Printf("Total Cycles: %d\n", s.stats.TotalCycles)
	fmt.Printf("Instructions Executed: %d\n", s.stats.InstructionsExecuted)
	fmt.Printf("IPC: %.2f\n", s.stats.IPC)
	fmt.Printf("Cache Hit Rate: %.2f%%\n", s.stats.CacheHitRate*100)
	fmt.Printf("Core Utilization: %.2f%%\n", s.stats.CoreUtilization[0]*100)
	fmt.Printf("Memory Access Latency: %.2f cycles\n", s.stats.MemoryAccessLatency)

	return nil
}

func (s *simulator) simulateOneCycle() {
	// placeholder for now
}

func (s *simulator) GetStatistics() Statistics {
	s.statsMutex.RLock()
	defer s.statsMutex.RUnlock()

	statsCopy := Statistics{
		TotalCycles:             s.stats.TotalCycles,
		InstructionsExecuted:    s.stats.InstructionsExecuted,
		IPC:                     s.stats.IPC,
		CacheHitRate:            s.stats.CacheHitRate,
		CoreUtilization:         make([]float64, len(s.stats.CoreUtilization)),
		MemoryAccessLatency:     s.stats.MemoryAccessLatency,
		InterconnectUtilization: s.stats.InterconnectUtilization,
	}
	copy(statsCopy.CoreUtilization, s.stats.CoreUtilization)

	return statsCopy
}

func (s *simulator) Shutdown() {
	if !s.running.Load() {
		return
	}

	close(s.stopChan)
	s.running.Store(false)
}
