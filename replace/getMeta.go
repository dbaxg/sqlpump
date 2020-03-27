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

package replace

import (
	"vitess.io/vitess/go/vt/sqlparser"
	"github.com/dbaxg/sqlpump/log"
	"database/sql"
	"strconv"
	"fmt"
)

func getStmt(query string) (sqlparser.SQLNode, error) {
	stmt, err := sqlparser.Parse(query)
	if err != nil {
		var errParse error
		errParse = &errorUserDefined{"parsing `" + query + "` failed: " + err.Error() + "."}
		log.LogIfWarn(errParse, "")
		return stmt, errParse
	}
	return stmt, nil
}

func getTables(stmt sqlparser.SQLNode) []string {
	var table []string
	tableNodes, _ := FindAllTableNodes(stmt)
	for _, tableNode := range tableNodes {
		switch node := tableNode.(type) {
		case *sqlparser.AliasedTableExpr:
			switch tableName := node.Expr.(type) {
			case sqlparser.TableName:
				table = append(table, tableName.Name.String())
			}
		}
	}
	return table
}

func getColumns(stmt sqlparser.SQLNode) [][]string {
	var columns [][]string
	conditions, _ := FindAllConditions(stmt)
	for _, condition := range conditions {
		switch node := condition.(type) {
		case *sqlparser.ComparisonExpr:
			// 获取 condition 左侧的信息
			switch nLeft := node.Left.(type) {
			case *sqlparser.SQLVal, sqlparser.ValTuple, *sqlparser.FuncExpr:
				err := sqlparser.Walk(func(nLeft sqlparser.SQLNode) (kontinue bool, err error) {
					switch vals := nLeft.(type) {
					case sqlparser.ValTuple:
						for _, val := range vals {
							switch vleft := val.(type) {
							case *sqlparser.SQLVal:
								//如果等号左边含有变量（？），则需提取右边的ColName
								if vleft.Type == 5 && vleft.Val[0] == 0x3a && vleft.Val[1] == 0x76 {
									switch col := node.Right.(type) {
									case *sqlparser.ColName:
										var column []string
										column = append(column, col.Name.String())
										//val.Val[2]是变量（？）在sql中出现的位置，从1开始
										column = append(column, string(vleft.Val[2]))
										////将列对应的表名/别名放入column
										//column = append(column, col.Qualifier.Name.String())
										//将变量对应列、变量位置放入columns
										columns = append(columns, column)
									}
								}

							}
						}

					case *sqlparser.SQLVal:
						//如果等号左边含有变量（？），则需提取右边的ColName
						if vals.Type == 5 && vals.Val[0] == 0x3a && vals.Val[1] == 0x76 {
							switch col := node.Right.(type) {
							case *sqlparser.ColName:
								var column []string
								column = append(column, col.Name.String())
								//val.Val[2]是变量（？）在sql中出现的位置，从1开始
								column = append(column, string(vals.Val[2]))
								////将列对应的表名/别名放入column
								//column = append(column, col.Qualifier.Name.String())
								//将变量对应列、变量位置放入columns
								columns = append(columns, column)
							}
						}
					}
					return true, nil
				}, nLeft)
				if err != nil {
					log.LogIfWarn(err, "")
					continue
				}
			}

			// 获取 condition 右侧的信息
			switch nRight := node.Right.(type) {
			case *sqlparser.SQLVal, sqlparser.ValTuple, *sqlparser.FuncExpr:
				err := sqlparser.Walk(func(nRight sqlparser.SQLNode) (kontinue bool, err error) {
					switch vals := nRight.(type) {
					case sqlparser.ValTuple:
						//fmt.Println("comming here 1")
						for _, val := range vals {
							switch vright := val.(type) {
							case *sqlparser.SQLVal:
								//fmt.Println("coming here 2")
								//如果等号右边含有变量（？），则需提取左边的ColName
								if vright.Type == 5 && vright.Val[0] == 0x3a && vright.Val[1] == 0x76 {
									switch col := node.Right.(type) {
									case *sqlparser.ColName:
										var column []string
										column = append(column, col.Name.String())
										//val.Val[2]是变量（？）在sql中出现的位置，从1开始
										column = append(column, string(vright.Val[2]))
										////将列对应的表名/别名放入column
										//column = append(column, col.Qualifier.Name.String())
										//将变量对应列、变量位置放入columns
										columns = append(columns, column)
									}
								}

							}
						}
					case *sqlparser.SQLVal:
						//fmt.Println("coming here 3")
						//如果等号右边含有变量（？），则需提取左边的ColName
						if vals.Type == 5 && vals.Val[0] == 0x3a && vals.Val[1] == 0x76 {
							switch col := node.Left.(type) {
							case *sqlparser.ColName:
								var column []string
								column = append(column, col.Name.String())
								//val.Val[2]是变量（？）在sql中出现的位置，从1开始
								column = append(column, string(vals.Val[2]))
								//将列对应的表名/别名放入column
								//column = append(column, col.Qualifier.Name.String())
								//将变量对应列、变量位置、对应表名放入columns
								columns = append(columns, column)
							}
						}
					}
					return true, nil
				}, nRight)
				if err != nil {
					log.LogIfWarn(err, "")
					continue
				}
			}
			// TODO 考虑改成walk的方式？
		case *sqlparser.RangeCond:
			switch fNode := node.From.(type) {
			case *sqlparser.SQLVal:
				if fNode.Type == 5 && fNode.Val[0] == 0x3a && fNode.Val[1] == 0x76 {
					switch col := node.Left.(type) {
					case *sqlparser.ColName:
						var column []string
						column = append(column, col.Name.String())
						column = append(column, string(fNode.Val[2]))
						////将列对应的表名/别名放入column
						//column = append(column, col.Qualifier.Name.String())
						columns = append(columns, column)
					}
				}
			}
			switch tNode := node.To.(type) {
			case *sqlparser.SQLVal:
				if tNode.Type == 5 && tNode.Val[0] == 0x3a && tNode.Val[1] == 0x76 {
					switch col := node.Left.(type) {
					case *sqlparser.ColName:
						var column []string
						column = append(column, col.Name.String())
						column = append(column, string(tNode.Val[2]))
						//将列对应的表名/别名放入column
						//column = append(column, col.Qualifier.Name.String())
						columns = append(columns, column)
					}
				}
			}
		case *sqlparser.IsExpr:
			// TODO 一般不需要处理这种情况，因为is后面只能是null，如果是?，则导致解析失败
		case *sqlparser.UpdateExpr:
			switch updateExpr := node.Expr.(type) {
			case *sqlparser.SQLVal:
				if updateExpr.Type==5 && updateExpr.Val[0] == 0x3a && updateExpr.Val[1] == 0x76{
						//update xxx set xx=?, xx只能是type *sqlparser.ColName
						//所以node.Name确定是type *sqlparser.ColName，此处无法使用type switch:
						//cannot type switch on non-interface value node.Name (type *sqlparser.ColName)
						var column []string
						column = append(column, node.Name.Name.String())
						//val.Val[2]是变量（？）在sql中出现的位置，从1开始
						column = append(column, string(updateExpr.Val[2]))
						//将变量对应列、变量位置放入columns
						columns = append(columns, column)
				}
			}
		}
	}
	return columns
}

