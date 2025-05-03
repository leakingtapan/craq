package server

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ChainTable represents a table of chain data
type ChainTable struct {
	Nodes []Node `yaml:"chains"`
}

// Node represents a single chain entry
type Node struct {
	ID   string `yaml:"id"`
	Addr string `yaml:"addr"`
}

// ParseChainTable parses a YAML chain-table file into a ChainTable struct
func ParseChainTable(filePath string) (*ChainTable, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read chain table file: %w", err)
	}

	table := &ChainTable{}
	if err := yaml.Unmarshal(data, table); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return table, nil
}

// String returns a YAML string representation of the ChainTable
func (ct *ChainTable) String() string {
	data, err := yaml.Marshal(ct)
	if err != nil {
		return fmt.Sprintf("Error marshaling to YAML: %v", err)
	}
	return string(data)
}
