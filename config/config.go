package config

import (
	"bufio"
	"errors"
	"github.com/bramvdbogaerde/go-scp"
	"github.com/fatih/color"
	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"sync"
)

var (
	ErrorConfig = errors.New("error")
)

type build struct {
	Command []string `yaml:"command"`
	Scp     struct {
		Command  []string `yaml:"command"`
		FilePath string   `yaml:"filePath"`
		SshPath  string   `yaml:"sshPath"`
	} `yaml:"scp"`
}

type deploy struct {
	Action []struct {
		Command []string `yaml:"command"`
	} `yaml:"action"`
	Scp struct {
		Command  []string `yaml:"command"`
		FilePath string   `yaml:"filePath"`
		SshPath  []string `yaml:"sshPath"`
	} `yaml:"scp"`
}

type serverConfig struct {
	Addr     string `yaml:"addr"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

type Setting struct {
	Name   string `yaml:"name"`
	Server []serverConfig
	Build  build  `yaml:"build"`
	Deploy deploy `yaml:"deploy"`
}

func DoSSHCommand(addr string, command []string, config *ssh.ClientConfig) {
	if len(command) == 0 {
		return
	}
	sshClient, err := ssh.Dial("tcp", addr, config)
	ThrowError("创建ssh client 失败", err)
	defer sshClient.Close()
	session, err := sshClient.NewSession()
	ThrowError("创建ssh session 失败", err)
	defer session.Close()

	//执行远程命令
	cmdList := ""
	for _, cmd := range command {
		cmdList += cmd + ";"
	}
	combo, err := session.CombinedOutput(cmdList)
	ThrowError("远程执行cmd 失败", err)
	color.HiYellow(string(combo))
}

func GetProjects() []string {
	files, _ := ioutil.ReadDir("yaml")
	var list []string
	for _, f := range files {
		list = append(list, f.Name())
	}
	return list
}

func ExecShell(command []string) {
	var data string
	for _, value := range command {
		data += value + "\n"
	}
	cmd := exec.Command("bash", "-c", data)
	stdout, err := cmd.StdoutPipe()
	ThrowError("执行shell失败", err)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		reader := bufio.NewReader(stdout)
		for {
			readString, err := reader.ReadString('\n')
			if err != nil || err == io.EOF {
				return
			}
			color.Green(readString)
			if strings.Contains(readString, "No such file or directory") {
				ThrowError("发生错误，没有该目录或文件", ErrorConfig)
			}
		}
	}()
	err = cmd.Start()
	wg.Wait()
}

func GetYamlInfo(name string) Setting {
	file, err := ioutil.ReadFile("yaml/" + name)
	ThrowError("读取yaml失败", err)
	//yaml文件内容影射到结构体中
	var data Setting
	err1 := yaml.Unmarshal(file, &data)
	ThrowError("解析yaml文件失败", err1)
	return data
}

func ScpFile(client scp.Client, filePath, sshPath string) {
	err := client.Connect()
	ThrowError("scp连接失败", err)
	f, err1 := os.Open(filePath)
	ThrowError("打开文件失败", err1)
	defer client.Close()
	defer f.Close()
	err = client.CopyFile(f, sshPath, "0755")
	ThrowError("传输文件错误", err)
}

func ThrowError(format string, err error) {
	if err == nil {
		return
	}
	color.Red(format, err)
	os.Exit(-1)
}
