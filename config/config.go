/*
 * Copyright 2020 myaudit Author. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package config

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/dbaxg/myaudit/log"
	"runtime/debug"
	"time"
	"strconv"
	"os/exec"
)

type configuration struct {
	Path *path
	Parm *parm
}

type parm struct {
	Filename  string
	Username  string
	Password  string
	TestDSN   string
	FileConf  string
	Timestamp string
	Help      bool
}

type path struct {
	PathRoot    string
	PathLib     string
	PathLog     string
	PathSql     string
	PathMybatis string
}

var Config = configuration{
	&path{},
	&parm{},
}

//parameters用于接收命令行传入的参数
var parameters parm

/*
    1.依赖
    myaudit依赖log4j-1.2.17.jar，mybatis-3.5.4.jar，mysql-connector-java-5.1.47.jar等3个jar包，
    这3个jar包存放于本项目的mybatis目录下。myaudit首次运行时会进行初始化，在PathRoot路径下创建lib子目录，
    并从$GOPATH/src/github.com/myaudit/mybatis/下拷贝这3个jar包至lib子目录。
    如果你的myaudit binary没有和myaudit project在同一台服务器，你需要手动将jar包拷贝至lib子目录（myaudit会给出提示）。

    2.配置文件和参数
    默认配置文件为/usr/etc/myaudit.toml，用户也可通过-c参数来指定配置文件。
    命令行传入的参数将覆盖配置文件中的参数。
 */

