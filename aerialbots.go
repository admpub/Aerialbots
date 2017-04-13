package Aerialbots

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"time"
)

const (
	PIPE = "/tmp/stdout"
)

// Ab 用来判断Command执行后的stdout是否包含Input中所定义的字符串
// 如果存在相对应的字符串，那么就相对应的将Output通过Stdin传输给Command
// 的Stdin,以此达到定向交互的目的
type Ab struct {
	Input map[int]string `json:"input"`  // Input 预定义字符串，Key为出现顺序,value为过滤字符
	Ouput map[int]string `json:"output"` // Output 欲交互的字符串,key为出现顺序与Input的key相对应, value为交互字符串
	Cmd   *exec.Cmd      // Cmd 封装好的Cmd指针
}

// Start 执行过滤，Ab会判断Stdout是否包含需要过滤的字符
func (a *Ab) Start() error {
	stop := make(chan int)
	out, err := os.Create(PIPE)
	if err != nil {
		return err
	}

	a.Cmd.Stdout = out
	a.Cmd.Stdin = os.Stdin
	if err := a.Cmd.Start(); err != nil {
		return err
	}

	go func() {
		c := time.Tick(500 * time.Millisecond)
		for {
			select {
			case <-c:
				info, _ := ioutil.ReadFile(PIPE)
				log.Printf("OUT:[%s]/n", string(info))
			case <-stop:
				return
			}
		}
	}()

	a.Cmd.Wait()
	stop <- 1
	return nil
}
