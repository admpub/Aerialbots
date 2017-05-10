package Aerialbots

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/kr/pty"
)

const (
	// PIOUT 执行日志保存位置
	PIOUT = "/tmp/stdout"
	// SWITCHPATH 切换工作目录
	SWITCHPATH = "SWITCH_PATH"
	// SCENEDESIGN 定制场景
	SCENEDESIGN = "SCENE_DESIGN"
)

// Ab 用来判断Command执行后的stdout是否包含Input中所定义的字符串
// 如果存在相对应的字符串，那么就相对应的将Output通过Stdin传输给Command
// 的Stdin,以此达到定向交互的目的
type Ab struct {
	Input  map[int]string   `json:"input"`  // Input 预定义字符串，Key为出现顺序,value为过滤字符
	Ouput  map[int]string   `json:"output"` // Output 欲交互的字符串,key为出现顺序与Input的key相对应, value为交互字符串
	Assist map[int][]string `json:"assist"` // Assist 辅助命令集，在输入Output之前会匹配是否有合适的Assist，如果有，则会首先执行Assist的命令
	Cmd    *exec.Cmd        // Cmd 封装好的Cmd指针
	debug  bool
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

	// disable colour in pty
	go func() {
		probe := 0
		// idr := 0
		guard := 0 //guard哨兵用于判断当前命令是否发生变化

		for {
			out, _ = ioutil.ReadFile(FOUT.Name())
			l := len(strings.TrimSpace(string(out)))

			if guard < l && l > 0 {
				str := string(out)[guard:l]
				guard = l
				if canExecute(a, probe, f, str) {
					// if strings.Contains(str, a.Input[probe]) {
					// execute assist frist
					err := assist(a, probe, f)
					if err != nil {
						fmt.Println(err.Error())
					}
					in := []byte(a.Ouput[probe] + "\n")
					f.Write(in)
					ensureCMD(a.Ouput[probe])
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
func assist(a *Ab, probe int, pty *os.File) error {

	if len(a.Assist[probe]) > 0 {
		for _, a := range a.Assist[probe] {
			as := strings.Split(a, "=")
			if len(as) < 2 {
				return errors.New("Wrong Assis format")
			}
			switch strings.ToUpper(as[0]) {
			case SWITCHPATH:
				in := []byte("cd " + as[1] + "\n")
				_, err := pty.Write(in)
				pty.Write([]byte{5})
				ensureSWITCH(as[1])
				return err
			}
		}
	}

	return nil
}

// canExecute 判断当前是否满足执行条件, 各条件间为或关系
// 条件1: 当前pty中包括Input指定锚点
// 条件2: 当前pty中满足Assist的条件判断
func canExecute(a *Ab, probe int, pty *os.File, str string) bool {
	str = strings.TrimSpace(str)
	if a.debug {
		fmt.Printf("compare [%s] [%s] [%v] [%d] [%d] [%v]\n", str, a.Input[probe], strings.Contains(str, a.Input[probe]), len(str), len(a.Input[probe]), []byte(str))
	}

	if len(str) == 0 {
		return false
	}
	if len(a.Assist[probe]) > 0 {
		for _, ap := range a.Assist[probe] {
			if strings.Contains(strings.ToUpper(ap), SCENEDESIGN) {
				as := strings.Split(ap, "=")
				if len(as) != 3 {
					return false
				}
				if strings.Contains(str, as[1]) {
					cmd := ""
					if strings.HasPrefix(as[2], "SEP") {
						idx, err := strconv.Atoi(as[2][3:])
						if err != nil {
							fmt.Printf("%s format error\n", as[2])
							return false
						}
						cmd = a.Ouput[idx]
					} else {
						cmd = as[2]
					}
					in := []byte(cmd + "\n")
					pty.Write(in)
					return true
				}
			}
		}
	}

	if strings.Contains(str, a.Input[probe]) {
		return true
	}

	return false
}

func ensureCMD(str string) {
	fmt.Printf("开始执行:[%s]\n", str)
}

func ensureSWITCH(str string) {
	fmt.Printf("开始切换路径:[%s]\n", str)
}
