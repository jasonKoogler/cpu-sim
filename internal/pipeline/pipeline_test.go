package pipeline

import (
	"testing"
)

func TestNewPipeline(t *testing.T) {
	tests := []struct {
		name       string
		depth      int
		isa        string
		wantErr    bool
		wantStages int
	}{
		{
			name:       "Invalid depth",
			depth:      0,
			isa:        "RISC-V",
			wantErr:    true,
			wantStages: 0,
		},
		{
			name:       "RISC-V 5-stage",
			depth:      5,
			isa:        "RISC-V",
			wantErr:    false,
			wantStages: 5,
		},
		{
			name:       "x86 6-stage",
			depth:      6,
			isa:        "x86",
			wantErr:    false,
			wantStages: 6,
		},
		{
			name:       "Generic 3-stage",
			depth:      3,
			isa:        "Custom",
			wantErr:    false,
			wantStages: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewPipeline(tt.depth, tt.isa)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewPipeline() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return
			}

			if len(got.Stages) != tt.wantStages {
				t.Errorf("NewPipeline() got %d stages, want %d", len(got.Stages), tt.wantStages)
			}

			// Validate first and last stages
			if got.Stages[0].Name != "Fetch" && got.Stages[0].Name != "Fetch1" {
				t.Errorf("First stage should be Fetch or Fetch1, got %s", got.Stages[0].Name)
			}

			if tt.depth >= 3 && got.Stages[tt.depth-1].Name != "Writeback" {
				t.Errorf("Last stage should be Writeback, got %s", got.Stages[tt.depth-1].Name)
			}
		})
	}
}

func TestPipelineAdvance(t *testing.T) {
	pipe, err := NewPipeline(5, "RISC-V")
	if err != nil {
		t.Fatalf("Failed to create pipeline: %v", err)
	}

	// Empty pipeline should not do any work
	if pipe.AdvanceStages() {
		t.Errorf("AdvanceStages() on empty pipeline returned work done")
	}

	// Insert an instruction
	inst := &Instruction{
		Address:    0x1000,
		Opcode:     0x01,
		Operands:   []uint8{1, 2, 3},
		Type:       "Integer",
		CyclesLeft: 1,
	}

	inserted := pipe.InsertInstruction(inst)
	if !inserted {
		t.Fatalf("Failed to insert instruction")
	}

	// Advance pipeline - should do work
	if !pipe.AdvanceStages() {
		t.Errorf("AdvanceStages() with instruction returned no work done")
	}

	// Check if instruction moved to next stage
	if pipe.Stages[0].Busy {
		t.Errorf("First stage still busy after advancing")
	}

	if !pipe.Stages[1].Busy {
		t.Errorf("Second stage not busy after advancing")
	}

	// Run until pipeline is empty
	for i := 0; i < 10; i++ {
		pipe.AdvanceStages()
	}

	// Check if pipeline is empty
	if !pipe.IsEmpty() {
		t.Errorf("Pipeline should be empty after advancing multiple times")
	}
}

func TestPipelineMultiCycleStage(t *testing.T) {
	// Create a custom pipeline with a multi-cycle stage
	pipe := &Pipeline{
		Stages: []*Stage{
			{Name: "Fetch", Busy: false, Latency: 1},
			{Name: "Decode", Busy: false, Latency: 3}, // 3-cycle decode stage
			{Name: "Execute", Busy: false, Latency: 1},
			{Name: "Memory", Busy: false, Latency: 1},
			{Name: "Writeback", Busy: false, Latency: 1},
		},
	}

	// Insert an instruction
	inst := &Instruction{
		Address:    0x1000,
		Opcode:     0x01,
		Operands:   []uint8{1, 2, 3},
		Type:       "Integer",
		CyclesLeft: 1,
	}

	pipe.InsertInstruction(inst)

	// After 1 cycle, instruction should move to decode
	pipe.AdvanceStages()
	if !pipe.Stages[1].Busy || pipe.Stages[1].Instruction == nil {
		t.Fatalf("Instruction should be in decode stage")
	}

	// After 1 more cycle, instruction should still be in decode
	pipe.AdvanceStages()
	if !pipe.Stages[1].Busy || pipe.Stages[1].Instruction == nil {
		t.Fatalf("Instruction should still be in decode stage")
	}

	// After 1 more cycle, instruction should still be in decode
	pipe.AdvanceStages()
	if !pipe.Stages[1].Busy || pipe.Stages[1].Instruction == nil {
		t.Fatalf("Instruction should still be in decode stage")
	}

	// After 1 more cycle, instruction should move to execute
	pipe.AdvanceStages()
	if !pipe.Stages[2].Busy || pipe.Stages[2].Instruction == nil {
		t.Fatalf("Instruction should be in execute stage")
	}
}

