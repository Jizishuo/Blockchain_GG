package db

import (
	"Blockchain_GG/utils"
	"Blockchain_GG/serialize/cp"
)

var (
	logger = utils.NewLogger("db")
	instance db  //实例
)

type db interface {
	Init(path string) error
	HasGenesis() bool  //成因
	PutGenesis(block *cp.Block)
}


func Init(path string) error {
	instance = newBadger()
	return instance.Init()
}