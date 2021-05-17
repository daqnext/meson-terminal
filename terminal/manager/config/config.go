package config

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/daqnext/meson-common/common/logger"
	"github.com/daqnext/meson-common/common/runpath"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

const (
	Token      string = "token"
	Port       string = "port"
	SpaceLimit string = "spacelimit"
	Server     string = "server"
	ApiProto   string = "apiProto"
	LogLevel   string = "loglevel"
	GinMode    string = "ginMode"
)

var ConfigPath string
var configMap = map[string]string{}
var (
	token      string
	port       string
	spacelimit int
	server     string
)

var (
	UsingToken      string
	UsingPort       string
	UsingSpaceLimit int
	Using           string
)

var fileLock sync.Mutex

func init() {
	ReadConfig()
}

func RecordConfigLineToFile(configName string, value string) {
	fileLock.Lock()
	defer fileLock.Unlock()

	of, err := os.Open(ConfigPath)
	if err != nil {
		logger.Error("open file record token and port error", "err")
		return
	}
	defer func() {
		if of != nil {
			e := of.Close()
			if e != nil {
				logger.Error("close config file error", "err", e)
			}
		}
	}()

	nf, err := os.OpenFile(ConfigPath+".mdf", os.O_RDWR|os.O_CREATE, 0766)
	if err != nil {
		logger.Error("open file record token and port error", "err")
		return
	}
	defer func() {
		if nf != nil {
			e := nf.Close()
			if e != nil {
				logger.Error("close config file error", "err", e)
			}
		}
	}()

	r := bufio.NewReader(of)
	isNewField := true
	for {
		originLine, _, err := r.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			logger.Error("readline error", "err", err)
			return
		}

		newLine := string(originLine)

		modify := false
		newLine, modify = overWriteLine(newLine, configName, value)
		if modify == true {
			isNewField = false
		}

		_, err = nf.WriteString(newLine + "\n")
		if err != nil {
			fmt.Println("write to file fail:", err)
			return
		}
	}

	if isNewField {
		newLine := configName + " = " + value
		_, err = nf.WriteString(newLine + "\n")
		if err != nil {
			fmt.Println("write to file fail:", err)
			return
		}
	}

	e := of.Close()
	if e != nil {
		logger.Error("close config file error", "err", e)
	}
	of = nil

	e = nf.Close()
	if e != nil {
		logger.Error("close config file error", "err", e)
	}
	nf = nil

	err = os.Remove(ConfigPath)
	if err == nil {
		e := os.Rename(ConfigPath+".mdf", ConfigPath)
		if e != nil {
			logger.Error("Rename config file error", "err", e)
		}
	}
}

func RecordConfigToFile(configs map[string]string) error {
	fileLock.Lock()
	defer fileLock.Unlock()

	of, err := os.Open(ConfigPath)
	if err != nil {
		logger.Error("open file record token and port error", "err", err)
		return err
	}
	defer func() {
		if of != nil {
			e := of.Close()
			if e != nil {
				logger.Error("close config file error", "err", e)
			}
		}
	}()

	nf, err := os.OpenFile(ConfigPath+".mdf", os.O_RDWR|os.O_CREATE, 0766)
	if err != nil {
		logger.Error("open file record token and port error", "err")
		return err
	}
	defer func() {
		if nf != nil {
			e := nf.Close()
			if e != nil {
				logger.Error("close config file error", "err", e)
			}
		}
	}()

	r := bufio.NewReader(of)
	configMapCopy := map[string]string{}
	for k, v := range configs {
		configMapCopy[k] = v
	}

	for {
		originLine, _, err := r.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			logger.Error("readline error", "err", err)
			return err
		}

		newLine := string(originLine)

		for k, v := range configs {
			modify := false
			newLine, modify = overWriteLine(newLine, k, v)
			if modify == true {
				delete(configMapCopy, k)
			}
		}

		_, err = nf.WriteString(newLine + "\n")
		if err != nil {
			fmt.Println("write to file fail:", err)
			return err
		}
	}

	for k, v := range configMapCopy {
		newLine := k + " = " + v
		_, err = nf.WriteString(newLine + "\n")
		if err != nil {
			fmt.Println("write to file fail:", err)
			return err
		}
	}

	e := of.Close()
	if e != nil {
		logger.Error("close config file error", "err", e)
		return e
	}
	of = nil
	e = nf.Close()
	if e != nil {
		logger.Error("close config file error", "err", e)
		return e
	}
	nf = nil

	err = os.Remove(ConfigPath)
	if err == nil {
		e := os.Rename(ConfigPath+".mdf", ConfigPath)
		if e != nil {
			logger.Error("Rename config file error", "err", e)
			return e
		}
	}
	return nil
}

