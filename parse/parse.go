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

package parse

import (
	"bufio"
	"io"
	"os"
	"strconv"
	"strings"
	"github.com/dbaxg/sqlpump/log"
)

func ParseMapper(filename string, pathMybatis string, pathLib string, testDb string, username string, password string) [][]string {
	log.LogIfInfo("Starting to parse mapper file.", "")

	//xmlName为文件名（不含`.xml`），后续将用于创建Mybatis project
	var xmlName string
	if strings.Contains(filename, "/") {
		xmlName = filename[strings.LastIndex(filename, "/")+1:]
		xmlName = xmlName[:strings.LastIndex(xmlName, ".")]
	} else {
		xmlName = filename[:strings.LastIndex(filename, ".")]
	}

	// 创建配置文件
	log.LogIfInfo("Starting to create configuration file for mybatis project.", "")
	err := createConf(pathMybatis, xmlName, testDb, username, password)
	if err == nil {
		log.LogIfInfo("Configuration file were created sucessfully.", "")
	} else {
		log.LogIfError(err, "")
	}

	// 创建log4j.properties
	log.LogIfInfo("Starting to create log4j.properties for mybatis project.", "")
	err = createLog4jProperties(pathMybatis, xmlName)
	if err == nil {
		log.LogIfInfo("log4j.properties was created sucessfully.", "")
	} else {
		log.LogIfError(err, "")
	}

	// 创建DBTools.java
	log.LogIfInfo("Starting to create DBtools.java for mybatis project.", "")
	packageName := "package SQLAudit." + xmlName + ";"
	err = createDBTools(pathMybatis, packageName)
	if err == nil {
		log.LogIfInfo("DBtools.java was created sucessfully.", "")
	} else {
		log.LogIfError(err, "")
	}

	//格式化xml文件
	log.LogIfInfo("Starting to create the mapperFormatted.xml.", "")
	xmlInput := filename
	xmlOutput := pathMybatis + "/mapperFormatted.xml"
	err = format(xmlInput, xmlOutput, xmlName)
	if err == nil {
		log.LogIfInfo("mapperFormatted.xml was created successfully.", "")
	} else {
		log.LogIfError(err, "")
	}

	//提取select update delete sql等标签
	log.LogIfInfo("Starting to create the mapperFiltered.xml.", "")
	xmlInput = xmlOutput
	xmlOutput = pathMybatis + "/mapperFiltered.xml"
	err = filter(xmlInput, xmlOutput, xmlName)
	if err == nil {
		log.LogIfInfo("mapperFiltered.xml was created successfully.", "")
	} else {
		log.LogIfError(err, "")
	}

	// 根据mapperFitered构件新的xml文件
	log.LogIfInfo("Starting to create newMapper.xml.", "")
	mapperFitered, err := os.Open(xmlOutput)
	defer mapperFitered.Close()
	if err != nil {
		log.LogIfError(err, "")
	}
	bufMapper := bufio.NewReader(mapperFitered)
	newMapper, err := os.Create(pathMybatis + "/newMapper.xml")
	if err != nil {
		log.LogIfError(err, "")
	}
	defer newMapper.Close()

	// 定义变量
	var idList []string
	var columnList []string
	var labelInfo []string
	var columnName string
	var ifNullSql string
	var ifNotNullSql string
	var whenNullSql string
	var whenNotNullSql string
	var sqlDynamic string
	var fixedList []string
	var pathBean string
	var idSqlList [][]string
	var i int
	var flag = 0

	// 开始构建newMapper.xml
	for {
		line, err := bufMapper.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				log.LogIfInfo("newMapper.xml was created successfully.", "")
				break
			} else {
				log.LogIfError(err, "%s", "something wrong happened,exit!")
			}
		}
		if len(strings.TrimSpace(line)) == 0 {
			continue
		} else if strings.Contains(line, "<select") {
			flag = 1
			i = 0
			columnList = []string{}
			labelInfo = getSelect(line)
			newMapper.WriteString("\n\n" + labelInfo[1] + "\n")
		} else if strings.Contains(line, "<update") {
			flag = 1
			i = 0
			columnList = []string{}
			labelInfo = getUpdate(line)
			newMapper.WriteString("\n\n" + labelInfo[1] + "\n")
		} else if strings.Contains(line, "<delete") {
			flag = 1
			i = 0
			columnList = []string{}
			labelInfo = getDelete(line)
			newMapper.WriteString("\n\n" + labelInfo[1] + "\n")
		} else if strings.Contains(line, "<if") && flag == 1 {
			if strings.Contains(line, "null") {
				columnName = labelInfo[0] + strconv.Itoa(i)
				columnList = append(columnList, columnName)
				ifNullSql = getIf(columnName, line)
				newMapper.WriteString(ifNullSql)
			} else {
				columnName = labelInfo[0] + strconv.Itoa(i)
				// 记录两次，把该列当做动态字段（动态字段会被记录两次）
				columnList = append(columnList, columnName)
				columnList = append(columnList, columnName)
				ifNotNullSql = getIf(columnName, line)
				newMapper.WriteString(ifNotNullSql)
				i++
			}
		} else if strings.Contains(line, "<when") && flag == 1 {
			if strings.Contains(line, "null") {
				columnName = labelInfo[0] + strconv.Itoa(i)
				columnList = append(columnList, columnName)
				whenNullSql = getWhen(columnName, line)
				newMapper.WriteString(whenNullSql)
			} else {
				columnName = labelInfo[0] + strconv.Itoa(i)
				// 记录两次，把该列当做动态字段（动态字段会被记录两次）
				columnList = append(columnList, columnName)
				columnList = append(columnList, columnName)
				whenNotNullSql = getIf(columnName, line)
				newMapper.WriteString(whenNotNullSql)
				i++
			}
		} else if strings.Contains(line, "<foreach") && flag == 1 {
			foreachColumn, foreachLine := getForeach(i, labelInfo, line)
			columnList = append(columnList, foreachColumn)
			newMapper.WriteString(foreachLine)
		} else if strings.Contains(line, "#") && flag == 1 {
			columnName = labelInfo[0] + strconv.Itoa(i)
			columnList = append(columnList, columnName)
			sqlDynamic = getSqlDynamic(columnName, line)
			newMapper.WriteString(sqlDynamic)
			i++
		} else if strings.Contains(line, "</select") || strings.Contains(line, "</update") || strings.Contains(line, "</delete") {
			flag = 0
			newMapper.WriteString(line + "\n\n")
			idList = append(idList, labelInfo[0])
			fixedList = getFixedList(columnList)
			columnList = removeDuplicates(columnList)

			//进一步处理collection的情况
			for _, s := range columnList {
				if strings.Contains(s, "collection") {
					columnList = handleCollection(columnList)
					break
				}
			}

			for _, v := range fixedList {
				if strings.Contains(v, "collection") {
					fixedList = handleCollection(fixedList)
					break
				}
			}

			//只对小于等于8个动态变量的sql进行组合，因为动态变量超过8个时，组合情况太多会导致负载太高
			n := len(columnList) - len(fixedList)
			if n <= 8 {
				combinedList := combine(columnList, fixedList)
				//创建java bean
				log.LogIfInfo("Starting to create java bean for mybatis project.", "")
				pathBean = pathMybatis + "/" + labelInfo[0] + ".java"
				err = createBean(labelInfo, pathBean, columnList, packageName)
				if err == nil {
					log.LogIfInfo("Java bean were created successfully.", "")
				} else {
					log.LogIfError(err, "")
				}

				//创建测试程序
				log.LogIfInfo("Starting to create test processe for mybatis project.", "")
				err = createTestProcess(labelInfo[0], combinedList, pathMybatis, packageName)
				if err == nil {
					log.LogIfInfo("Test processe were created successfully.", "")
				} else {
					log.LogIfError(err, "")
				}
			} else {
				//动态字段超过8个时，仅取最完整的一种组合
				var combinedList [][]string
				combinedList = append(combinedList, columnList)
				log.LogIfInfo("Starting to create java bean for mybatis project.", "")
				pathBean = pathMybatis + "/" + labelInfo[0] + ".java"
				err = createBean(labelInfo, pathBean, columnList, packageName)
				if err == nil {
					log.LogIfInfo("Java bean were created successfully.", "")
				} else {
					log.LogIfError(err, "")
				}

				//创建测试程序
				log.LogIfInfo("Starting to create test processe for mybatis project.", "")
				err = createTestProcess(labelInfo[0], combinedList, pathMybatis, packageName)
				if err == nil {
					log.LogIfInfo("Test processe were created successfully.", "")
				} else {
					log.LogIfError(err, "")
				}
			}
		} else {
			newMapper.WriteString(line)
		}
	}

	// 创建接口文件
	log.LogIfInfo("Starting to create interface file for mybatis project.", "")
	err = createInterface(idList, pathMybatis, packageName)
	if err == nil {
		log.LogIfInfo("Interface file were created successfully.", "")
	} else {
		log.LogIfError(err, "%s")
	}

	// 创建Sh文件
	log.LogIfInfo("Starting to create shell script which will be used to compile and execute mybatis project.", "")
	err = createSh(pathMybatis, pathLib, idList, xmlName)
	if err == nil {
		log.LogIfInfo("Shell script was created sucessfully.", "")
	} else {
		log.LogIfError(err, "")
	}

	// 执行shell调起Mybatis工程
	log.LogIfInfo("Starting to execute mybatis project to generate sql log.", "")
	err = execMybatisProject(pathMybatis, xmlName)
	if err == nil {
		log.LogIfInfo("Mybatis project was executed sucessfully.", "")
	} else {
		log.LogIfError(err, "")
	}

	// 解析Mybatis日志获取select_id和sql文本
	log.LogIfInfo("Starting to parse mybatis sql log to get sql.", "")
	logPath := pathMybatis + "/" + xmlName + ".log"
	idSqlList, err = getSQL(logPath)
	if err == nil {
		log.LogIfInfo("Mybatis sql log was parsed sucessfully.", "")
	} else {
		log.LogIfError(err, "")
	}
	return idSqlList
}