func TestPipelineFlush(t *testing.T) {
	pipe, err := NewPipeline(5, "RISC-V")
	if err != nil {
		t.Fatalf("Failed to create pipeline: %v", err)
	}

	// Insert instructions in all stages
	for i := 0; i < 5; i++ {
		inst := &Instruction{
			Address:    uint64(0x1000 + i),
			Opcode:     uint8(i),
			Operands:   []uint8{1, 2, 3},
			Type:       "Integer",
			CyclesLeft: 1,
		}

		pipe.Stages[i].Instruction = inst
		pipe.Stages[i].Busy = true
	}

	// Verify pipeline is not empty
	if pipe.IsEmpty() {
		t.Fatalf("Pipeline should not be empty")
	}

	// Flush pipeline
	pipe.Flush()

	// Check if all stages are empty
	if !pipe.IsEmpty() {
		t.Errorf("Pipeline not empty after flush")
	}

	for i, stage := range pipe.Stages {
		if stage.Busy {
			t.Errorf("Stage %d still busy after flush", i)
		}

		if stage.Instruction != nil {
			t.Errorf("Stage %d still has instruction after flush", i)
		}
	}
}

func TestPipelineStall(t *testing.T) {
	pipe, err := NewPipeline(5, "RISC-V")
	if err != nil {
		t.Fatalf("Failed to create pipeline: %v", err)
	}

	// Create a stall by making a later stage busy
	pipe.Stages[2].Busy = true
	pipe.Stages[2].Instruction = &Instruction{
		Address:    0x2000,
		Opcode:     0x02,
		Operands:   []uint8{4, 5, 6},
		Type:       "Integer",
		CyclesLeft: 10, // Long-running instruction
	}

	// Insert an instruction in the first stage
	inst1 := &Instruction{
		Address:    0x1000,
		Opcode:     0x01,
		Operands:   []uint8{1, 2, 3},
		Type:       "Integer",
		CyclesLeft: 1,
	}

	inserted := pipe.InsertInstruction(inst1)
	if !inserted {
		t.Fatalf("Failed to insert first instruction")
	}

	// Advance the pipeline
	pipe.AdvanceStages()

	// Instruction should move to decode
	if !pipe.Stages[1].Busy || pipe.Stages[1].Instruction == nil {
		t.Fatalf("Instruction should be in decode stage")
	}

	// Try to advance again - decode should complete but execute is busy
	pipe.AdvanceStages()

	// Decode should still have the instruction (stalled)
	if !pipe.Stages[1].Busy || pipe.Stages[1].Instruction == nil {
		t.Fatalf("Instruction should still be in decode stage (stalled)")
	}

	// Try to insert another instruction
	inst2 := &Instruction{
		Address:    0x1004,
		Opcode:     0x03,
		Operands:   []uint8{7, 8, 9},
		Type:       "Integer",
		CyclesLeft: 1,
	}

	inserted = pipe.InsertInstruction(inst2)
	if !inserted {
		t.Fatalf("Failed to insert second instruction")
	}

	// Advance several times - everything should stall
	for i := 0; i < 5; i++ {
		pipe.AdvanceStages()
	}

	// Execute stage should still be busy with original instruction
	if !pipe.Stages[2].Busy || pipe.Stages[2].Instruction == nil {
		t.Fatalf("Execute stage should still be busy")
	}

	// Decode should still have the first inserted instruction
	if !pipe.Stages[1].Busy || pipe.Stages[1].Instruction == nil {
		t.Fatalf("Decode should still have the first instruction")
	}

	// Fetch should still have the second inserted instruction
	if !pipe.Stages[0].Busy || pipe.Stages[0].Instruction == nil {
		t.Fatalf("Fetch should still have the second instruction")
	}
}