func RecordUserInputConfigToFile(token string, port string, space string) {
	fileLock.Lock()
	defer fileLock.Unlock()

	of, err := os.Open(ConfigPath)
	if err != nil {
		logger.Error("open file record token and port error", "err")
		return
	}
	defer func() {
		if of != nil {
			e := of.Close()
			if e != nil {
				logger.Error("close config file error", "err", e)
			}
		}
	}()

	nf, err := os.OpenFile(ConfigPath+".mdf", os.O_RDWR|os.O_CREATE, 0766)
	if err != nil {
		logger.Error("open file record token and port error", "err")
		return
	}
	defer func() {
		if nf != nil {
			e := nf.Close()
			if e != nil {
				logger.Error("close config file error", "err", e)
			}
		}
	}()

	r := bufio.NewReader(of)
	for {
		originLine, _, err := r.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			logger.Error("readline error", "err", err)
			return
		}

		newLine := string(originLine)

		newLine, _ = overWriteLine(newLine, Token, token)
		newLine, _ = overWriteLine(newLine, Port, port)
		newLine, _ = overWriteLine(newLine, SpaceLimit, space)

		_, err = nf.WriteString(newLine + "\n")
		if err != nil {
			fmt.Println("write to file fail:", err)
			return
		}
	}

	e := of.Close()
	if e != nil {
		logger.Error("close config file error", "err", e)
	}
	of = nil
	e = nf.Close()
	if e != nil {
		logger.Error("close config file error", "err", e)
	}
	nf = nil

	err = os.Remove(ConfigPath)
	if err == nil {
		e := os.Rename(ConfigPath+".mdf", ConfigPath)
		if e != nil {
			logger.Error("Rename config file error", "err", e)
		}
	}
}

func overWriteLine(line string, vname string, v string) (string, bool) {
	if strings.Index(line, "#") == 0 {
		return line, false
	}
	index := strings.Index(line, "=")
	if index < 0 {
		return line, false
	}
	first := strings.TrimSpace(line[:index])
	if len(first) == 0 {
		return line, false
	}

	if first == vname {
		newLine := vname + " = " + v
		return newLine, true
	}
	return line, false
}

func ReadConfig() {
	//cmd line
	ReadFlag()
	//config file
	ReadConfigFile()

	UsingToken = token
	if UsingToken == "" {
		UsingToken = GetString(Token)
	}

	UsingPort = port
	if UsingPort == "" {
		UsingPort = GetString(Port)
	}

	UsingSpaceLimit = spacelimit
	if UsingSpaceLimit == 0 {
		UsingSpaceLimit = GetInt(SpaceLimit)
	}

	Using = server
	if Using == "" {
		Using = GetString(Server)
	}
}