//读取参数
func ReadParms() {
	//覆盖flag自带的usege方法
	flag.Usage = usage
	flag.StringVar(&parameters.Filename, "f", "", "file waitting for audit ~")
	flag.StringVar(&parameters.Username, "u", "", "database user ~")
	flag.StringVar(&parameters.Password, "p", "gzz1992dba2020", "user password ~")
	flag.StringVar(&parameters.TestDSN, "s", "", "connect string of database, like 127.0.0.1:3306/sakila ~")
	flag.StringVar(&parameters.FileConf, "c", "/etc/myaudit.toml", "configuration file defined by users.")
	flag.BoolVar(&Config.Parm.Help, "h", false, "this help")
	flag.Parse()

	if Config.Parm.Help {
		flag.Usage()
		os.Exit(0)
	}

	// xmlName为文件名（不含`.xml`），后续将用于创建Mybatis project
	var xmlName string
	// xmlNameWithTimestamp将xml打上timestamp标记，防止同一个或同名xml文件在并发调用myaudit进行解析时产生错误
	var xmlNameWithTimestamp string

	// 判断用户指定的配置文件（如果不指定，则为默认配置文件）是否存在
	if _, err := os.Stat(parameters.FileConf); err != nil {
		errInfo := "Configuration file does not exist, " + err.Error()
		stackInfo := "\n" + string(debug.Stack()) + "\n"
		fmt.Println("{\n\"resultCode\": 1,\n\"sqlPath\": \"\",\n\"errorInfo\": \"" + errInfo + "\",\n\"panicInfo\": \"\",\n\"stackInfo\": \"" + stackInfo + "\"\n}")
		os.Exit(1)
	}
	if _, err := toml.DecodeFile(parameters.FileConf, &Config); err != nil {
		errInfo := "Something wrong happened when paring the configuration file: " + err.Error()
		stackInfo := "\n" + string(debug.Stack()) + "\n"
		fmt.Println("{\n\"resultCode\": 1,\n\"sqlPath\": \"\",\n\"errorInfo\": \"" + errInfo + "\",\n\"panicInfo\": \"\",\n\"stackInfo\": \"" + stackInfo + "\"\n}")
		os.Exit(1)
	}

	// 如果通过-f等flag传入参数，则覆盖配置文件中的参数
	if parameters.Filename != "" {
		Config.Parm.Filename = parameters.Filename
	}
	if parameters.Username != "" {
		Config.Parm.Username = parameters.Username
	}
	if parameters.Password != "gzz1992dba2020" {
		//因为密码可以为空，所以parameters的password默认值不能为空，此处设为了gzz1992dba2020
		Config.Parm.Password = parameters.Password
	}
	if parameters.TestDSN != "" {
		Config.Parm.TestDSN = parameters.TestDSN
	}

	//校验参数是否合规
	VerifyParms(Config)

	//判断根目录是否存在（根目录需用户自行创建，并在配置文件中配置）
	_, err := os.Stat(Config.Path.PathRoot)
	if err != nil {
		errInfo := "PathRoot does not exist, you need to create the dir first: " + err.Error()
		stackInfo := "\n" + string(debug.Stack()) + "\n"
		fmt.Println("{\n\"resultCode\": 1,\n\"sqlPath\": \"\",\n\"errorInfo\": \"" + errInfo + "\",\n\"panicInfo\": \"\",\n\"stackInfo\": \"" + stackInfo + "\"\n}")
		os.Exit(1)
	}

	//为避免并发对同一xml文件进行解析时，导致后面执行的操作删除前面操作生成的文件，对每次操作加时间戳
	if len(Config.Parm.Filename) < 4 {
		errInfo := "Wrong file name!"
		stackInfo := "\n" + string(debug.Stack()) + "\n"
		fmt.Println("{\n\"resultCode\": 1,\n\"sqlPath\": \"\",\n\"errorInfo\": \"" + errInfo + "\",\n\"panicInfo\": \"\",\n\"stackInfo\": \"" + stackInfo + "\"\n}")
		os.Exit(1)
	}
	if Config.Parm.Filename[len(Config.Parm.Filename)-4:] != ".xml" {
		errInfo := "Wrong file name!"
		stackInfo := "\n" + string(debug.Stack()) + "\n"
		fmt.Println("{\n\"resultCode\": 1,\n\"sqlPath\": \"\",\n\"errorInfo\": \"" + errInfo + "\",\n\"panicInfo\": \"\",\n\"stackInfo\": \"" + stackInfo + "\"\n}")
		os.Exit(1)
	}

	if strings.Contains(Config.Parm.Filename, "/") {
		xmlName = Config.Parm.Filename[strings.LastIndex(Config.Parm.Filename, "/")+1:]
		xmlName = xmlName[:strings.LastIndex(xmlName, ".")]
	} else {
		xmlName = Config.Parm.Filename[:strings.LastIndex(Config.Parm.Filename, ".")]
	}

	//获取时间戳
	TimeStamp := strconv.FormatInt(time.Now().UnixNano()/1e6, 10)

	//将TimeStamp记录下来，后续如果有审核sql的需求时，可以从Config.Parm.Timestamp中获取该次解析的时间戳，并从sql子目录找到该次解析出来的sql文件
	Config.Parm.Timestamp = TimeStamp
	xmlNameWithTimestamp = xmlName + "-" + TimeStamp

	//配置子目录
	Config.Path.PathLib = Config.Path.PathRoot + "/lib"
	Config.Path.PathLog = Config.Path.PathRoot
	Config.Path.PathSql = Config.Path.PathRoot + "/sql/" + xmlNameWithTimestamp
	Config.Path.PathMybatis = Config.Path.PathRoot + "/tmp/" + xmlNameWithTimestamp + "/mybatis"

	//判断xml文件是否存在
	if _, err := os.Stat(Config.Parm.Filename); err != nil {
		errInfo := "xml file does not exist, " + err.Error()
		stackInfo := "\n" + string(debug.Stack()) + "\n"
		fmt.Println("{\n\"resultCode\": 1,\n\"sqlPath\": \"\",\n\"errorInfo\": \"" + errInfo + "\",\n\"panicInfo\": \"\",\n\"stackInfo\": \"" + stackInfo + "\"\n}")
		os.Exit(1)
	}

	//初始化
	initialize(Config.Path.PathLib, Config.Path.PathSql, Config.Path.PathMybatis)

	//设置日志参数，在此之前发生error将无法被记录，将通过fmt.Println的方式打印错误信息和堆栈信息
	var jsonConfig = `{
						"filename" : "` + Config.Path.PathLog + "/myaudit.log\"" + `,
						"maxlines" : 100000,
						"maxsize"  : 10240000
                           }`
	log.LogSetting(jsonConfig)

	//设置日志的前缀，用于标记日志是由哪一个xml文件、哪次解析产生的
	log.Log.SetPrefix("[" + xmlNameWithTimestamp + "]")
}

//校验参数是否正确
func VerifyParms(config configuration) {
	if config.Parm.Filename == "" {
		//fmt.Println("Wrong parameters: filename is empty, please check!\n")
		errInfo := "Wrong parameters: filename is empty, please check!"
		stackInfo := "\n" + string(debug.Stack()) + "\n"
		fmt.Println("{\n\"resultCode\": 1,\n\"sqlPath\": \"\",\n\"errorInfo\": \"" + errInfo + "\",\n\"panicInfo\": \"\",\n\"stackInfo\": \"" + stackInfo + "\"\n}")
		//usage()
		os.Exit(1)

	}
	if config.Parm.Username == "" {
		//fmt.Println("Wrong parameters: username is empty, please check!\n")
		errInfo := "Wrong parameters: username is empty, please check!"
		stackInfo := "\n" + string(debug.Stack()) + "\n"
		fmt.Println("{\n\"resultCode\": 1,\n\"sqlPath\": \"\",\n\"errorInfo\": \"" + errInfo + "\",\n\"panicInfo\": \"\",\n\"stackInfo\": \"" + stackInfo + "\"\n}")
		//usage()
		os.Exit(1)
	}
	if config.Parm.TestDSN == "" {
		//fmt.Println("Wrong parameters: connection string is empty, please check!\n")
		errInfo := "Wrong parameters: connection string is empty, please check!"
		stackInfo := "\n" + string(debug.Stack()) + "\n"
		fmt.Println("{\n\"resultCode\": 1,\n\"sqlPath\": \"\",\n\"errorInfo\": \"" + errInfo + "\",\n\"panicInfo\": \"\",\n\"stackInfo\": \"" + stackInfo + "\"\n}")
		//usage()
		os.Exit(1)
	}
}

