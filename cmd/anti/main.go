package main

import (
	"Blockchain_GG/core"
	"Blockchain_GG/core/blockchain"
	"Blockchain_GG/crypto"
	"Blockchain_GG/db"
	"Blockchain_GG/p2p"
	"Blockchain_GG/p2p/peer"
	"Blockchain_GG/rpc"
	"Blockchain_GG/utils"
	"flag"
	"fmt"
	"github.com/btcsuite/btcd/btcec"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
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

	// conf.Key.Path 文件路径 载入 私钥 key
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
	pubKey := privKey.PubKey() // 公钥

	// p2p 供应 peer provider  创建和管理比特币对等网络，及上行下行数据库的处理
	provider := peer.NewProvider(conf.IP, conf.Port, pubKey)
	seeds := parseSeeds(conf.Seeds)
	provider.AddSeeds(seeds)
	provider.Start()

	// p2p node
	nodeConfig := &p2p.Config{
		NodeIP:     conf.IP,
		NodePort:   conf.Port,
		Provider:   provider,
		MaxPeerNum: conf.MaxPeers,
		PrivKey:    privKey,
		Type:       conf.NodeType,
		ChainID:    conf.ChainID,
	}
	node := p2p.NewNode(nodeConfig)
	node.Start()

	// db
	if err = db.Init(conf.DataPath); err != nil {
		logger.Fatal("init db failed: %v\n", err)
	}
	logger.Info("database initialize successfully under the data path: %s\n", conf.DataPath)

	// core 模块
	// 返回字符串表示的整数值，用于无符号整型
	blockDiffLimit, err := strconv.ParseUint(conf.BlockDifficultyLimit, 16, 32)
	if err != nil {
		logger.Fatalln(err)
	}
	evidenceDiffLimit, err := strconv.ParseUint(conf.EvidenceDifficultyLimit, 16, 32)
	if err != nil {
		logger.Fatalln(err)
	}
	coreInstance := core.NewCore(&core.Config{
		Node:         node,
		NodeType:     conf.NodeType,
		PrivKey:      privKey,
		ParallelMine: conf.ParalleMine,
		Config: &blockchain.Config{
			BlockTargetLimit:    uint32(blockDiffLimit),
			EvidenceTargetLimit: uint32(evidenceDiffLimit),
			BlockInterval:       conf.BlockInterval,
			Genesis:             conf.Genesis,
		},
	})

	// local http server
	httpConfig := &rpc.Config{
		Port: conf.HTTPPort,
		C:    coreInstance,
	}
	httpServer := rpc.NewServer(httpConfig)
	httpServer.Start()

	// pprof
	if *pprofPort != 0 {
		go func() {
			pprofAddress := fmt.Sprintf("localhost:%d", pprofPort)
			log.Println(http.ListenAndServe(pprofAddress, nil))
		}()
	}
	// 正常等待关机
	sc := make(chan os.Signal)
	// Notify函数让signal包将输入信号转发到c。
	//如果没有列出要传递的信号，会将所有输入信号传递到c；否则只传递列出的输入信号。
	signal.Notify(sc, os.Interrupt) //Interrupt（中断信号）和Kill（强制退出信号）
	signal.Notify(sc, syscall.SIGTERM)
	select {
	case <-sc:
		logger.Infoln("Quiting....")
		httpServer.Stop()
		coreInstance.Stop()
		node.Stop()
		db.Close()
		logger.Infoln("Bye!...")
		return
	}
}
