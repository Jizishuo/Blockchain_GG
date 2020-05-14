package main

import (
	"Blockchain_GG/crypto"
	"Blockchain_GG/p2p/peer"
	"Blockchain_GG/utils"
	"Blockchain_GG/db"
	"flag"
	"Blockchain_GG/p2p"
	"log"
	"github.com/btcsuite/btcd/btcec"
)

func main() {
	cf := flag.String("c", "", "congig file")
	pprofPort := flag.Int("pprof", 0, "pprof prot, used by developers")
	flag.Parse()

	conf, err := parserConfig(*cf)
	if err != nil {
		log.Fatal(err)
	}
	utils.SetLogLevel(conf.LogLevel)
	logger := utils.GetStdoutLog()

	// 载入 key
	var privKey *btcec.PrivateKey
	if conf.Key.Type == crypto.PlainKeyType {
		privKey, err = crypto.RestorePKey(conf.Key.Path)
		if err != nil {
			logger.Fatal("restorn sKey failed: %v\n", err)
		}
	}
	if conf.Key.Type == crypto.SealKeyType {
		privKey, err = crypto.RestorePKey(conf.Key.Path)
		if err != nil {
			logger.Fatal("resotore pkey failed: %v\n", err)
		}
	}
	pubKey := privKey.PubKey()

	// p2p 供应 peer provider
	provider := peer.NewProvider(conf.IP, conf.Port, pubKey)
	seeds := parseSeeds(conf.Seeds)
	provider.AddSeeds(seeds)
	provider.Start()

	// p2p node
	nodeConfig := &p2p.Config{
		NodeIP: conf.IP,
		NodePort: conf.Port,
		Provider: provider,
		MaxPeerNum: conf.MaxPeers,
		PrivKey: privKey,
		Type: conf.NodeType,
		ChainID: conf.ChainID,
	}
	node := p2p.NewNode(nodeConfig)
	node.Start()

	// db
	if err = db.Init(conf.DataPath); err!=nil {
		logger.Fatal("init db failed: %v\n", err)
	}
	logger.Info("database initialize successfully under the data path: %s\n", conf.DataPath)



}
