package server

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ChainTable represents a table of chain data
type ChainTable struct {
	Nodes []Node `yaml:"nodes"`
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

type Role int

const (
	// represent the head node
	HEAD Role = iota
	// represent any of the middle node
	MIDDLE
	// represent the tail node
	TAIL

	Unknown
)

// get the role of the node given the node id
func (ct *ChainTable) Role(id int) Role {
	size := len(ct.Nodes)
	fmt.Println(ct.Nodes)
	if id == 0 {
		return HEAD
	} else if id < size-1 {
		return MIDDLE
	} else if id == size-1 {
		return TAIL
	}
	return Unknown
}

// String returns a YAML string representation of the ChainTable
func (ct *ChainTable) String() string {
	data, err := yaml.Marshal(ct)
	if err != nil {
		return fmt.Sprintf("Error marshaling to YAML: %v", err)
	}
	return string(data)
}
