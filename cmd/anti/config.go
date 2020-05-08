package main

import (
	"Blockchain_GG/params"
	"Blockchain_GG/utils"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
)

type config struct {
	NodeType                params.NodeType `json:"node_type"` // uint8
	IP                      string          `json:"ip"`
	Port                    int             `json:"port"`
	Seeds                   []string        `json:"seeds"`
	MaxPeers                int             `json:"max_peers"`
	LogLevel                int             `json:"log_level"`
	DataPath                string          `json:"data_path"`
	Key                     KeyConfig       `json:"key"`
	ChainID                 uint8           `json:"chain_id"`
	BlockDifficultyLimit    string          `json:"block_diff_limit"`
	EvidenceDifficultyLimit string          `json:"evidence_diff_limit"`
	BlockInterval           int             `json:"block_interval"`
	ParalleMine             int             `json:"parallel_mine"`
	Genesis                 string          `json:"genesis"`
	HTTPPort                int             `json:"http_port"`
}

type KeyConfig struct {
	Type int    `json:"type"`
	Path string `json:"path"`
}

func parserConfig(cf string) (*config, error) {
	if len(cf) == 0 {
		return nil, fmt.Errorf("miss config file")
	}

	if err := utils.AccessCheck(cf); err != nil {
		return nil, err
	}
	jsonContent, err := ioutil.ReadFile(cf)
	if err != nil {
		return nil, fmt.Errorf("read config file failed:%v", err)
	}
	conf := &config{}
	if err := json.Unmarshal(jsonContent, &conf); err != nil {
		return nil, fmt.Errorf("config parse failed:%v", err)
	}
	//if err := ver

}

func verifyConfig(c *config) error {
	if c.NodeType != params.FullNode &&c.NodeType != params.LifhtNode {
		return fmt.Errorf("invalid node type:%d", c.NodeType)
	}

	if c.NodeType == params.LifhtNode {
		return fmt.Errorf("not support light node now")
	}
	// 解析为IP地址，并返回该地址 To4将一个IPv4地址转换为4字节表示。如果ip不是IPv4地址，To4会返回nil。
	if ip := net.ParseIP(c.IP);ip == nil || ip.To4() == nil {
		return fmt.Errorf("invalid IPv4:%d", c.IP)
	}
	if c.Port <=0 || c.Port > 65535 {
		return fmt.Errorf("invalid port :%d", c.Port)
	}
	if c. MaxPeers <= 0 {
		return fmt.Errorf("invalid max perr number: %d", c.MaxPeers)
	}

	if c.LogLevel < utils.
}