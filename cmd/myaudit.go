package main

import (
    "database/sql"
    "fmt"
    "strings"

    _ "github.com/go-sql-driver/mysql"
    "github.com/myaudit/common"
    "github.com/myaudit/magic"
)

func main() {
    common.ReadCmdPram()
    config := &common.Config
    fmt.Println(config)

    // 获取select_id及其对应sql
    id_sql := magic.ParseMapper(config.FileName, config.Path_root, config.Path_file, config.Path_sh, config.Path_lib, config.Path_log, config.TestDSN, config.UserName, config.Password)

    path := config.UserName + ":" + config.Password + "@tcp(" + strings.Split(config.TestDSN, "/")[0] + ")/information_schema?charset=utf8"
    fmt.Println(path)
    db, _ := sql.Open("mysql", path)
    if err := db.Ping(); err != nil {
        fmt.Println("Open database failed!")
        return
    } else {
        fmt.Println("Database connnect success!")
    }
    defer db.Close()

    // 将sql按select_id分别写入文件并返回select_id
    select_id := magic.WriteSQL2File(id_sql, config.Path_sql, config.FileName, db, strings.Split(config.TestDSN, "/")[1])

    // 按select_id进行审计并生成对应审计报告
    magic.MapperAudit(config.TestDSN, config.UserName, config.Password, select_id, config.Path_sql, config.Path_audit, config.FileName, config.ReportType)
}
