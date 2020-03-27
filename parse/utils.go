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

package parse

import (
	"os"
	"io/ioutil"
	"encoding/xml"
	"strings"
	"os/exec"
	"bufio"
	"io"
	"database/sql"
	"github.com/dbaxg/sqlpump/log"
	"github.com/dbaxg/sqlpump/replace"
	"runtime"
	"bytes"
)

// 创建目录
func Makedir(path string) error {
	err := os.RemoveAll(path)
	if err != nil {
		return err
	} else {
		err = os.Mkdir(path, os.ModePerm)
	}
	return err
}

func format(xmlInput string, xmlOutput string, xmlName string) error {
	var t xml.Token
	var err error

	// 读取原始xml
	content, err := ioutil.ReadFile(xmlInput)
	if err != nil {
		log.LogIfError(err, "%s")
	}
	xmlStr := string(content[:])
	inputReader := strings.NewReader(xmlStr)
	decoder := xml.NewDecoder(inputReader)

	// 构建mapperFormated.xml
	mapperFormated, err := os.Create(xmlOutput)
	if err != nil {
		log.LogIfError(err, "%s")
	}
	defer mapperFormated.Close()
	mapperFormated.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	mapperFormated.WriteString("<!DOCTYPE mapper PUBLIC \"-//mybatis.org//DTD Mapper 3.0//EN\" \"http://mybatis.org/dtd/mybatis-3-mapper.dtd\">\n")

	for t, err = decoder.Token(); err == nil; t, err = decoder.Token() {
		switch token := t.(type) {
		// 处理开始标签
		case xml.StartElement:
			name := token.Name.Local
			if name == "mapper" {
				mapperFormated.WriteString("\n<mapper namespace=\"SQLAudit." + xmlName + ".newMapper\">\n")
			} else if len(token.Attr) == 0 {
				mapperFormated.WriteString("\n<" + name + ">\n")
			} else {
				var attrStr = " "
				for _, attr := range token.Attr {
					attrName := attr.Name.Local
					attrValue := strings.Replace(attr.Value, "\n", "", -1)
					attrStr = attrStr + attrName + "=" + "\"" + attrValue + "\" "
				}
				mapperFormated.WriteString("\n<" + name + attrStr + ">\n")
			}
			// 处理结束标签
		case xml.EndElement:
			name := token.Name.Local
			mapperFormated.WriteString("\n</" + name + ">\n")
			// 处理字符数据
		case xml.CharData:
			content := string([]byte(token))
			if strings.Contains(content, "<") || strings.Contains(content, ">") {
				cdataStr := "<![CDATA[ " + strings.TrimSpace(content) + " ]]>"
				if strings.Contains(cdataStr, "#") {
					mapperFormated.WriteString(strings.Replace(cdataStr, "#", "\n#", -1))
				} else {
					mapperFormated.WriteString(cdataStr)
				}
			} else {
				if strings.Contains(content, "#") {
					mapperFormated.WriteString(strings.Replace(strings.TrimSpace(content), "#", "\n#", -1))
				} else {
					mapperFormated.WriteString(strings.TrimSpace(content))
				}
			}
		default:
			// do nothing
		}
	}
	if err == io.EOF {
		err = nil
	}
	return err
}

// 编译并执行创建好的mybatis project
func execMybatisProject(pathMybatis string, xmlName string) error {
	cmdStr := "sh " + pathMybatis + "/" + xmlName + ".sh"
	cmd := exec.Command("/bin/bash", "-c", cmdStr)
	err := cmd.Run()
	return err
}

// 从Mybatis日志中获取SQL
func getSQL(logPath string) ([][]string, error) {
	sqlLog, err := os.Open(logPath)
	if err != nil {
		log.LogIfError(err, "")
	}
	bufTmp := bufio.NewReader(sqlLog)
	defer sqlLog.Close()
	var id_sql [][]string
	var select_id string
	var select_sql string
	var result []string
	for {
		line, err := bufTmp.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				err = nil
				break
			} else {
				log.LogIfError(err, "")
			}
		}
		if strings.Contains(line, "==>  Preparing:") {
			result = strings.Split(line, " ")
			// result的第3个元素按“.”分割后的倒数第2个元素为label_id
			select_id = strings.Split(result[2], ".")[len(strings.Split(result[2], "."))-2]
			select_sql = ""
			for i := 6; i < len(result); i++ {
				// 去掉开头的空格
				if len(select_sql) == 0 {
					select_sql = result[i]
				} else {
					select_sql = select_sql + " " + result[i]
				}
			}
			id_sql = append(id_sql, []string{select_id, select_sql})
		}
	}
	return id_sql, err
}

