package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the simulator configuration
type Config struct {
	// Core configuration
	NumCores       int    `yaml:"numCores"`
	ClockFrequency int    `yaml:"clockFrequency"` // MHz
	ISA            string `yaml:"isa"`            // Instruction Set Architecture
	PipelineDepth  int    `yaml:"pipelineDepth"`

	// Memory hierarchy
	L1Size          int `yaml:"l1Size"` // KB
	L1Associativity int `yaml:"l1Associativity"`
	L1Latency       int `yaml:"l1Latency"` // cycles

	L2Size          int `yaml:"l2Size"` // KB
	L2Associativity int `yaml:"l2Associativity"`
	L2Latency       int `yaml:"l2Latency"` // cycles

	L3Size          int `yaml:"l3Size"` // KB
	L3Associativity int `yaml:"l3Associativity"`
	L3Latency       int `yaml:"l3Latency"` // cycles

	MemoryLatency int `yaml:"memoryLatency"` // cycles

	// Cache coherence protocol
	CoherenceProtocol string `yaml:"coherenceProtocol"` // MESI, MOESI, etc.

	// Interconnect
	InterconnectType      string `yaml:"interconnectType"`      // bus, ring, mesh, etc.
	InterconnectBandwidth int    `yaml:"interconnectBandwidth"` // GB/s

	// Workload
	WorkloadPath string `yaml:"workloadPath"`
}

// LoadConfig loads configuration from a YAML file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if err := validateConfig(&cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// validateConfig checks if the configuration is valid
func validateConfig(cfg *Config) error {
	if cfg.NumCores <= 0 {
		return fmt.Errorf("number of cores must be positive")
	}

	if cfg.ClockFrequency <= 0 {
		return fmt.Errorf("clock frequency must be positive")
	}

	if cfg.PipelineDepth <= 0 {
		return fmt.Errorf("pipeline depth must be positive")
	}

	// Validate ISA
	validISAs := map[string]bool{"RISC-V": true, "x86": true, "ARM": true, "MIPS": true, "Custom": true}
	if !validISAs[cfg.ISA] {
		return fmt.Errorf("unsupported ISA: %s", cfg.ISA)
	}

	// Validate coherence protocol
	validProtocols := map[string]bool{"MESI": true, "MOESI": true, "MSI": true, "MESIF": true, "None": true}
	if !validProtocols[cfg.CoherenceProtocol] {
		return fmt.Errorf("unsupported coherence protocol: %s", cfg.CoherenceProtocol)
	}

	// Validate interconnect type
	validInterconnects := map[string]bool{"bus": true, "ring": true, "mesh": true, "crossbar": true, "torus": true}
	if !validInterconnects[cfg.InterconnectType] {
		return fmt.Errorf("unsupported interconnect type: %s", cfg.InterconnectType)
	}

	return nil
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		NumCores:       4,
		ClockFrequency: 3000, // 3 GHz
		ISA:            "RISC-V",
		PipelineDepth:  5, // 5-stage pipeline

		L1Size:          32, // 32 KB
		L1Associativity: 8,
		L1Latency:       3, // 3 cycles

		L2Size:          256, // 256 KB
		L2Associativity: 8,
		L2Latency:       12, // 12 cycles

		L3Size:          8192, // 8 MB
		L3Associativity: 16,
		L3Latency:       40, // 40 cycles

		MemoryLatency: 200, // 200 cycles

		CoherenceProtocol: "MESI",

		InterconnectType:      "ring",
		InterconnectBandwidth: 256, // 256 GB/s

		WorkloadPath: "workloads/default.bin",
	}
}
