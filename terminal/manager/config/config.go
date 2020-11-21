package config

import (
	"bufio"
	"flag"
	"fmt"
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

//var V *viper.Viper
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

func ReadConfig() {
	//先读取命令行
	ReadFlag()
	//读取配置文件
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
	//如果用户没有输入token 提示用户输入token
	var mytoken string
	if UsingToken == "" {
		fmt.Println("can not find your token. Please login https://meson.network")
		fmt.Printf("Please enter your token: ")
		fmt.Scanln(&mytoken)
		UsingToken = mytoken
	}

	var myport string
	if UsingPort == "" {
		fmt.Printf("Please enter your port,CAN NOT be 80 or 443(default 19091): ")
		fmt.Scanln(&myport)
		num, err := strconv.Atoi(myport)
		if err != nil {
			UsingPort = "19091"
			fmt.Println("input port error,server will be run in port:19091")
			return
		}
		if num < 0 || num > 65535 {
			UsingPort = "19091"
			fmt.Println("input port error,server will be run in port:19091")
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
	var configPath string
	flag.StringVar(&configPath, "config", "./config.txt", "path to config file")
	flag.Parse()
	if len(configPath) == 0 {
		log.Fatalln("failed to find config file, please provide config file!")
		return
	}
	loadConfigFromTxt(configPath)

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