// 将SQL写入文件
func ReplaceAndWriteSQL2File(idSqlList [][]string, pathSql string, db *sql.DB, dbName string) ([]string) {
	newQuery := ""
	labelIdList := []string{}
	for _, i := range idSqlList {
		labelIdList = append(labelIdList, i[0])
	}
	labelIdList = removeDuplicates(labelIdList)
	for _, labelID := range labelIdList {
		sqltext, err := os.Create(pathSql + "/" + labelID + ".sql")
		defer sqltext.Close()
		if err != nil {
			log.LogIfError(err, "")
		}
		// 替换问号
		for _, idSql := range idSqlList {
			if idSql[0] == labelID {
				newQuery, err = replace.ReplaceQuestionMark(idSql[1], db, dbName)
				// 语法错误的不予记录
				if err != nil && !strings.Contains(err.Error(), "syntax error at position") {
					// 在原始sql前加"--"进行注释
					sqltext.WriteString("--" + strings.TrimSpace(idSql[1]) + ";\n")
					// 记录无法进行变量替换的sql
					sqltext.WriteString("--cannot replace: " + strings.TrimSpace(idSql[1]) + ";\n")
				} else if err == nil {
					// 在原始sql前加"--"进行注释
					sqltext.WriteString("--" + strings.TrimSpace(idSql[1]) + ";\n")
					// 记录变量替换后的sql
					sqltext.WriteString("  " + strings.TrimSpace(newQuery) + ";\n")
				}
			}
		}
	}
	log.LogIfInfo("Ending of parsing, you can find the .sql file in dir "+pathSql+".", "")
	log.Log.Flush()
	return labelIdList
}

// 提取 select update delete sql等标签
// 当标签中引用了不在当前xml文件中的<sql></sql>标签时，整个xml将不会被Mybatis识别和解析，相当于啥也没干，但不会报错。
// 当标签引用的<sql></sql>中含有if等其他标签或者含有变量时，该标签会被Mybatis忽略
// 一般情况下，<sql></sql>中配的是文本内容，如：<sql id="Base_Column_List"> ID,MAJOR,BIRTHDAY,AGE,NAME,HOBBY </sql>，这种情况是可以正常解析的
func filter(xmlInput string, xmlOutput string, xmlName string) error {
	mapperFormated, err := os.Open(xmlInput)
	defer mapperFormated.Close()
	if err != nil {
		log.LogIfError(err, "%s")
	}
	mapperFiltered, err := os.Create(xmlOutput)
	defer mapperFiltered.Close()
	if err != nil {
		log.LogIfError(err, "%s")
	}
	mapperFiltered.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	mapperFiltered.WriteString("<!DOCTYPE mapper PUBLIC \"-//mybatis.org//DTD Mapper 3.0//EN\" \"http://mybatis.org/dtd/mybatis-3-mapper.dtd\">\n")
	mapperFiltered.WriteString("<mapper namespace=\"SQLAudit." + xmlName + ".newMapper\">\n")
	mapperFiltered.WriteString("\n")
	buf := bufio.NewReader(mapperFormated)
	flag := 0
	for {
		line, err := buf.ReadString('\n')
		if err != nil && err == io.EOF {
			break
		} else {
			log.LogIfError(err, "")
		}
		if strings.Contains(line, "<select") || strings.Contains(line, "<update") || strings.Contains(line, "<delete") || strings.Contains(line, "<sql") {
			flag = 1
			mapperFiltered.WriteString(line)
		} else if strings.Contains(line, "</select>") || strings.Contains(line, "</update>") || strings.Contains(line, "</delete>") || strings.Contains(line, "</sql>") {
			flag = 0
			mapperFiltered.WriteString(line)
		} else if flag == 1 {
			mapperFiltered.WriteString(line)
		}
	}
	_, err = mapperFiltered.WriteString("</mapper>\n")
	return err
}

// 去重
func removeDuplicates(columnList []string) []string {
	distinctList := []string{""}
	for _, i := range columnList {
		if isNotExist(distinctList, i) {
			distinctList = append(distinctList, i)
		}
	}
	return distinctList[1:]
}

// 校验密码和权限
func VerifyDbPass(db *sql.DB) {
	_, err := db.Query("select 1 from information_schema.columns limit 1")
	if err != nil {
		log.LogIfError(err, "")
	}
}

// 捕获panic的堆栈信息
func PanicTrace(kb int) []byte {
	s := []byte("/src/runtime/panic.go")
	e := []byte("\ngoroutine ")
	line := []byte("\n")
	stack := make([]byte, kb<<10) //4KB
	length := runtime.Stack(stack, true)
	start := bytes.Index(stack, s)
	stack = stack[start:length]
	start = bytes.Index(stack, line) + 1
	stack = stack[start:]
	end := bytes.LastIndex(stack, line)
	if end != -1 {
		stack = stack[:end]
	}
	end = bytes.Index(stack, e)
	if end != -1 {
		stack = stack[:end]
	}
	stack = bytes.TrimRight(stack, "\n")
	return stack
}

////调用soar生成审核报告
//func MapperAudit(testDsn string, username string, password string, labelId []string, pathSql string, pathMybatis string) {
//	dsn4soar := "\"" + username + ":" + password + "@" + testDsn + "\""
//	for _, i := range labelId {
//		cmdStr := "soar -test-dsn=" + dsn4soar + " -allow-online-as-test -query " + pathSql + "/" + i + ".sql -report-type html" + " >" + pathMybatis + "/" + i + "_audit.html"
//		cmd := exec.Command("/bin/bash", "-c", cmdStr)
//		err := cmd.Run()
//		if err != nil {
//			log.LogIfError(err, "")
//		}
//	}
//	fmt.Println("Ending of audit, you can find the report in `" + pathMybatis + "`.")
//	log.LogIfInfo("Ending of audit, you can find the report in `"+pathMybatis+"`.", "")
//}
