package core

import (
	"testing"

	"github.com/jasonKoogler/cpu-sim/internal/config"
)

func TestNewProcessor(t *testing.T) {
	cfg := config.DefaultConfig()

	proc, err := NewProcessor(0, cfg)
	if err != nil {
		t.Fatalf("NewProcessor() error = %v", err)
	}

	if proc == nil {
		t.Fatal("NewProcessor() returned nil processor")
	}

	if proc.ID != 0 {
		t.Errorf("NewProcessor() processor ID = %d, want 0", proc.ID)
	}

	if proc.config != cfg {
		t.Errorf("NewProcessor() did not store the configuration")
	}

	// Check registers
	if len(proc.registersInt) != 32 {
		t.Errorf("NewProcessor() registersInt length = %d, want 32", len(proc.registersInt))
	}

	if len(proc.registersFloat) != 32 {
		t.Errorf("NewProcessor() registersFloat length = %d, want 32", len(proc.registersFloat))
	}

	// Check execution units
	if len(proc.executionUnits["ALU"]) < 1 {
		t.Errorf("NewProcessor() should have at least one ALU")
	}

	if len(proc.executionUnits["FPU"]) < 1 {
		t.Errorf("NewProcessor() should have at least one FPU")
	}

	if len(proc.executionUnits["LoadStore"]) < 1 {
		t.Errorf("NewProcessor() should have at least one LoadStore unit")
	}

	if len(proc.executionUnits["Branch"]) < 1 {
		t.Errorf("NewProcessor() should have at least one Branch unit")
	}
}

func TestNewProcessor_NilConfig(t *testing.T) {
	_, err := NewProcessor(0, nil)
	if err == nil {
		t.Fatal("NewProcessor() with nil config should return error")
	}
}

func TestNewProcessor_DifferentISAs(t *testing.T) {
	tests := []struct {
		name          string
		isa           string
		wantIntRegs   int
		wantFloatRegs int
		wantError     bool
	}{
		{
			name:          "RISC-V",
			isa:           "RISC-V",
			wantIntRegs:   32,
			wantFloatRegs: 32,
			wantError:     false,
		},
		{
			name:          "x86",
			isa:           "x86",
			wantIntRegs:   16,
			wantFloatRegs: 8,
			wantError:     false,
		},
		{
			name:          "ARM",
			isa:           "ARM",
			wantIntRegs:   16,
			wantFloatRegs: 32,
			wantError:     false,
		},
		{
			name:          "MIPS",
			isa:           "MIPS",
			wantIntRegs:   32,
			wantFloatRegs: 32,
			wantError:     false,
		},
		{
			name:          "Custom",
			isa:           "Custom",
			wantIntRegs:   32, // Default
			wantFloatRegs: 32, // Default
			wantError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.DefaultConfig()
			cfg.ISA = tt.isa

			proc, err := NewProcessor(0, cfg)
			if err != nil {
				if !tt.wantError {
					t.Fatalf("NewProcessor() error = %v", err)
				}
				return
			}

			if len(proc.registersInt) != tt.wantIntRegs {
				t.Errorf("NewProcessor() registersInt length = %d, want %d",
					len(proc.registersInt), tt.wantIntRegs)
			}

			if len(proc.registersFloat) != tt.wantFloatRegs {
				t.Errorf("NewProcessor() registersFloat length = %d, want %d",
					len(proc.registersFloat), tt.wantFloatRegs)
			}
		})
	}
}

func TestCycle(t *testing.T) {
	cfg := config.DefaultConfig()
	proc, _ := NewProcessor(0, cfg)

	// Run 100 cycles
	for i := 0; i < 100; i++ {
		proc.Cycle()
	}

	// Check cycle count
	if proc.cycleCount != 100 {
		t.Errorf("After 100 cycles, cycleCount = %d, want 100", proc.cycleCount)
	}

	// Check executed instructions (should be about 10 with our synthetic workload)
	executedInstructions := proc.GetExecutedInstructions()
	if executedInstructions != 10 {
		t.Errorf("After 100 cycles, executedInstructions = %d, want 10", executedInstructions)
	}

	// Check utilization
	utilization := proc.GetUtilization()
	expectedUtilization := float64(10) / float64(100)
	if utilization != expectedUtilization {
		t.Errorf("After 100 cycles, utilization = %f, want %f", utilization, expectedUtilization)
	}
}

func TestReset(t *testing.T) {
	cfg := config.DefaultConfig()
	proc, _ := NewProcessor(0, cfg)

	// Run some cycles
	for i := 0; i < 50; i++ {
		proc.Cycle()
	}

	// Modify some registers
	proc.registersInt[1] = 42
	proc.registersFloat[2] = 3.14

	// Reset the processor
	proc.Reset()

	// Check everything was reset
	if proc.cycleCount != 0 {
		t.Errorf("After Reset(), cycleCount = %d, want 0", proc.cycleCount)
	}

	if proc.GetExecutedInstructions() != 0 {
		t.Errorf("After Reset(), executedInstructions = %d, want 0", proc.GetExecutedInstructions())
	}

	if proc.registersInt[1] != 0 {
		t.Errorf("After Reset(), registersInt[1] = %d, want 0", proc.registersInt[1])
	}

	if proc.registersFloat[2] != 0.0 {
		t.Errorf("After Reset(), registersFloat[2] = %f, want 0.0", proc.registersFloat[2])
	}

	if proc.pc != 0 {
		t.Errorf("After Reset(), pc = %d, want 0", proc.pc)
	}

	// Check execution units are all not busy
	for _, units := range proc.executionUnits {
		for i, unit := range units {
			if unit.Busy {
				t.Errorf("After Reset(), executionUnits[%s][%d].Busy = true, want false", unit.Type, i)
			}
		}
	}
}