//myaudit使用说明
func usage() {
	fmt.Fprintf(os.Stderr, `version: myaudit-2.0
Usage: myaudit [-h] [-t mapper] [-f filename] [-s connStr] [-u username] [-p password] [-c fileConf]
Example: myaudit -f mapperTest.xml -s 127.0.0.1:3306/sakila -u xxx -p xxx -c /usr/etc/myaudit.toml
Options:
`)
	fmt.Println("   -h show the usage of myaudit ~")
	fmt.Println("   -f file to parse ~")
	fmt.Println("   -s $IP:$PORT/$DB, like 127.0.0.1:3306/sakila ~")
	fmt.Println("   -u database username ~")
	fmt.Println("   -p database password ~")
	fmt.Println("   -c configuration file, default `/usr/etc/myaudit.toml`~")
	fmt.Println("Tips: If you don't declare these parameters above, myaudit will use the parameters in the configuration file.")
}

//初始化子目录和拷贝依赖jar包
func initialize(PathLib string, PathSql string, PathMybatis string) {

	// 在PathRoot下创建lib子目录
	// 注：只会在首次执行myaudit创建
	if _, err := os.Stat(PathLib); err != nil {
		os.MkdirAll(PathLib, os.ModePerm)
	}

	// 由于本项目依赖log4j-1.2.17.jar，mybatis-3.5.4.jar，mysql-connector-java-5.1.47.jar等jar包，
	// 如果lib子目录下的3个jar包有缺失，则获取GOPATH环境变量，并从$GOPATH/src/github.com/dba/myaudit/mybatis下拷贝
	// 如果用户提前在PathRoot下创建了lib子目录，并上传了上述3个jar包，则此步不会执行。
	copyJar(PathLib)

	// 在PathRoot/sql下创建xmlNameWithTimestamp子目录，用于存放解析生成的sql
	if _, err := os.Stat(PathSql); err != nil {
		os.MkdirAll(PathSql, os.ModePerm)
	}

	// 在PathRoot/tmp/xmlNameWithTimestamp下创建mybatis子目录，用于存放生成的Mybatis project
	if _, err := os.Stat(PathMybatis); err != nil {
		os.MkdirAll(PathMybatis, os.ModePerm)
	}
}

