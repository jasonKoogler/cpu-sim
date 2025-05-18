package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	content := `
numCores: 8
clockFrequency: 4000
isa: "x86"
pipelineDepth: 14
l1Size: 64
l1Associativity: 8
l1Latency: 2
l2Size: 1024
l2Associativity: 16
l2Latency: 10
l3Size: 16384
l3Associativity: 16
l3Latency: 35
memoryLatency: 150
coherenceProtocol: "MOESI"
interconnectType: "mesh"
interconnectBandwidth: 512
workloadPath: "workloads/test.bin"
`
	tmpfile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	// Load config
	cfg, err := LoadConfig(tmpfile.Name())
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	// Verify values
	if cfg.NumCores != 8 {
		t.Errorf("Expected NumCores = 8, got %d", cfg.NumCores)
	}
	if cfg.ClockFrequency != 4000 {
		t.Errorf("Expected ClockFrequency = 4000, got %d", cfg.ClockFrequency)
	}
	if cfg.ISA != "x86" {
		t.Errorf("Expected ISA = x86, got %s", cfg.ISA)
	}
	if cfg.CoherenceProtocol != "MOESI" {
		t.Errorf("Expected CoherenceProtocol = MOESI, got %s", cfg.CoherenceProtocol)
	}
	if cfg.InterconnectType != "mesh" {
		t.Errorf("Expected InterconnectType = mesh, got %s", cfg.InterconnectType)
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name: "Valid config",
			cfg: Config{
				NumCores:          4,
				ClockFrequency:    3000,
				ISA:               "RISC-V",
				PipelineDepth:     5,
				CoherenceProtocol: "MESI",
				InterconnectType:  "ring",
			},
			wantErr: false,
		},
		{
			name: "Invalid cores",
			cfg: Config{
				NumCores:          0,
				ClockFrequency:    3000,
				ISA:               "RISC-V",
				PipelineDepth:     5,
				CoherenceProtocol: "MESI",
				InterconnectType:  "ring",
			},
			wantErr: true,
		},
		{
			name: "Invalid ISA",
			cfg: Config{
				NumCores:          4,
				ClockFrequency:    3000,
				ISA:               "Invalid",
				PipelineDepth:     5,
				CoherenceProtocol: "MESI",
				InterconnectType:  "ring",
			},
			wantErr: true,
		},
		{
			name: "Invalid protocol",
			cfg: Config{
				NumCores:          4,
				ClockFrequency:    3000,
				ISA:               "RISC-V",
				PipelineDepth:     5,
				CoherenceProtocol: "Invalid",
				InterconnectType:  "ring",
			},
			wantErr: true,
		},
		{
			name: "Invalid interconnect",
			cfg: Config{
				NumCores:          4,
				ClockFrequency:    3000,
				ISA:               "RISC-V",
				PipelineDepth:     5,
				CoherenceProtocol: "MESI",
				InterconnectType:  "Invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateConfig(&tt.cfg); (err != nil) != tt.wantErr {
				t.Errorf("validateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg == nil {
		t.Fatalf("DefaultConfig() returned nil")
	}

	if cfg.NumCores != 4 {
		t.Errorf("Expected default NumCores = 4, got %d", cfg.NumCores)
	}

	if cfg.ISA != "RISC-V" {
		t.Errorf("Expected default ISA = RISC-V, got %s", cfg.ISA)
	}

	if cfg.PipelineDepth != 5 {
		t.Errorf("Expected default PipelineDepth = 5, got %d", cfg.PipelineDepth)
	}

	if cfg.CoherenceProtocol != "MESI" {
		t.Errorf("Expected default CoherenceProtocol = MESI, got %s", cfg.CoherenceProtocol)
	}
}
