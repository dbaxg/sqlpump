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

package replace

import (
	"database/sql"
	"github.com/dbaxg/sqlpump/log"
	"strings"
	"strconv"
)

func ReplaceQuestionMark(query string, db *sql.DB, dbName string) (string, error) {
	//获取SQL语法树
	query = strings.Replace(query, "\n", "", -1)
	stmt, err := getStmt(query)
	if err != nil {
		return query, err
	}

	//获取表信息
	tables := getTables(stmt)
	tableColumnDataType, err := getTableColumnDataType(query, db, dbName, tables)
	if err != nil {
		return query, err
	}

	//获取列信息
	columns := getColumns(stmt)
	//如果不含变量列，则直接返回
	if len(columns) == 0 {
		return query, nil
	}

	for i, column := range columns {
		//收集含有变量列的表
		var tableContainColumn []string
		for k, v := range tableColumnDataType {
			if _, ok := v[column[0]]; ok {
				tableContainColumn = append(tableContainColumn, k)
			}
		}

		//如果多个表有同名列
		if len(tableContainColumn) > 1 {
			var dataType []string
			for _, tab := range tableContainColumn {
				dataType = append(dataType, tableColumnDataType[tab][column[0]])
			}
			//d ataType去重后长度为1，则说明多个表同名列的字段类型相同,取dataType中的任意元素作为列的类型即可
			if len(removeDuplicates(dataType)) == 1 {
				column = append(column, dataType[0])
				// 要将columns中的值换掉， 因为column是局部变量
				columns[i] = column
			} else {
				// 否则无法判断该列的类型，返回未修改的query
				var tableName string
				for i, tab := range tableContainColumn {
					if i != (len(tableContainColumn) - 1) {
						tableName = "`" + tab + "` and "
					} else {
						tableName = tableName + "`" + tab + "`"
					}
				}
				var err error
				err = &errorUserDefined{"query `" + query + "` cannot be replaced, because " +
					"column `" + column[0] + "` exists in table " + tableName + ", and their data type is different, so we cannot figure out the data type of column `" + column[0] + "` correctly."}
				log.LogIfWarn(err, "")
				return query, err
			}
		} else {
			//只有一个表含有该字段，则可以明确该字段的类型
			if len(tableContainColumn) == 1 {
				column = append(column, tableColumnDataType[tableContainColumn[0]][column[0]])
				columns[i] = column
			} else { //该列在不存在，请检查！
				var err error
				err = &errorUserDefined{"query `" + query + "` cannot be replaced, because column `" + column[0] + "` does not exist, please check!"}
				log.LogIfWarn(err, "")
				return query, err
			}
		}
	}

	var i = 0
	// 至此，columns形如[[column position dataType] ...]
	for _, column := range columns {
		// i++是因为每循环一次问号都少一个，所以col中的问号始终都是question_index中的第一个
		i++
		question_index := GetIndex(query, "?")
		dataType := column[2]
		if question_index != nil {
			// idx为变量（问号）在sql中的顺序，从1开始
			idx, _ := strconv.Atoi(column[1])
			// 获取变量（问号）在sql中的index
			index := question_index[idx-i] // idx-i=0
			if strings.Contains(dataType, "char") || strings.Contains(dataType, "text") || strings.Contains(dataType, "lob") || strings.Contains(dataType, "binary") {
				query = query[:index] + "'a'" + query[index+1:]
			} else if strings.Contains(dataType, "json") {
				query = query[:index] + "'{\"name\":\"gaozhongzheng\", \"job\":\"dba\"}'" + query[index+1:]
			} else if strings.Contains(dataType, "int") || strings.Contains(dataType, "bit") {
				query = query[:index] + "1" + query[index+1:]
			} else if strings.Contains(dataType, "decimal") || strings.Contains(dataType, "float") || strings.Contains(dataType, "double") || strings.Contains(dataType, "numeric") {
				query = query[:index] + "1.1" + query[index+1:]
			} else if strings.Contains(dataType, "timestamp") || strings.Contains(dataType, "datetime") {
				query = query[:index] + "'2019-01-01 00:00:00'" + query[index+1:]
			} else if strings.Contains(dataType, "date") {
				query = query[:index] + "'2019-01-01'" + query[index+1:]
			} else if strings.Contains(dataType, "time") {
				query = query[:index] + "'01:01:01'" + query[index+1:]
			} else if strings.Contains(dataType, "year") {
				query = query[:index] + "'2019'" + query[index+1:]
			} else if strings.Contains(dataType, "enum") || strings.Contains(dataType, "set") {
				query = query[:index] + "'a'" + query[index+1:]
			} else {
				query = query[:index] + "none" + query[index+1:]
			}
		}
	}

	// 此时query中只有limit部分的?没有被替换
	stmt, err = getStmt(query)
	if err != nil {
		return query, err
	}

	// 替换limit中的变量
	limits := getLimitInfo(stmt)
	for _, ques := range limits {
		question_index := GetIndex(query, "?")
		if ques[0] != -1 {
			index := question_index[ques[0]-1]
			query = query[:index] + "1" + query[index+1:]
		}
		if ques[1] != -1 {
			index := question_index[ques[1]-1]
			query = query[:index] + "1" + query[index+1:]
		}
	}
	return query, nil
}

