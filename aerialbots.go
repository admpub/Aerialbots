package Aerialbots

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/kr/pty"
)

const (
	PIOUT      = "/tmp/stdout"
	SWITCHPATH = "switch_path"
)

// Ab 用来判断Command执行后的stdout是否包含Input中所定义的字符串
// 如果存在相对应的字符串，那么就相对应的将Output通过Stdin传输给Command
// 的Stdin,以此达到定向交互的目的
type Ab struct {
	Input  map[int]string   `json:"input"`  // Input 预定义字符串，Key为出现顺序,value为过滤字符
	Ouput  map[int]string   `json:"output"` // Output 欲交互的字符串,key为出现顺序与Input的key相对应, value为交互字符串
	Assist map[int][]string `json:"assist"` // Assist 辅助命令集，在输入Output之前会匹配是否有合适的Assist，如果有，则会首先执行Assist的命令
	Cmd    *exec.Cmd        // Cmd 封装好的Cmd指针
}

// Start 执行过滤，Ab会判断Stdout是否包含需要过滤的字符
func (a *Ab) Start() error {

	FOUT, err := os.Create(PIOUT)
	if err != nil {
		return err
	}

	defer func() {
		FOUT.Close()
	}()

	f, err := pty.Start(a.Cmd)
	if err != nil {
		return err
	}
	var out []byte
	go func() {
		probe := 0
		idr := 0
		guard := 0 //guard哨兵用于判断当前命令是否发生变化

		for {
			out, _ = ioutil.ReadFile(FOUT.Name())
			l := len(strings.TrimSpace(string(out)))

			if guard < l && l > 0 {
				guard = l
				str := string(out)[idr:]

				if strings.Contains(str, a.Input[probe]) {
					// execute assist frist
					err := assist(probe, a)
					if err != nil {
						fmt.Println(err.Error())
					}
					idr += strings.Index(str, a.Input[probe])
					in := []byte(a.Ouput[probe] + "\n")
					f.Write(in)
					guard += len(in)
					probe++
				}
			}
			if probe == len(a.Input) {
				f.Write([]byte{4})
				return
			}
		}
	}()
	io.Copy(FOUT, f)

	err = a.Cmd.Wait()
	if err != nil {
		return errors.New(err.Error() + "\n log:\n" + string(out))
	}

	return nil
}

// assist 判断指定Probe是否存在辅助命令，如果有则首先执行辅助命令
func assist(probe int, a *Ab) error {
	if len(a.Assist[probe]) > 0 {
		for _, a := range a.Assist[probe] {
			as := strings.Split(a, "=")
			if len(as) < 2 {
				return errors.New("Wrong Assis format")
			}
			switch as[0] {
			case SWITCHPATH:
				return os.Chdir(as[1])
			}
		}
	}

	return nil
}
