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

	// With the synthetic workload, each core should execute about cycles/10 instructions
	expectedInstructions := int64(cycles / 10 * int64(cfg.NumCores))
	if stats.InstructionsExecuted != expectedInstructions {
		t.Errorf("Run() InstructionsExecuted = %d, want %d", stats.InstructionsExecuted, expectedInstructions)
	}

	// IPC should be about 0.1 with the synthetic workload
	expectedIPC := float64(0.1)
	if stats.IPC < expectedIPC*0.9 || stats.IPC > expectedIPC*1.1 {
		t.Errorf("Run() IPC = %f, want approximately %f", stats.IPC, expectedIPC)
	}

	// Each core should have ~10% utilization with the synthetic workload
	for i, util := range stats.CoreUtilization {
		if util < 0.09 || util > 0.11 {
			t.Errorf("Run() CoreUtilization[%d] = %f, want approximately 0.1", i, util)
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