func removeDuplicates(s []string) []string {
	result := []string{}
	tempMap := map[string]byte{} // 存放不重复主键
	for _, e := range s {
		l := len(tempMap)
		tempMap[e] = 0
		if len(tempMap) != l { // 加入map后，map长度变化，则元素不重复
			result = append(result, e)
		}
	}
	return result
}

// 获取变量（问号）在sql中的index
func GetIndex(str string, subStr string) []int {
	var result []int
	length := len(subStr)
	tmpStr := ""
	for i := 0; i < length; i++ {
		tmpStr = tmpStr + "1"
	}
	for {
		if !strings.Contains(str, subStr) {
			break
		}
		index := strings.Index(str, subStr)
		result = append(result, index)
		str = str[:index] + tmpStr + str[index+1:]
	}
	return result
}

// TODO 获取column_type中的类型和值范围来赋值，这样更精确些（注意，不是data_type）。

//func getRandomValue(dataType string) string {
//	if strings.Contains(dataType, "char") || strings.Contains(dataType, "text") || strings.Contains(dataType, "lob") {
//		return "'" + RandString(10) + "'"
//	}
//	if strings.Contains(dataType, "int") {
//		return strconv.Itoa(rand.Intn(1000))
//	}
//	if strings.Contains(dataType, "decimal") || strings.Contains(dataType, "float") || strings.Contains(dataType, "double") {
//		return strconv.FormatFloat(float64(rand.Intn(10000))/100, 'f', 2, 32)
//	}
//	if strings.Contains(dataType, "timestamp") || strings.Contains(dataType, "datetime") {
//		return "'" + time.Now().Format("2006-01-02 15:04:08") + "'"
//	}
//	if strings.Contains(dataType, "date") {
//		return "'" + time.Now().Format("2006-01-02") + "'"
//	}
//	if strings.Contains(dataType, "time") {
//		return "'" + time.Now().Format("15:04:08") + "'"
//	}
//	if strings.Contains(dataType, "year") {
//		return "'" + time.Now().Format("2006") + "'"
//	}
//	if strings.Contains(dataType, "enum") || strings.Contains(dataType, "set") {
//		//enum和set类型时，dataType形如enum('M','F')和set('M','F')，取第一个值即可
//		firstIdx := strings.Index(dataType, "'")
//		secondIdx := strings.Index(dataType, ",")
//		return dataType[firstIdx:secondIdx]
//	}
//	if strings.Contains(dataType, "binary") {
//		return "'a'"
//	}
//	return ""
//}

//func RandString(n int) string {
//	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
//
//	var src = rand.NewSource(time.Now().UnixNano())
//
//	const (
//		letterIdxBits = 6
//		letterIdxMask = 1<<letterIdxBits - 1
//		letterIdxMax  = 63 / letterIdxBits
//	)
//	b := make([]byte, n)
//	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
//		if remain == 0 {
//			cache, remain = src.Int63(), letterIdxMax
//		}
//		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
//			b[i] = letterBytes[idx]
//			i--
//		}
//		cache >>= letterIdxBits
//		remain--
//	}
//	return string(b)
//}

/*
MySQL支持的类型如下：
mysql> select distinct data_type from columns order by 1;
+------------+
| data_type  |
+------------+
| bigint     |
| bit        |
| blob       |
| char       |
| date       |
| datetime   |
| decimal    |
| double     |
| enum       |
| float      |
| int        |
| json       |
| longblob   |
| longtext   |
| mediumint  |
| mediumtext |
| set        |
| smallint   |
| text       |
| time       |
| timestamp  |
| tinyint    |
| varchar    |
| year       |
+------------+
23 rows in set (0.08 sec)
 */
