package simulator

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jasonKoogler/cpu-sim/internal/config"
	"github.com/jasonKoogler/cpu-sim/internal/core"
)

// Statistics contains various metrics about the simulation
type Statistics struct {
	TotalCycles             int64
	InstructionsExecuted    int64
	IPC                     float64 // Instructions Per Cycle
	CacheHitRate            float64
	CoreUtilization         []float64
	MemoryAccessLatency     float64 // Average memory access latency
	InterconnectUtilization float64
}

// Simulator represents the multi-core processor simulator
type simulator struct {
	config     *config.Config
	cores      []*core.Processor
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

	// Initialize cores
	sim.cores = make([]*core.Processor, cfg.NumCores)
	for i := 0; i < cfg.NumCores; i++ {
		proc, err := core.NewProcessor(i, cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize core %d: %v", i, err)
		}
		sim.cores[i] = proc
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

	// for i := int64(0); i < cycles; i++ {
	// 	select {
	// 	case <-s.stopChan:
	// 		s.running.Store(false)
	// 		return nil
	// 	default:
	// 		atomic.AddInt64(&s.clock, 1)
	// 		s.simulateOneCycle()
	// 	}
	// }

	for _, proc := range s.cores {
		s.wg.Add(1)
		go func(p *core.Processor) {
			defer s.wg.Done()
			for i := int64(0); i < cycles; i++ {
				select {
				case <-s.stopChan:
					return
				default:
					p.Cycle()
				}
			}
		}(proc)
	}

	s.wg.Wait()
	s.running.Store(false)
	duration := time.Since(startTime)

	// s.statsMutex.Lock()
	// s.stats.TotalCycles = atomic.LoadInt64(&s.clock)
	// s.statsMutex.Unlock()

	s.calculateStatistics(cycles)

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

func (s *simulator) calculateStatistics(cycles int64) {
	s.statsMutex.Lock()
	defer s.statsMutex.Unlock()

	s.stats.TotalCycles = cycles

	totalInstructions := int64(0)
	for i, proc := range s.cores {
		instructions := proc.GetExecutedInstructions()
		totalInstructions += instructions

		// Update per-core utilizaiton
		s.stats.CoreUtilization[i] = proc.GetUtilization()
	}

	s.stats.InstructionsExecuted = totalInstructions

	// Calculate IPC (Instructions per Cycle per Core)
	if cycles > 0 {
		// Important! IPC is calculated by dividing the total instructions by the product of cycles and the number of cores
		s.stats.IPC = float64(totalInstructions) / float64(cycles*int64(len(s.cores)))
	}

	// TODO: other stats in the future
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
	s.wg.Wait()
	s.running.Store(false)
}

func (s *simulator) Reset() {
	s.statsMutex.Lock()
	defer s.statsMutex.Unlock()

	s.clock = 0
	s.running.Store(false)
	s.stopChan = make(chan struct{})

	// Reset Statistics
	for i := range s.stats.CoreUtilization {
		s.stats.CoreUtilization[i] = 0.0
	}
	s.stats.TotalCycles = 0
	s.stats.InstructionsExecuted = 0
	s.stats.IPC = 0.0
	s.stats.CacheHitRate = 0.0
	s.stats.MemoryAccessLatency = 0.0
	s.stats.InterconnectUtilization = 0.0

	// Reset Cores
	for _, proc := range s.cores {
		proc.Reset()
	}
}
