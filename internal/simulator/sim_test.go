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
