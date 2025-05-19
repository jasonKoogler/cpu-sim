package core

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/jasonKoogler/cpu-sim/internal/config"
	"github.com/jasonKoogler/cpu-sim/internal/pipeline"
)

type ExecutionUnit struct {
	Type     string // "ALU", "FPU", "LoadStore", "Branch"
	Busy     bool   // true if the unit is currently executing an instruction
	Pipeline int    // number of stages in this unit
}

type Processor struct {
	ID                   int
	config               *config.Config
	pipeline             *pipeline.Pipeline
	instructionQueue     []Instruction
	executionUnits       map[string][]*ExecutionUnit
	registersInt         []uint64
	registersFloat       []float64
	pc                   uint64 // program counter
	executedInstructions int64
	cycleCount           int64
	busyCycles           int64
	mutex                sync.RWMutex
}

type Instruction struct {
	Address    uint64
	Opcode     uint8
	Operands   []uint8
	Type       string // "Integer", "Float", "Memory", "Branch", "System"
	Stage      string // Current pipeline stage
	CyclesLeft int    // Number of cycles left in the current stage
}

func NewProcessor(id int, cfg *config.Config) (*Processor, error) {
	if cfg == nil {
		return nil, fmt.Errorf("nil configuation provided")
	}

	pipe, err := pipeline.NewPipeline(cfg.PipelineDepth, cfg.ISA)
	if err != nil {
		return nil, fmt.Errorf("failed to create pipeline: %w", err)
	}

	var numIntRegs, numFloatRegs int
	switch cfg.ISA {
	case "RISC-V":
		numIntRegs = 32
		numFloatRegs = 32
	case "x86":
		numIntRegs = 16
		numFloatRegs = 8
	case "ARM":
		numIntRegs = 16
		numFloatRegs = 32
	case "MIPS":
		numIntRegs = 32
		numFloatRegs = 32
	default:
		numIntRegs = 32
		numFloatRegs = 32
	}

	proc := &Processor{
		ID:               id,
		config:           cfg,
		pipeline:         pipe,
		instructionQueue: make([]Instruction, 0, 32), // Default queue size
		registersInt:     make([]uint64, numIntRegs),
		registersFloat:   make([]float64, numFloatRegs),
		pc:               0,
		executionUnits:   make(map[string][]*ExecutionUnit),
	}

	// Initialize execution units
	// TODO: Make this configurable

	// ALUs (Arithmetic Logic Units)
	numALUs := 2
	proc.executionUnits["ALU"] = make([]*ExecutionUnit, numALUs)
	for i := 0; i < numALUs; i++ {
		proc.executionUnits["ALU"][i] = &ExecutionUnit{
			Type:     "ALU",
			Busy:     false,
			Pipeline: 1, // Simple ALU has one stage
		}
	}

	// FPUs (Floating Point Units)
	numFPUs := 1
	proc.executionUnits["FPU"] = make([]*ExecutionUnit, numFPUs)
	for i := 0; i < numFPUs; i++ {
		proc.executionUnits["FPU"][i] = &ExecutionUnit{
			Type:     "FPU",
			Busy:     false,
			Pipeline: 3, // FPU has 3 stages
		}
	}

	// LoadStore unit
	numLSUs := 1
	proc.executionUnits["LoadStore"] = make([]*ExecutionUnit, numLSUs)
	for i := 0; i < numLSUs; i++ {
		proc.executionUnits["LoadStore"][i] = &ExecutionUnit{
			Type:     "LoadStore",
			Busy:     false,
			Pipeline: 1, // LoadStore has 1 stage
		}
	}

	// Branch unit
	numBranches := 1
	proc.executionUnits["Branch"] = make([]*ExecutionUnit, numBranches)
	for i := 0; i < numBranches; i++ {
		proc.executionUnits["Branch"][i] = &ExecutionUnit{
			Type:     "Branch",
			Busy:     false,
			Pipeline: 1, // Branch has 1 stage
		}
	}

	return proc, nil
}

// Cycle executes a single processor cycle
func (p *Processor) Cycle() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	atomic.AddInt64(&p.cycleCount, 1)

	// Check if any work is being done in this cycle
	workDone := false

	// Process pipeline stages
	if p.pipeline.AdvanceStages() {
		workDone = true
	}

	// Fetch new instruction if pipeline can accept it
	if !p.pipeline.IsFull() && p.cycleCount%5 == 0 { // Fetch every 5 cycles (synthetic workload)
		inst := p.fetchNextInstruction()
		if inst != nil {
			pipelineInst := &pipeline.Instruction{
				Address:    inst.Address,
				Opcode:     inst.Opcode,
				Operands:   inst.Operands,
				Type:       inst.Type,
				CyclesLeft: 1,
			}

			if p.pipeline.InsertInstruction(pipelineInst) {
				workDone = true
			}
		}
	}

	// Count a completed instruction if one reached the end of the pipeline
	stages := p.pipeline.GetStages()
	if len(stages) > 0 && !stages[len(stages)-1].Busy && p.cycleCount%5 == 0 {
		atomic.AddInt64(&p.executedInstructions, 1)
	}

	// If any work was done, count as a busy cycle
	if workDone {
		atomic.AddInt64(&p.busyCycles, 1)
	}
}

// fetchNextInstruction creates a synthetic instruction for simulation
func (p *Processor) fetchNextInstruction() *Instruction {
	// This is a simplified synthetic instruction generator
	// In a real simulator, this would fetch from memory

	// Create a simple ALU instruction
	inst := &Instruction{
		Address:    p.pc,
		Opcode:     0x01,             // ADD
		Operands:   []uint8{1, 2, 3}, // r1 = r2 + r3
		Type:       "Integer",
		Stage:      "Fetch",
		CyclesLeft: 1,
	}

	// Increment PC
	p.pc += 4 // Assuming 4-byte instructions

	return inst
}

// GetExecutedInstructions returns the number of instructions executed by this core
func (p *Processor) GetExecutedInstructions() int64 {
	return atomic.LoadInt64(&p.executedInstructions)
}

// GetUtilization returns the core utilization (busy cycles / total cycles)
func (p *Processor) GetUtilization() float64 {
	cycles := atomic.LoadInt64(&p.cycleCount)
	if cycles == 0 {
		return 0.0
	}

	busyCycles := atomic.LoadInt64(&p.busyCycles)
	return float64(busyCycles) / float64(cycles)
}

func (p *Processor) GetID() int {
	return p.ID
}

// GetPipelineState returns a copy of the current pipeline state
func (p *Processor) GetPipelineState() []*pipeline.Stage {
	return p.pipeline.GetStages()
}

func (p *Processor) Reset() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.pc = 0
	p.instructionQueue = make([]Instruction, 0, 32)
	atomic.StoreInt64(&p.executedInstructions, 0)
	atomic.StoreInt64(&p.cycleCount, 0)
	atomic.StoreInt64(&p.busyCycles, 0)

	p.pipeline.Flush()

	for i := range p.registersInt {
		p.registersInt[i] = 0
	}

	for i := range p.registersFloat {
		p.registersFloat[i] = 0.0
	}

	for _, units := range p.executionUnits {
		for _, unit := range units {
			unit.Busy = false
		}
	}
}
