package simulator

import (
	"testing"
	"time"

	"github.com/jasonKoogler/cpu-sim/internal/config"
)

func TestNew(t *testing.T) {
	cfg := config.DefaultConfig()

	sim, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if sim == nil {
		t.Fatal("New() returned nil simulator")
	}

	if sim.config != cfg {
		t.Errorf("New() did not store the configuration")
	}

	if sim.running.Load() {
		t.Errorf("New() simulator should not be running initially")
	}

	if len(sim.stats.CoreUtilization) != cfg.NumCores {
		t.Errorf("New() stats.CoreUtilization length = %d, want %d",
			len(sim.stats.CoreUtilization), cfg.NumCores)
	}

	if len(sim.cores) != cfg.NumCores {
		t.Errorf("New() cores length = %d, want %d", len(sim.cores), cfg.NumCores)
	}

	for i, core := range sim.cores {
		if core == nil {
			t.Errorf("New() core[%d] is nil", i)
		}
	}
}

func TestNew_NilConfig(t *testing.T) {
	_, err := New(nil)
	if err == nil {
		t.Fatal("New() with nil config should return error")
	}
}

func TestRun(t *testing.T) {
	cfg := config.DefaultConfig()
	sim, _ := New(cfg)

	cycles := int64(100)
	err := sim.Run(cycles)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	stats := sim.GetStatistics()
	if stats.TotalCycles != int64(cycles) {
		t.Errorf("Run() TotalCycles = %d, want %d", stats.TotalCycles, cycles)
	}

	// With the pipeline implementation, each core should execute about cycles/5 instructions
	// (instructions are fetched every 5 cycles in the core's Cycle() method)
	expectedInstructions := int64(cycles / 5 * int64(cfg.NumCores))
	minInstructions := int64(float64(expectedInstructions) * 0.8)
	maxInstructions := int64(float64(expectedInstructions) * 1.2)
	if stats.InstructionsExecuted < minInstructions || stats.InstructionsExecuted > maxInstructions {
		t.Errorf("Run() InstructionsExecuted = %d, want approximately %d (between %d and %d)",
			stats.InstructionsExecuted, expectedInstructions, minInstructions, maxInstructions)
	}

	// IPC should be about 0.2 with the pipeline implementation (1 instruction every 5 cycles)
	expectedIPC := float64(0.2)
	if stats.IPC < expectedIPC*0.8 || stats.IPC > expectedIPC*1.2 {
		t.Errorf("Run() IPC = %f, want approximately %f", stats.IPC, expectedIPC)
	}

	// Each core should have higher utilization with the pipeline implementation
	// The pipeline stages advance each cycle, so utilization is higher
	for i, util := range stats.CoreUtilization {
		if util < 0.5 || util > 1.0 {
			t.Errorf("Run() CoreUtilization[%d] = %f, want between 0.5 and 1.0", i, util)
		}
	}
}

func TestRun_NegativeCycles(t *testing.T) {
	cfg := config.DefaultConfig()
	sim, _ := New(cfg)

	err := sim.Run(-10)
	if err == nil {
		t.Fatal("Run() with negative cycles should return error")
	}
}

func TestRun_AlreadyRunning(t *testing.T) {
	cfg := config.DefaultConfig()
	sim, _ := New(cfg)

	// Artificially set the running flag to true
	sim.running.Store(true)

	// Try to start a simulation while it's "running"
	err := sim.Run(100)
	if err == nil {
		t.Fatal("Run() while already running should return error")
	}

	// Clean up
	sim.running.Store(false)
}

func TestShutdown(t *testing.T) {
	cfg := config.DefaultConfig()
	sim, _ := New(cfg)

	// Create a channel to signal when the simulation has started
	started := make(chan struct{})

	// Start a long simulation in a goroutine
	go func() {
		// Set up a separate goroutine to signal when running is true
		go func() {
			for {
				if sim.running.Load() {
					close(started)
					return
				}
				time.Sleep(1 * time.Millisecond)
			}
		}()

		sim.Run(10000)
	}()

	// Wait for the simulation to start
	select {
	case <-started:
		// Simulation is now running
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Simulation failed to start within timeout")
	}

	// Verify it's running (should be true since we got the channel signal)
	if !sim.running.Load() {
		t.Fatal("Simulator should be running")
	}

	// Shut it down
	sim.Shutdown()

	// Give it time to stop
	time.Sleep(10 * time.Millisecond)

	// Verify it's stopped
	if sim.running.Load() {
		t.Fatal("Simulator should be stopped after Shutdown()")
	}
}

