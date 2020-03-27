/*
 * Copyright 2020 sqlpump Author. All Rights Reserved.
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

package main

import (
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/dbaxg/sqlpump/parse"
	"database/sql"
	"github.com/dbaxg/sqlpump/log"
	"fmt"
	"os"
	"github.com/dbaxg/sqlpump/config"
)

func main() {
	// 捕获panic
	defer func() {
		if r := recover(); r != nil {
			//check exactly what the panic was and create error.
			switch x := r.(type) {
			case error:
				errInfo := x.Error()
				stackInfo := "\n" + string(parse.PanicTrace(4)) + "\n"
				fmt.Println("{\n\"resultCode\": 2,\n\"sqlPath\": \"\",\n\"errorInfo\": \"\",\n\"panicInfo\": \"" + errInfo + "\",\n\"stackInfo\": " + stackInfo + "\"\n}")
			}
			os.Exit(1)
		}
	}()

	// 读取参数
	config.ReadParms()
	conf := &config.Config
	//fmt.Println("Starting to parse `" + config.Parm.FileName + "`, it will take for a while.")

	// 连接数据库
	connStr := conf.Parm.Username + ":" + conf.Parm.Password + "@tcp(" + strings.Split(conf.Parm.TestDSN, "/")[0] + ")/information_schema?charset=utf8"
	db, err := sql.Open("mysql", connStr)
	defer db.Close()
	if err != nil {
		log.LogIfError(err, "")
	}
	if err := db.Ping(); err != nil {
		log.LogIfError(err, "")
	}

	// golang里即使用户密码错误，仍然可以正常open数据库，
	// 所以需执行`select 1 from information_schema.columns limit 1`来验证密码和权限
	// sqlpump仅需information_schema.columns表的查询权限
	parse.VerifyDbPass(db)

	// 获取labelId及其对应sql
	idSqlList := parse.ParseMapper(conf.Parm.Filename, conf.Path.PathMybatis, conf.Path.PathLib, conf.Parm.TestDSN, conf.Parm.Username, conf.Parm.Password)

	// 将sql按labelId分别写入文件并返回labelId，如果后续有审核sql的需求，可通过返回的labelId来找到对应的`.sql`文件
	_ = parse.ReplaceAndWriteSQL2File(idSqlList, conf.Path.PathSql, db, strings.Split(conf.Parm.TestDSN, "/")[1])
	//fmt.Println("Parsing completed, the timestamp mark is " + config.Parm.Timestamp + ", you can find the .sql files in directory `" + config.Path.PathSql + "`.")
	fmt.Println("{\n\"resultCode\": 0,\n\"sqlPath\": \"" + conf.Path.PathSql + "\",\n\"errorInfo\": \"\",\n\"panicInfo\": \"\",\n\"stackInfo\": \"\"\n}")

	//// 按labelId进行审计并生成对应审计报告
	// parse.MapperAudit(config.Para.TestDSN, config.Para.UserName, config.Para.Password, labelId, config.Path.PathSql, config.Path.PathAudit, config.Para.FileName, config.Para.ReportType)
}