// 拷贝$GOPATH/src/github.com/dba/myaudit/mybatis下的jar包至PathLib
func copyJar(PathLib string) {
	_, errLog4j := os.Stat(PathLib + "/log4j-1.2.17.jar")
	_, errMybatis := os.Stat(PathLib + "/mybatis-3.5.4.jar")
	_, errMysqlConnectorJava := os.Stat(PathLib + "/mysql-connector-java-5.1.47.jar")

	GOPATH := os.Getenv("GOPATH")
	if errLog4j != nil || errMybatis != nil || errMysqlConnectorJava != nil {
		if _, err := os.Stat(GOPATH + "/src/github.com/dbaxg/myaudit"); err != nil {
			//fmt.Println("myaudit project does not exist, the process need to copy the jar files located in $GOPATH/src/github.com/dba/myaudit/mybatis for initialization at the first execution! You can also copy the jar files(log4j-1.2.17.jar,mybatis-3.5.4.jar,mysql-connector-java-5.1.47.jar) to `" + PathLib + "` manually.")
			errInfo := "myaudit project does not exist, the process need to copy the jar files located in $GOPATH/src/github.com/dbaxg/myaudit/mybatis for initialization at the first execution!" +
				" You can also copy the jar files(log4j-1.2.17.jar,mybatis-3.5.4.jar,mysql-connector-java-5.1.47.jar) to `" + PathLib + "` manually."
			stackInfo := "\n" + string(debug.Stack()) + "\n"
			fmt.Println("{\n\"resultCode\": 1,\n\"sqlPath\": \"\",\n\"errorInfo\": \"" + errInfo + "\",\n\"panicInfo\": \"\",\n\"stackInfo\": \"" + stackInfo + "\"\n}")
			os.Exit(1)
		}
	}

	//如果lib子目录中的jar包不存在，则从myaudit project中拷贝
	if errLog4j != nil {
		//判断mybatis project中的jar包是否存在
		if _, err := os.Stat(GOPATH + "/src/github.com/dbaxg/myaudit/mybatis/log4j-1.2.17.jar"); err != nil {
			//errInfo := err.Error() + "\n you need to copy `log4j-1.2.17.jar` to `" + PathLib + "` manually."
			errInfo := "You need to copy `log4j-1.2.17.jar` to `" + PathLib + "` manually: " + err.Error()
			stackInfo := "\n" + string(debug.Stack()) + "\n"
			fmt.Println("{\n\"resultCode\": 1,\n\"sqlPath\": \"\",\n\"errorInfo\": \"" + errInfo + "\",\n\"panicInfo\": \"\",\n\"stackInfo\": \"" + stackInfo + "\"\n}")
			os.Exit(1)
		}
		//执行拷贝
		cmdStr := "cp " + GOPATH + "/src/github.com/dbaxg/myaudit/mybatis/log4j-1.2.17.jar " + PathLib
		cmd := exec.Command("/bin/bash", "-c", cmdStr)
		err := cmd.Run()
		if err != nil {
			errInfo := "copy " + GOPATH + "/src/github.com/dbaxg/myaudit/mybatis/log4j-1.2.17.jar failed: " + err.Error()
			stackInfo := "\n" + string(debug.Stack()) + "\n"
			fmt.Println("{\n\"resultCode\": 1,\n\"sqlPath\": \"\",\n\"errorInfo\": \"" + errInfo + "\",\n\"panicInfo\": \"\",\n\"stackInfo\": \"" + stackInfo + "\"\n}")
			os.Exit(1)
		}
	}

	if errMybatis != nil {
		//判断mybatis project中的jar包是否存在
		if _, err := os.Stat(GOPATH + "/src/github.com/dbaxg/myaudit/mybatis/mybatis-3.5.4.jar"); err != nil {
			errInfo := "You need to copy `mybatis-3.5.4.jar` to `" + PathLib + "` manually: " + err.Error()
			stackInfo := "\n" + string(debug.Stack()) + "\n"
			fmt.Println("{\n\"resultCode\": 1,\n\"sqlPath\": \"\",\n\"errorInfo\": \"" + errInfo + "\",\n\"panicInfo\": \"\",\n\"stackInfo\": \"" + stackInfo + "\"\n}")
			os.Exit(1)
		}
		//执行拷贝
		cmdStr := "cp " + GOPATH + "/src/github.com/dbaxg/myaudit/mybatis/mybatis-3.5.4.jar " + PathLib
		cmd := exec.Command("/bin/bash", "-c", cmdStr)
		err := cmd.Run()
		if err != nil {
			errInfo := "copy " + GOPATH + "/src/github.com/dbaxg/myaudit/mybatis/mybatis-3.5.4.jar failed: " + err.Error()
			stackInfo := "\n" + string(debug.Stack()) + "\n"
			fmt.Println("{\n\"resultCode\": 1,\n\"sqlPath\": \"\",\n\"errorInfo\": \"" + errInfo + "\",\n\"panicInfo\": \"\",\n\"stackInfo\": \"" + stackInfo + "\"\n}")
			os.Exit(1)
		}
	}

	if errMysqlConnectorJava != nil {
		//判断mybatis project中的jar包是否存在
		if _, err := os.Stat(GOPATH + "/src/github.com/dbaxg/myaudit/mybatis/mysql-connector-java-5.1.47.jar"); err != nil {
			errInfo := "you need to copy `mysql-connector-java-5.1.47.jar` to `" + PathLib + "` manually: " + err.Error()
			stackInfo := "\n" + string(debug.Stack()) + "\n"
			fmt.Println("{\n\"resultCode\": 1,\n\"sqlPath\": \"\",\n\"errorInfo\": \"" + errInfo + "\",\n\"panicInfo\": \"\",\n\"stackInfo\": \"" + stackInfo + "\"\n}")
			os.Exit(1)
		}
		//执go行拷贝
		cmdStr := "cp " + GOPATH + "/src/github.com/dbaxg/myaudit/mybatis/mysql-connector-java-5.1.47.jar " + PathLib
		cmd := exec.Command("/bin/bash", "-c", cmdStr)
		err := cmd.Run()
		if err != nil {
			errInfo := "copy " + GOPATH + "/src/github.com/dbaxg/myaudit/mybatis/mysql-connector-java-5.1.47.jar failed: " + err.Error()
			stackInfo := "\n" + string(debug.Stack()) + "\n"
			fmt.Println("{\n\"resultCode\": 1,\n\"sqlPath\": \"\",\n\"errorInfo\": \"" + errInfo + "\",\n\"panicInfo\": \"\",\n\"stackInfo\": \"" + stackInfo + "\"\n}")
			os.Exit(1)
		}
	}
}
