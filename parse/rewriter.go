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
	"strings"
	"strconv"
)

func isNotExist(slice []string, str string) bool {
	for _, i := range slice {
		if i == str {
			return false
		}
	}
	return true
}

func getSelect(line string) []string {
	var s []int
	for i, v := range line {
		if string(v) == "\"" {
			s = append(s, i)
		}
	}
	idx1 := s[0] + 1
	idx2 := s[1]
	selectId := line[idx1:idx2]
	selectLabel := "<select id=\"" + selectId + "\">"
	var selectInfo []string
	selectInfo = append(selectInfo, selectId)
	selectInfo = append(selectInfo, selectLabel)
	return selectInfo
}

func getUpdate(line string) []string {
	var s []int
	for i, v := range line {
		if string(v) == "\"" {
			s = append(s, i)
		}
	}
	idx1 := s[0] + 1
	idx2 := s[1]
	updateId := line[idx1:idx2]
	updateLabel := "<update id=\"" + updateId + "\">"
	var updateInfo []string
	updateInfo = append(updateInfo, updateId)
	updateInfo = append(updateInfo, updateLabel)
	return updateInfo
}

func getDelete(line string) []string {
	var s []int
	for i, v := range line {
		if string(v) == "\"" {
			s = append(s, i)
		}
	}
	idx1 := s[0] + 1
	idx2 := s[1]
	deleteId := line[idx1:idx2]
	deleteLabel := "<delete id=\"" + deleteId + "\">"
	var deleteInfo []string
	deleteInfo = append(deleteInfo, deleteId)
	deleteInfo = append(deleteInfo, deleteLabel)
	return deleteInfo
}

func getIf(columnName string, line string) string {
	idx1 := strings.Index(line, "\"")
	tmp := line[0:idx1+1] + columnName + " != null\">\n"
	return tmp
}

func getWhen(columnName string, line string) string {
	idx1 := strings.Index(line, "\"")
	tmp := line[0:idx1+1] + columnName + " != null\">\n"
	return tmp
}

func getForeach(num int, selectInfo []string, line string) (string, string) {
	result := strings.Split(line, " ")
	// 获取“<”前面的空格并赋值给tmp，为了美观
	idx1 := strings.Index(line, "<")
	tmp := line[0:idx1]
	collection_column := ""
	for _, j := range result {
		if strings.Contains(j, "collection") {
			j = "collection=\"collection" + strconv.Itoa(num) + "\""
			collection_column = "collection" + strconv.Itoa(num)
		} else if strings.Contains(j, "item") {
			// 将item换成select_id+str(num),使其与foreach标签对中的变量相同
			j = "item=\"" + string(selectInfo[0]) + strconv.Itoa(num) + "\""
		}
		tmp = tmp + j + " "
	}
	return collection_column, tmp
}

func getSqlDynamic(columnName string, line string) string {
	// 获取紧跟“#”号后面的“{”的索引idx2
	idx1 := strings.Index(line, "#")
	idx2 := 0
	idx3 := 0
	var s1 []int
	for i, v := range line {
		//if string(v) == "\"" {
		if string(v) == "{" {
			s1 = append(s1, i)
		}
	}
	if len(s1) > 1 {
		for _, i := range s1 {
			if i > idx1 {
				idx2 = i
				break
			}
		}
	} else {
		idx2 = strings.Index(line, "{")
	}

	// 获取line中与上述“{”匹配的“}”的索引idx3
	var s2 []int
	for i, v := range line {
		//if string(v) == "\"" {
		if string(v) == "}" {
			s2 = append(s2, i)
		}
	}
	if len(s2) > 1 {
		for _, i := range s2 {
			if i > idx2 {
				idx3 = i
				break
			}
		}
	} else {
		idx3 = strings.Index(line, "}")
	}
	tmp := line[0:idx2+1] + columnName + line[idx3:]
	return tmp
}

func handleCollection(columnList []string) []string {
	var columnListFinal []string
	num := "0"
	numLenReverse := 0
	for _, t := range columnList {
		if strings.Contains(t, "collection") {
			num = t[10:]
			numLenReverse = -len(num)
			columnListFinal = append(columnListFinal, t)
		} else if t[len(t)+numLenReverse:] == num {
			continue
		} else {
			columnListFinal = append(columnListFinal, t)
		}
	}
	return columnListFinal
}

func removeAll(s []string, a string) []string {
	var result []string
	for _, v := range s {
		if v == a {
			continue
		} else {
			result = append(result, v)
		}
	}
	return result
}

func remove(s []string, a string) []string {
	for i, v := range s {
		if v == a {
			s = append(s[:i], s[i+1:]...)
			return s
		}
	}
	return s
}

func getFixedList(columnList []string) []string {
	tmp := make([]string, len(columnList))
	copy(tmp, columnList)
	// 实际用于组合的动态字段会被记录两次
	tmp1 := removeDuplicates(tmp)
	// fixed_list为固定列，即collection列和非null列
	fixedList := make([]string, len(columnList))
	copy(fixedList, columnList)

	for _, i := range tmp1 {
		tmp = remove(tmp, i)
	}

	if len(tmp) > 0 {
		for _, m := range tmp {
			fixedList = removeAll(fixedList, m)
		}
	}
	return fixedList
}
