package utils

import "sync"

/*
	LoopMode 是一种工作模式，指示所有者在一个或多个长期运行的 go 例程中分离其工作逻辑。
	所有者应在其设置功能中调用 Start工作（），并在其清理函数中调用 Stop（），
	并且其每个长期运行的例程都应该像：
	```go
	func loop() {
		lm.Add()
		defer lm.Done()
		for {
			select {
			case <-lm.D:
				return
			// case ...:
			// do the jobs
			}
		}
	}
	```
*/

type LoopMode struct {
	working bool
	routinesNum int
	waitGroup sync.WaitGroup
	D chan bool
}

// NewLoop 返回 LoopMode.Param 例程是长期运行运行操作例程的数量（必须 |0）
func NewLoop(routines int) *LoopMode {
	if routines <= 0 {
		return nil
	}
	return &LoopMode{
		working: false,
		routinesNum: routines,
		D: make(chan bool, routines),
	}
}
func (l *LoopMode) StartWorking() {
	l.working = true
}

func (l *LoopMode) Stop() bool {
	if !l.working {
		return false
	}
	l.working = false
	for i := 0; i < l.routinesNum; i++ {
		l.D <- true
	}
	l.waitGroup.Wait()
	return true
}

func (l *LoopMode) Add() {
	l.waitGroup.Add(1)
}

func (l *LoopMode) Done() {
	l.waitGroup.Done()
}

func (l *LoopMode) IdWorking() bool {
	return l.working
}