package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/jasonKoogler/cpu-sim/internal/config"
)

func main() {
	configPath := flag.String("config", "configs/default.yaml", "Path to the configuration file")
	verbose := flag.Bool("v", false, "Enable verbose output")
	flag.Parse()

	logger := log.New(os.Stdout, "", log.LstdFlags)

	if *verbose {
		logger.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
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
}
