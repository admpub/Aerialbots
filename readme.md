# Aerialbots

Aerialbots 是一个可以在交互式Shell中输入自定义数据的golang package. 

## 应用场景

Aerialbots 主要用在exec.Command场景中。 传统的exec.Command在执行command时，受制于pty无法与Child Process进行交互时输入。 通过Aerialbots，用户可以在exec.Command过程中创建command运行时的pty，通过匹配用户预定义的字符串，在pty中输入相对应的字符，借此达到自动化的目的。

## Demo

下面实例假设,用户在代码中需要通过ssh登录到另外一个节点中，执行一些预定操作。虽然golang也有SSH package可以完成这个User Case。 但这里仅仅是为了演示如何通过Aerialbots完成这个Case。
```
	cmd := exec.Command("ssh", "xxxx@10.10.1.1")
	ab := new(Aerialbots.Ab)
	ab.Cmd = cmd

	input := make(map[int]string)
	output := make(map[int]string)

	//假设执行ssh xxxx@10.10.1.1后，返回如下字符:
	//admin@10.10.1.1's password:
	//所以我们设定当返回的字符中包含password后，输入admin
	input[0] = "password" 
	output[0] = "admin"

	//当执行ssh登录成功后,对方节点返回的信息如下：
	//Debian GNU/Linux comes with ABSOLUTELY NO WARRANTY, to the extent
	//permitted by applicable law.
	//Last login: Fri Apr 14 11:49:45 2017 from 10.11.1.113
	//admin@server:~$
	//因此我们设定当返回的字符中包含server时，留一个"到此一游"的标记
	input[1] = "server"
	output[1] = "echo \"have visited this place\" >/tmp/1.log"

    
	ab.Input = input
	ab.Ouput = output

	err := ab.Start()
	if err != nil {
		fmt.Println("ERROR", err.Error())
	}
```
