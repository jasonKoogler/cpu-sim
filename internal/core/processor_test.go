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

	// Check pipeline
	if proc.pipeline == nil {
		t.Errorf("NewProcessor() should create a pipeline")
	}

	stages := proc.pipeline.GetStages()
	if len(stages) != cfg.PipelineDepth {
		t.Errorf("Pipeline depth = %d, want %d", len(stages), cfg.PipelineDepth)
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
			wantError:     false,
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

	// Check executed instructions (should be about 20 with the synthetic pipeline workload)
	executedInstructions := proc.GetExecutedInstructions()
	if executedInstructions < 10 || executedInstructions > 30 {
		t.Errorf("After 100 cycles, executedInstructions = %d, want between 10 and 30",
			executedInstructions)
	}

	// Check utilization
	utilization := proc.GetUtilization()
	if utilization < 0.1 || utilization > 1.0 {
		t.Errorf("After 100 cycles, utilization = %f, should be between 0.1 and 1.0",
			utilization)
	}
}

func TestCycle_PipelineFlow(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.PipelineDepth = 5 // Ensure 5-stage pipeline

	proc, _ := NewProcessor(0, cfg)

	// Run a few cycles to prime the pipeline
	for i := 0; i < 20; i++ {
		proc.Cycle()
	}

	// Check that pipeline has some activity
	stages := proc.GetPipelineState()

	// Count busy stages
	busyStages := 0
	for _, stage := range stages {
		if stage.Busy {
			busyStages++
		}
	}

	// Pipeline should have at least one busy stage
	if busyStages == 0 {
		t.Errorf("Pipeline should have at least one busy stage")
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

	// Check that pipeline is empty
	stages := proc.GetPipelineState()
	for i, stage := range stages {
		if stage.Busy {
			t.Errorf("After Reset(), pipeline stage %d should not be busy", i)
		}
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

func TestGetPipelineState(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.PipelineDepth = 5
	proc, _ := NewProcessor(0, cfg)

	// Check initial pipeline state
	stages := proc.GetPipelineState()
	if len(stages) != 5 {
		t.Errorf("Initial pipeline stages length = %d, want 5", len(stages))
	}

	// All stages should be empty
	for i, stage := range stages {
		if stage.Busy {
			t.Errorf("Initial pipeline stage %d should not be busy", i)
		}

		if stage.Instruction != nil {
			t.Errorf("Initial pipeline stage %d should not have an instruction", i)
		}
	}

	// Run a few cycles to populate the pipeline
	for i := 0; i < 20; i++ {
		proc.Cycle()
	}

	// Get pipeline state again
	stages = proc.GetPipelineState()

	// At least one stage should be busy now
	busyStages := 0
	for _, stage := range stages {
		if stage.Busy {
			busyStages++
		}
	}

	if busyStages == 0 {
		t.Errorf("After 20 cycles, at least one pipeline stage should be busy")
	}
}