func TestReset(t *testing.T) {
	cfg := config.DefaultConfig()
	sim, _ := New(cfg)

	// Run a simulation
	sim.Run(100)

	// Verify that some stats were collected
	beforeStats := sim.GetStatistics()
	if beforeStats.TotalCycles == 0 || beforeStats.InstructionsExecuted == 0 {
		t.Fatal("Simulation should have generated some statistics")
	}

	// Reset the simulator
	sim.Reset()

	// Verify that stats were reset
	afterStats := sim.GetStatistics()
	if afterStats.TotalCycles != 0 {
		t.Errorf("After Reset(), TotalCycles = %d, want 0", afterStats.TotalCycles)
	}

	if afterStats.InstructionsExecuted != 0 {
		t.Errorf("After Reset(), InstructionsExecuted = %d, want 0", afterStats.InstructionsExecuted)
	}

	if afterStats.IPC != 0.0 {
		t.Errorf("After Reset(), IPC = %f, want 0.0", afterStats.IPC)
	}

	for i, util := range afterStats.CoreUtilization {
		if util != 0.0 {
			t.Errorf("After Reset(), CoreUtilization[%d] = %f, want 0.0", i, util)
		}
	}

	// Verify that all core pipelines are empty
	for i, core := range sim.cores {
		pipelineState := core.GetPipelineState()
		for j, stage := range pipelineState {
			if stage.Busy {
				t.Errorf("After Reset(), core[%d].pipeline.stage[%d].Busy = true, want false", i, j)
			}
			if stage.Instruction != nil {
				t.Errorf("After Reset(), core[%d].pipeline.stage[%d].Instruction is not nil", i, j)
			}
		}
	}

	// Run another simulation to verify the simulator still works
	err := sim.Run(50)
	if err != nil {
		t.Fatalf("Run() after Reset() error = %v", err)
	}

	// Verify that new stats were collected
	finalStats := sim.GetStatistics()
	if finalStats.TotalCycles != 50 {
		t.Errorf("After Reset() and Run(50), TotalCycles = %d, want 50", finalStats.TotalCycles)
	}
}

func TestPipelineIntegration(t *testing.T) {
	cfg := config.DefaultConfig()
	sim, _ := New(cfg)

	// Run a short simulation
	cycles := int64(20)
	err := sim.Run(cycles)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	// Check that instructions have moved through the pipeline
	for i, core := range sim.cores {
		// Get the pipeline state
		pipelineState := core.GetPipelineState()

		// Check that the pipeline has the expected number of stages
		if len(pipelineState) != cfg.PipelineDepth {
			t.Errorf("core[%d] pipeline has %d stages, want %d",
				i, len(pipelineState), cfg.PipelineDepth)
		}

		// At least some pipeline stages should have been active
		activeStageCounts := 0
		for _, stage := range pipelineState {
			if stage.Busy {
				activeStageCounts++
			}
		}

		// After running for 20 cycles, we expect the pipeline to have
		// at least a few busy stages if it's working properly
		if activeStageCounts == 0 {
			t.Errorf("core[%d] has no active pipeline stages after simulation", i)
		}
	}

	// Verify instruction execution
	stats := sim.GetStatistics()
	if stats.InstructionsExecuted == 0 {
		t.Errorf("No instructions executed during pipeline test")
	}

	// Higher core utilization is expected with pipeline integration
	for i, util := range stats.CoreUtilization {
		if util < 0.5 {
			t.Errorf("Core[%d] utilization = %f, expected > 0.5 with pipeline integration",
				i, util)
		}
	}
}
