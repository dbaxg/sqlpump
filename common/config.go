package common

import (
    "flag"
    "fmt"
    "os"
    "strings"
)

const (
    path_root = "/usr/local/myaudit"
)

type Configuration struct {
    FileType   string
    FileName   string
    UserName   string
    Password   string
    TestDSN    string
    ReportType string
    Path_root  string
    Path_file  string
    Path_sh    string
    Path_lib   string
    Path_log   string
    Path_sql   string
    Path_audit string
}

var Config = Configuration{
    Path_root:  path_root,
    Path_file:  path_root + "/file",
    Path_sh:    path_root + "/sh",
    Path_lib:   path_root + "/lib",
    Path_log:   path_root + "/log/mapperLog",
    Path_sql:   path_root + "/sql",
    Path_audit: path_root + "/audit",
}

func ReadCmdPram() {
    flag.StringVar(&Config.FileType, "ftype", "mapper", "file type, support 'mapper', 'slowlog' and 'normal', default 'mapper'")
    flag.StringVar(&Config.FileName, "fname", "MapperDB1", "file waitting for audit ~")
    flag.StringVar(&Config.UserName, "u", "root", "database user ~")
    flag.StringVar(&Config.Password, "p", "123456", "user password ~")
    flag.StringVar(&Config.TestDSN, "conn", "", "connect string of database, like 127.0.0.1:3306/sakila ~")
    flag.StringVar(&Config.ReportType, "report-type", "html", "report type, support 'html' and 'json', default 'html'.")
    flag.Parse()
    VerifyPram(Config)
}

func VerifyPram(config Configuration) {
    if config.FileName == "" || config.UserName == "" || config.Password == "" || config.TestDSN == "" {
        Help()
        os.Exit(1)
    } else if !(strings.Contains(config.FileType, "mapper") || strings.Contains(config.FileType, "slowlog") || strings.Contains(config.FileType, "normal")) {
        Help()
        os.Exit(1)
    } else if !(strings.Contains(config.ReportType, "html") || strings.Contains(config.ReportType, "json")) {
        Help()
        os.Exit(1)
    }
}

func Help() {
    fmt.Println("Wrong parameters!")
    fmt.Println("Usage:")
    fmt.Println("   -ftype mapper|slowlog|normal, default mapper")
    fmt.Println("   -fname file waitting for audit ~")
    fmt.Println("   -conn $IP:$PORT/$DB, like 127.0.0.1:3306/sakila ~")
    fmt.Println("   -u username")
    fmt.Println("   -p password")
    fmt.Println("   -report-type html|json, default html")
    fmt.Println("   Example: myaudit -ftype mapper -fname MapperRole -conn 127.0.0.1:3306/sakila -u root -p 123456 -report-type html")
}

