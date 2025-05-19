package pipeline

import (
	"fmt"
	"sync"
)

// Stage represents a stage in the processor pipeline
type Stage struct {
	Name        string
	Instruction *Instruction // Currently processing instruction
	Busy        bool
	Latency     int // cycles needed to complete this stage
}

// Pipeline represents the processor pipeline
type Pipeline struct {
	Stages []*Stage
	mutex  sync.RWMutex
}

// Instruction represents an instruction in the pipeline
type Instruction struct {
	Address    uint64
	Opcode     uint8
	Operands   []uint8
	Type       string // "Integer", "Float", "Memory", "Branch", "System"
	CyclesLeft int    // Cycles remaining in current stage
}

// NewPipeline creates a new pipeline with the specified depth
func NewPipeline(depth int, isa string) (*Pipeline, error) {
	if depth <= 0 {
		return nil, fmt.Errorf("pipeline depth must be positive")
	}

	pipeline := &Pipeline{
		Stages: make([]*Stage, 0, depth),
	}

	// Create stages based on ISA
	switch {
	case depth == 5 && (isa == "RISC-V" || isa == "MIPS"):
		// Classic 5-stage RISC pipeline
		pipeline.Stages = []*Stage{
			{Name: "Fetch", Busy: false, Latency: 1},
			{Name: "Decode", Busy: false, Latency: 1},
			{Name: "Execute", Busy: false, Latency: 1},
			{Name: "Memory", Busy: false, Latency: 1},
			{Name: "Writeback", Busy: false, Latency: 1},
		}
	case depth == 6 && isa == "x86":
		// Simplified x86 pipeline
		pipeline.Stages = []*Stage{
			{Name: "Fetch", Busy: false, Latency: 1},
			{Name: "Decode", Busy: false, Latency: 2}, // x86 decode is more complex
			{Name: "Issue", Busy: false, Latency: 1},
			{Name: "Execute", Busy: false, Latency: 1},
			{Name: "Memory", Busy: false, Latency: 1},
			{Name: "Writeback", Busy: false, Latency: 1},
		}
	case depth > 10 && isa == "x86":
		// Modern x86 deep pipeline (simplified model)
		pipeline.Stages = make([]*Stage, depth)
		pipeline.Stages[0] = &Stage{Name: "Fetch1", Busy: false, Latency: 1}
		pipeline.Stages[1] = &Stage{Name: "Fetch2", Busy: false, Latency: 1}
		pipeline.Stages[2] = &Stage{Name: "Decode1", Busy: false, Latency: 1}
		pipeline.Stages[3] = &Stage{Name: "Decode2", Busy: false, Latency: 1}
		pipeline.Stages[4] = &Stage{Name: "Decode3", Busy: false, Latency: 1}
		pipeline.Stages[5] = &Stage{Name: "Rename", Busy: false, Latency: 1}
		pipeline.Stages[6] = &Stage{Name: "Schedule", Busy: false, Latency: 1}
		pipeline.Stages[7] = &Stage{Name: "Dispatch", Busy: false, Latency: 1}
		pipeline.Stages[8] = &Stage{Name: "Execute", Busy: false, Latency: 1}
		pipeline.Stages[9] = &Stage{Name: "Memory", Busy: false, Latency: 1}
		pipeline.Stages[10] = &Stage{Name: "Writeback", Busy: false, Latency: 1}

		// Fill remaining stages if depth > 11
		for i := 11; i < depth; i++ {
			pipeline.Stages[i] = &Stage{
				Name:    fmt.Sprintf("ExtraStage%d", i-10),
				Busy:    false,
				Latency: 1,
			}
		}
	default:
		// Generic pipeline with specified depth
		pipeline.Stages = make([]*Stage, depth)

		// First and last stages are always Fetch and Writeback
		pipeline.Stages[0] = &Stage{Name: "Fetch", Busy: false, Latency: 1}
		pipeline.Stages[depth-1] = &Stage{Name: "Writeback", Busy: false, Latency: 1}

		// Middle stages depend on depth
		if depth == 3 {
			pipeline.Stages[1] = &Stage{Name: "Execute", Busy: false, Latency: 1}
		} else {
			// Add Decode after Fetch
			pipeline.Stages[1] = &Stage{Name: "Decode", Busy: false, Latency: 1}

			// Add Execute before Writeback
			pipeline.Stages[depth-2] = &Stage{Name: "Execute", Busy: false, Latency: 1}

			// Fill middle stages
			for i := 2; i < depth-2; i++ {
				var name string
				switch {
				case i == 2 && depth > 4:
					name = "Issue"
				case i == 3 && depth > 5:
					name = "Memory"
				default:
					name = fmt.Sprintf("Stage%d", i)
				}

				pipeline.Stages[i] = &Stage{Name: name, Busy: false, Latency: 1}
			}
		}
	}

	return pipeline, nil
}

// AdvanceStages moves instructions through the pipeline, returns true if any work was done
func (p *Pipeline) AdvanceStages() bool {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	workDone := false

	// Process stages in reverse order to avoid overwriting
	for i := len(p.Stages) - 1; i >= 0; i-- {
		stage := p.Stages[i]

		if stage.Busy && stage.Instruction != nil {
			workDone = true

			// Decrement cycles left in this stage
			stage.Instruction.CyclesLeft--

			// If instruction completed this stage
			if stage.Instruction.CyclesLeft <= 0 {
				// If this is the last stage, remove instruction from pipeline
				if i == len(p.Stages)-1 {
					stage.Instruction = nil
					stage.Busy = false
				} else {
					// Otherwise, try to pass to next stage
					nextStage := p.Stages[i+1]
					if !nextStage.Busy {
						// Move to next stage
						nextStage.Instruction = stage.Instruction
						nextStage.Busy = true
						nextStage.Instruction.CyclesLeft = nextStage.Latency

						// Clear current stage
						stage.Instruction = nil
						stage.Busy = false
					}
					// If next stage is busy, stall in current stage
				}
			}
		}
	}

	return workDone
}

// InsertInstruction inserts a new instruction into the first pipeline stage
func (p *Pipeline) InsertInstruction(inst *Instruction) bool {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Check if first stage is available
	if p.Stages[0].Busy {
		return false // Pipeline stalled
	}

	// Insert instruction
	p.Stages[0].Instruction = inst
	p.Stages[0].Busy = true
	inst.CyclesLeft = p.Stages[0].Latency

	return true
}

// IsFull checks if the pipeline is full (stalled)
func (p *Pipeline) IsFull() bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	return p.Stages[0].Busy
}

// IsEmpty checks if the pipeline is empty
func (p *Pipeline) IsEmpty() bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	for _, stage := range p.Stages {
		if stage.Busy {
			return false
		}
	}

	return true
}

// Flush clears all instructions from the pipeline
func (p *Pipeline) Flush() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	for _, stage := range p.Stages {
		stage.Instruction = nil
		stage.Busy = false
	}
}

// GetStages returns a copy of the pipeline stages (for observation)
func (p *Pipeline) GetStages() []*Stage {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	stagesCopy := make([]*Stage, len(p.Stages))
	for i, stage := range p.Stages {
		stageCopy := *stage // Make a copy of the stage
		stagesCopy[i] = &stageCopy
	}

	return stagesCopy
}

// GetCompletedInstructions returns the number of instructions that have completed execution
func (p *Pipeline) GetCompletedInstructions() int64 {
	// Not implemented in this version - will be tracked by processor
	return 0
}
