package main

import (
	"Blockchain_GG/crypto"
	"Blockchain_GG/p2p/peer"
	"Blockchain_GG/params"
	"Blockchain_GG/utils"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"runtime"
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
	if err := verifyConfig(conf); err != nil {
		return nil, err
	}
	return conf, nil

}

func verifyConfig(c *config) error {
	if c.NodeType != params.FullNode && c.NodeType != params.LightNode {
		return fmt.Errorf("invalid node type:%d", c.NodeType)
	}

	if c.NodeType == params.LightNode {
		return fmt.Errorf("not support light node now")
	}
	// 解析为IP地址，并返回该地址 To4将一个IPv4地址转换为4字节表示。如果ip不是IPv4地址，To4会返回nil。
	if ip := net.ParseIP(c.IP); ip == nil || ip.To4() == nil {
		return fmt.Errorf("invalid IPv4:%d", c.IP)
	}
	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("invalid port :%d", c.Port)
	}
	if c.MaxPeers <= 0 {
		return fmt.Errorf("invalid max perr number: %d", c.MaxPeers)
	}

	if c.LogLevel < utils.LogErrorLevel || c.LogLevel > utils.LogDebugLevel {
		return fmt.Errorf("invalid log level: %d", c.LogLevel)
	}
	if err := utils.AccessCheck(c.DataPath); err != nil {
		return err
	}

	fmt.Printf("data path: %d\n", c.DataPath)

	if c.Key.Type != crypto.SealKeyType && c.Key.Type != crypto.PlainKeyType {
		return fmt.Errorf("invalid key type")
	}
	if err := utils.AccessCheck(c.Key.Path); err != nil {
		return err
	}

	if len(c.BlockDifficultyLimit) != 8 || len(c.EvidenceDifficultyLimit) != 8 {
		return fmt.Errorf("invalid difficulty limit")
	}

	if c.BlockInterval <= 0 {
		return fmt.Errorf("invalid block interval")
	}
	if c.ParalleMine < 0 || c.ParalleMine > runtime.NumCPU() {
		return fmt.Errorf("invalid parallel num")
	}

	if len(c.Genesis) == 0 {
		return fmt.Errorf("invalid genesis")
	}

	if c.HTTPPort <= 0 || c.HTTPPort > 65535 || c.HTTPPort == c.Port {
		return fmt.Errorf("invalid http port: %d", c.HTTPPort)
	}

	return nil
}

func parseSeeds(seeds []string) []*peer.Peer {
	var result []*peer.Peer
	for _, seed := range seeds {
		ip, port := utils.ParseUPPort(seed)
		if ip == nil {
			continue
		}
		p := peer.NewPeer(ip, port, nil)
		result = append(result, p)
	}
	return result
}
