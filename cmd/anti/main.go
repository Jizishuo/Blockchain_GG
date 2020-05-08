package main

import (
	"Blockchain_GG/crypto"
	"Blockchain_GG/utils"
	"flag"
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

	// p2p 供应

}
