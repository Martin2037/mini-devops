package main

import (
	"devops/config"
	"github.com/bramvdbogaerde/go-scp"
	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"golang.org/x/crypto/ssh"
	"sync"
	"time"
)

type devops struct {
	yaml config.Setting
	spin *spinner.Spinner
}

func (do devops) getResult() string {
	color.Blue("Author：朱珺阳\nDescription：devops小工具\n\n")
	color.Cyan("以下为当前已有项目列表，选择项目\n")
	prompt := promptui.Select{
		Label: "选择项目",
		Items: config.GetProjects(),
	}
	_, result, err := prompt.Run()
	config.ThrowError("prompt运行失败", err)
	return result
}

func (do devops) initSettings() {
	do.spin.Color("magenta", "bold")
	do.spin.FinalMSG = "\n完成✅\t\n"
	if len(do.yaml.Server) == 0 {
		config.ThrowError("没有指定服务器", config.ErrorConfig)
	}
}

func (do devops) build() {
	if len(do.yaml.Build.Command) == 0 {
		config.ThrowError("没有Build任务", config.ErrorConfig)
	}
	// 执行build命令
	do.spin.Suffix = "⏰ 正在执行Build任务........"
	do.spin.Start()
	config.ExecShell(do.yaml.Build.Command)
	do.spin.Stop()
}

func (do devops) deploy() {
	if len(do.yaml.Deploy.Action) == 0 {
		config.ThrowError("没有deploy任务", config.ErrorConfig)
	}
	var wg sync.WaitGroup
	do.spin.Suffix = "⏰ 正在执行deploy任务........"
	do.spin.Start()
	wg.Add(len(do.yaml.Server))
	for index, server := range do.yaml.Server {
		server := server
		go func() {
			conf := &ssh.ClientConfig{
				Timeout:         time.Second,
				User:            server.User,
				HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			}
			conf.Auth = []ssh.AuthMethod{ssh.Password(server.Password)}
			if len(do.yaml.Deploy.Scp.Command) != 0 {
				config.DoSSHCommand(server.Addr+":"+server.Port, do.yaml.Deploy.Scp.Command, conf)
			}
			client := scp.NewClient(server.Addr+":"+server.Port, conf)
			client.Timeout = 10 * time.Minute
			config.ScpFile(client, do.yaml.Deploy.Scp.FilePath, do.yaml.Deploy.Scp.SshPath[index])
			config.DoSSHCommand(server.Addr+":"+server.Port, do.yaml.Deploy.Action[index].Command, conf)
			wg.Done()
		}()
	}
	wg.Wait()
	do.spin.Stop()
}

func main() {
	devopsTool := new(devops)
	devopsTool.spin = spinner.New(spinner.CharSets[35], 100*time.Millisecond)
	devopsTool.yaml = config.GetYamlInfo(devopsTool.getResult())
	//初始化配置
	devopsTool.initSettings()
	//执行build
	devopsTool.build()
	//执行deploy
	devopsTool.deploy()
	color.Cyan("所有任务执行完毕✅ \n")
}
