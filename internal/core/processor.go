package core

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/jasonKoogler/cpu-sim/internal/config"
)

type ExecutionUnit struct {
	Type     string // "ALU", "FPU", "LoadStore", "Branch"
	Busy     bool   // true if the unit is currently executing an instruction
	Pipeline int    // number of stages in this unit
}

type Processor struct {
	ID                   int
	config               *config.Config
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
	Opcode     string
	Operands   []uint64
	Type       string // "Integer", "Float", "Memory", "Branch", "System"
	Stage      string // Current pipeline stage
	CyclesLeft int    // Number of cycles left in the current stage
}

func NewProcessor(id int, cfg *config.Config) (*Processor, error) {
	if cfg == nil {
		return nil, fmt.Errorf("nil configuation provided")
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
		return nil, fmt.Errorf("unsupported ISA: %s", cfg.ISA)
	}

	proc := &Processor{
		ID:               id,
		config:           cfg,
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

	// Temp: Sim synthetic workload
	// TODO: temeporary, replace with actual instruction handling
	if p.cycleCount%10 == 0 { // every 10 cycles, execute an instruction
		atomic.AddInt64(&p.executedInstructions, 1)
		workDone = true
	}

	if workDone {
		atomic.AddInt64(&p.busyCycles, 1)
	}
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

func (p *Processor) Reset() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.pc = 0
	p.instructionQueue = make([]Instruction, 0, 32)
	atomic.StoreInt64(&p.executedInstructions, 0)
	atomic.StoreInt64(&p.cycleCount, 0)
	atomic.StoreInt64(&p.busyCycles, 0)

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