func CheckConfig() {
	//if did not get token, user need input token
	var mytoken string
	if UsingToken == "" {
		fmt.Println("can not find your token. Please login https://meson.network")
		fmt.Printf("Please enter your token: ")
		_, err := fmt.Scanln(&mytoken)
		if err != nil {
			log.Fatalln("read input token error")
		}
		UsingToken = mytoken
	}

	var myport string
	if UsingPort == "80" || UsingPort == "443" {
		fmt.Printf("CAN NOT use port " + UsingPort + " ,please input a new port \n")
		UsingPort = ""
	}

	if UsingPort == "" {
		fmt.Printf("Please enter your port,CAN NOT be 80 or 443(default 19091): ")
		_, err := fmt.Scanln(&myport)
		if err != nil {
			UsingPort = "19091"
			fmt.Println("read input port error,server will be run in port:19091.You can modify this value in config.txt")
		}
		num, err := strconv.Atoi(myport)
		if err != nil {
			UsingPort = "19091"
			fmt.Println("input port error,server will be run in port:19091.You can modify this value in config.txt")
			//return
		} else {
			UsingPort = myport
			if num < 1 || num > 65535 {
				UsingPort = "19091"
				fmt.Println("input port error,server will be run in port:19091.You can modify this value in config.txt")
				//return
			}
			if num == 80 || num == 443 {
				UsingPort = "19091"
				fmt.Printf("port CAN NOT be %d,server will be run in port:19091.You can modify this value in config.txt \n", num)
				//return
			}
		}
	}

	var space string
	if UsingSpaceLimit == 0 {
		fmt.Println("Please input the disk space you want to provide.The more space you provide, the higher profit you will get")
		fmt.Printf("For example if you provide 100GB, please input 100 (40GB disk space is the minimum, default will be 80GB):")
		_, err := fmt.Scanln(&space)
		if err != nil {
			UsingSpaceLimit = 80
			fmt.Println("read input error,server will use default 80G.You can modify this value in config.txt")
			return
		}
		num, err := strconv.Atoi(space)
		if err != nil {
			UsingSpaceLimit = 80
			fmt.Println("input space error,server will use default 80G.You can modify this value in config.txt")
			return
		}
		UsingSpaceLimit = num
	}

}

func ReadFlag() {
	flag.StringVar(&token, Token, "", "token register and login in https://meson.network")
	flag.StringVar(&port, Port, "", "server port")
	flag.IntVar(&spacelimit, SpaceLimit, 0, "cdn space use limit")
	flag.StringVar(&server, Server, "", "server")
	//flag.Parse()
}

func ReadConfigFile() {
	configPath := filepath.Join(runpath.RunPath, "./config.txt")
	flag.StringVar(&ConfigPath, "config", configPath, "path to config file")
	flag.Parse()
	if len(ConfigPath) == 0 {
		log.Fatalln("failed to find config file, please provide config file!")
		return
	}
	loadConfigFromTxt(ConfigPath)

	SetDefault(Token, "")
	SetDefault(Port, "")
	SetDefault(Server, "http://coldcdn.com")
	SetDefault(SpaceLimit, "0")
	SetDefault(ApiProto, "https")
	SetDefault(LogLevel, "4")
	SetDefault(GinMode, "release")
}
func loadConfigFromTxt(configPath string) {
	fileLock.Lock()
	defer fileLock.Unlock()

	f, err := os.Open(configPath)
	if err != nil {
		logger.Error("open config file error", "err", err)
		return
	}
	defer func() {
		if f != nil {
			e := f.Close()
			if e != nil {
				logger.Error("close config file error", "err", e)
			}
		}
	}()

	r := bufio.NewReader(f)
	for {
		b, _, err := r.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}

		s := strings.TrimSpace(string(b))
		//fmt.Println(s)
		if strings.Index(s, "#") == 0 {
			continue
		}

		index := strings.Index(s, "=")
		if index < 0 {
			continue
		}

		frist := strings.TrimSpace(s[:index])
		if len(frist) == 0 {
			continue
		}
		second := strings.TrimSpace(s[index+1:])

		pos := strings.Index(second, "\t#")
		if pos > -1 {
			second = second[0:pos]
		}

		pos = strings.Index(second, " #")
		if pos > -1 {
			second = second[0:pos]
		}

		if len(second) == 0 {
			continue
		}

		key := frist
		configMap[key] = strings.TrimSpace(second)
	}
}

func SetDefault(key, defaultValue string) {
	_, exist := configMap[key]
	if exist {
		return
	}
	configMap[key] = defaultValue
}

func GetString(key string) string {
	value, exist := configMap[key]
	if !exist {
		return ""
	}
	return value
}

func GetInt(key string) int {
	value, exist := configMap[key]
	if !exist {
		return 0
	}
	num, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return num
}
