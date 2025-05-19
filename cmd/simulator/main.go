package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jasonKoogler/cpu-sim/internal/config"
	"github.com/jasonKoogler/cpu-sim/internal/pipeline"
	"github.com/jasonKoogler/cpu-sim/internal/simulator"
)

func main() {
	configPath := flag.String("config", "configs/default.yaml", "Path to the configuration file")
	verbose := flag.Bool("v", false, "Enable verbose output")
	numCycles := flag.Int64("cycles", 1000, "Number of cycles to simulate")
	showPipeline := flag.Bool("show-pipeline", false, "Show the pipeline structure")
	flag.Parse()

	logger := log.New(os.Stdout, "", log.LstdFlags)

	if *verbose {
		logger.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
	}

	if *numCycles <= 0 {
		logger.Fatalf("Invalid cycle count: %d", *numCycles)
	}

	logger.Println("Multicore Processor Simulator")

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		logger.Fatalf("Failed to load configuration: %v", err)
	}

	fmt.Println("\nConfiguration Summary:")
	fmt.Printf("	Cores: %d @ %d MHz\n", cfg.NumCores, cfg.ClockFrequency)
	fmt.Printf("	ISA: %s\n", cfg.ISA)
	fmt.Printf("	Pipeline Depth: %d stages\n", cfg.PipelineDepth)
	fmt.Printf("	Cache Coherence: %s\n", cfg.CoherenceProtocol)
	fmt.Printf("	Interconnect: %s, %d GB/s\n", cfg.InterconnectType, cfg.InterconnectBandwidth)
	fmt.Printf("	Memory Latency: %d cycles\n", cfg.MemoryLatency)
	fmt.Printf("	Workload: %s\n", cfg.WorkloadPath)

	fmt.Println("\nMemory Hierarchy:")
	fmt.Printf("	L1 Cache: %d KB, %d-way, %d cycles\n", cfg.L1Size, cfg.L1Associativity, cfg.L1Latency)
	fmt.Printf("	L2 Cache: %d KB, %d-way, %d cycles\n", cfg.L2Size, cfg.L2Associativity, cfg.L2Latency)
	fmt.Printf("	L3 Cache: %d KB, %d-way, %d cycles\n", cfg.L3Size, cfg.L3Associativity, cfg.L3Latency)

	// Show pipeline structure if requested
	if *showPipeline {
		pipe, err := pipeline.NewPipeline(cfg.PipelineDepth, cfg.ISA)
		if err != nil {
			logger.Fatalf("Failed to create pipeline: %v", err)
		}

		fmt.Println("\nPipeline Structure:")
		fmt.Printf("  Total Stages: %d\n", len(pipe.GetStages()))

		fmt.Print("  Pipeline Flow: ")
		stages := pipe.GetStages()
		for i, stage := range stages {
			fmt.Printf("%s", stage.Name)
			if i < len(stages)-1 {
				fmt.Print(" → ")
			}
		}
		fmt.Println()
	}

	sim, err := simulator.New(cfg)
	if err != nil {
		logger.Fatalf("Failed to initialize simulator: %v", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Printf("Starting simulation for %d cycles...", *numCycles)

		if err := sim.Run(*numCycles); err != nil {
			logger.Fatalf("Simulation failed: %v", err)
		}

		stats := sim.GetStatistics()
		fmt.Println("\nSimulation Statistics:")
		fmt.Printf("	Total Cycles: %d\n", stats.TotalCycles)
		fmt.Printf("	Instructions Executed: %d\n", stats.InstructionsExecuted)
		fmt.Printf("	IPC: %.2f\n", stats.IPC)
		fmt.Printf("	Cache Hit Rate: %.2f%%\n", stats.CacheHitRate*100)
		fmt.Printf("	Core Utilization: %.2f%%\n", stats.CoreUtilization[0]*100)
		fmt.Printf("	Memory Access Latency: %.2f cycles\n", stats.MemoryAccessLatency)
		fmt.Printf("	Interconnect Utilization: %.2f%%\n", stats.InterconnectUtilization*100)

		fmt.Println("\nCore Utilization:")
		for i, util := range stats.CoreUtilization {
			fmt.Printf("	Core %d: %.2f%%\n", i, util*100)
		}

		os.Exit(0)
	}()

	<-sigChan
	logger.Println("Received termination signal. Shutting down...")
	sim.Shutdown()
	logger.Println("Simulation terminated successfully")
}
