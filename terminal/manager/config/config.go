package config

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/daqnext/meson-common/common/logger"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

const (
	Token        string = "token"
	Port         string = "port"
	SpaceLimit   string = "spacelimit"
	ServerDomain string = "serverdomain"
	ApiProto     string = "apiProto"
	LogLevel     string = "loglevel"
	GinMode      string = "ginMode"
)

var ConfigPath string
var configMap = map[string]string{}
var (
	token        string
	port         string
	spacelimit   int
	serverdomain string
)

var (
	UsingToken        string
	UsingPort         string
	UsingSpaceLimit   int
	UsingServerDomain string
)

func init() {
	ReadConfig()
}

func RecordTokenAndPortToFile(token string, port string) {
	of, err := os.Open(ConfigPath)
	if err != nil {
		logger.Error("open file record token and port error", "err")
		return
	}
	defer func() {
		if of != nil {
			of.Close()
		}
	}()

	nf, err := os.OpenFile(ConfigPath+".mdf", os.O_RDWR|os.O_CREATE, 0766)
	if err != nil {
		logger.Error("open file record token and port error", "err")
		return
	}
	defer func() {
		if nf != nil {
			of.Close()
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

		newLine = overWriteLine(newLine, "token", token)
		newLine = overWriteLine(newLine, "port", port)

		_, err = nf.WriteString(newLine + "\n")
		if err != nil {
			fmt.Println("write to file fail:", err)
			return
		}
	}

	of.Close()
	of = nil
	nf.Close()
	nf = nil

	err = os.Remove(ConfigPath)
	if err == nil {
		os.Rename(ConfigPath+".mdf", ConfigPath)
	}

}

func overWriteLine(line string, vname string, v string) string {
	if strings.Index(line, "#") == 0 {
		return line
	}
	index := strings.Index(line, "=")
	if index < 0 {
		return line
	}
	first := strings.TrimSpace(line[:index])
	if len(first) == 0 {
		return line
	}

	if first == vname {
		new := vname + " = " + v
		return new
	}
	return line
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

	UsingServerDomain = serverdomain
	if UsingServerDomain == "" {
		UsingServerDomain = GetString(ServerDomain)
	}
}

func CheckConfig() {
	//if did not get token, user need input token
	var mytoken string
	if UsingToken == "" {
		fmt.Println("can not find your token. Please login https://meson.network")
		fmt.Printf("Please enter your token: ")
		fmt.Scanln(&mytoken)
		UsingToken = mytoken
	}

	var myport string
	if UsingPort == "80" || UsingPort == "443" {
		fmt.Printf("CAN NOT use port " + UsingPort + " ,please input a new port \n")
		UsingPort = ""
	}

	if UsingPort == "" {
		fmt.Printf("Please enter your port,CAN NOT be 80 or 443(default 19091): ")
		fmt.Scanln(&myport)
		num, err := strconv.Atoi(myport)
		if err != nil {
			UsingPort = "19091"
			fmt.Println("input port error,server will be run in port:19091")
			return
		}
		if num < 1 || num > 65535 {
			UsingPort = "19091"
			fmt.Println("input port error,server will be run in port:19091")
			return
		}
		if num == 80 || num == 443 {
			UsingPort = "19091"
			fmt.Printf("port CAN NOT be %d,server will be run in port:19091 \n", num)
			return
		}
		UsingPort = myport
	}

}

func ReadFlag() {
	flag.StringVar(&token, Token, "", "token register and login in https://meson.network")
	flag.StringVar(&port, Port, "", "server port")
	flag.IntVar(&spacelimit, SpaceLimit, 0, "cdu space use limit")
	flag.StringVar(&serverdomain, ServerDomain, "", "server domain")
	//flag.Parse()
}

func ReadConfigFile() {
	flag.StringVar(&ConfigPath, "config", "./config.txt", "path to config file")
	flag.Parse()
	if len(ConfigPath) == 0 {
		log.Fatalln("failed to find config file, please provide config file!")
		return
	}
	loadConfigFromTxt(ConfigPath)

	SetDefault(Token, "")
	SetDefault(Port, "")
	SetDefault(ServerDomain, "https://coldcdn.com")
	SetDefault(SpaceLimit, "200")
	SetDefault(ApiProto, "https")
	SetDefault(LogLevel, "4")
	SetDefault(GinMode, "release")
}
func loadConfigFromTxt(configPath string) {
	f, err := os.Open(configPath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

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