//[tab1:[col1:dtype1 col2:dtype2 ...] ...]
func getTableColumnDataType(query string, db *sql.DB, dbName string, tables []string) (map[string]map[string]string, error) {
	tableColumnDataType := make(map[string]map[string]string)
	var column string
	var dataType string
	for _, tab := range tables {
		//查找表中的字段和字段类型
		//初始化子map
		columnDataType := make(map[string]string)
		//从information_schema.columns中
		rows, err := db.Query("select column_name, data_type from information_schema.columns where table_name = ? and TABLE_SCHEMA = '"+dbName+"'", tab)
		defer rows.Close()
		if err != nil {
			fmt.Println(err.Error())
			log.LogIfError(err,"")
		}

		for rows.Next() {
			if err := rows.Scan(&column, &dataType); err == nil {
				if _, ok := columnDataType[column]; !ok {
					columnDataType[column] = dataType
				}
			}
		}
		if len(columnDataType) == 0 { //说明表在库中不存在
			var err error
			err = &errorUserDefined{"`" + query + "`: " + "table `" + tab + "` doesn't exist in database `" + dbName + "`, please check!"}
			log.LogIfWarn(err, "")
			return tableColumnDataType, err
		}
		tableColumnDataType[tab] = columnDataType
	}
	return tableColumnDataType, nil
}

func getLimitInfo(stmt sqlparser.SQLNode) [][]int {
	var limits [][]int
	limitNodes, _ := FindLimitNodes(stmt)
	for _, lim := range limitNodes {
		switch node := lim.(type) {
		case *sqlparser.Limit:
			var limit []int
			if node != nil {
				if node.Offset == nil {
					limit = append(limit, -1)
				} else {
					switch offVal := node.Offset.(type) {
					case *sqlparser.SQLVal:
						if offVal.Type == 5 && offVal.Val[0] == 0x3a && offVal.Val[1] == 0x76 {
							position, _ := strconv.Atoi(string(offVal.Val[2]))
							limit = append(limit, position)
						} else {
							limit = append(limit, -1)
						}
					}
				}

				if node.Rowcount == nil {
					limit = append(limit, -1)
				} else {
					switch rowVal := node.Rowcount.(type) {
					case *sqlparser.SQLVal:
						if rowVal.Type == 5 && rowVal.Val[0] == 0x3a && rowVal.Val[1] == 0x76 {
							position, _ := strconv.Atoi(string(rowVal.Val[2]))
							limit = append(limit, position)
						} else {
							limit = append(limit, -1)
						}
					}
				}
			} else {
				limit = append(limit, -1)
				limit = append(limit, -1)
			}
			//将每个limit中的offeset和rowcount的位置信息放入limits列表
			limits = append(limits, limit)
		}
	}
	return limits
}
