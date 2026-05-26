package env

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

var (
	active Environment
	dev    Environment = &environment{value: "dev"}
	fat    Environment = &environment{value: "fat"}
	uat    Environment = &environment{value: "uat"}
	pro    Environment = &environment{value: "pro"}
)

var _ Environment = (*environment)(nil)

// Environment 环境配置
type Environment interface {
	Value() string
	IsDev() bool
	IsFat() bool
	IsUat() bool
	IsPro() bool
	t()
}

type environment struct {
	value string
}

func (e *environment) Value() string {
	return e.value
}

func (e *environment) IsDev() bool {
	return e.value == "dev"
}

func (e *environment) IsFat() bool {
	return e.value == "fat"
}

func (e *environment) IsUat() bool {
	return e.value == "uat"
}

func (e *environment) IsPro() bool {
	return e.value == "pro"
}

func (e *environment) t() {}

func init() {
	// 定义 flag，但不调用 Parse
	flag.String("env", "", "请输入运行环境:\n dev:开发环境\n fat:测试环境\n uat:预上线环境\n pro:正式环境\n")
}

// Active 当前配置的env
func Active() Environment {
	if active == nil {
		// Avoid flag.Parse during init; detect -env from args to select config early.
		if envVal := envFromArgs(os.Args); envVal != "" {
			setActive(envVal)
		} else {
			active = fat
		}
	}
	return active
}

// Init 显式初始化，应该在 main.go 中调用
func Init() {
	if !flag.Parsed() {
		flag.Parse()
	}

	env := flag.Lookup("env")
	var envVal string
	if env != nil {
		envVal = env.Value.String()
	}

	setActive(envVal)
}

func setActive(envVal string) {
	switch strings.ToLower(strings.TrimSpace(envVal)) {
	case "dev":
		active = dev
	case "fat":
		active = fat
	case "uat":
		active = uat
	case "pro":
		active = pro
	default:
		active = fat
		if strings.TrimSpace(envVal) != "" {
			fmt.Println("Warning: '-env' cannot be found, or it is illegal. The default 'fat' will be used.")
		}
	}
}

func envFromArgs(args []string) string {
	for i := 0; i < len(args); i++ {
		arg := strings.TrimSpace(args[i])
		if arg == "-env" && i+1 < len(args) {
			return strings.TrimSpace(args[i+1])
		}
		if strings.HasPrefix(arg, "-env=") {
			return strings.TrimSpace(strings.TrimPrefix(arg, "-env="))
		}
	}
	return ""
}
