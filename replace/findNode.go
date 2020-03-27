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

import "vitess.io/vitess/go/vt/sqlparser"

func FindAllConditions(node sqlparser.SQLNode) ([]interface{}, error) {
	var conditions []interface{}
	err := sqlparser.Walk(func(node sqlparser.SQLNode) (kontinue bool, err error) {
		switch node := node.(type) {
		case *sqlparser.ComparisonExpr, *sqlparser.RangeCond, *sqlparser.IsExpr, *sqlparser.UpdateExpr:
			conditions = append(conditions, node)
		}
		return true, nil
	}, node)
	return conditions, err
}

func FindAllTableNodes(node sqlparser.SQLNode) ([]interface{}, error) {
	var tableNodes []interface{}
	err := sqlparser.Walk(func(node sqlparser.SQLNode) (kontinue bool, err error) {
		switch node := node.(type) {
		case *sqlparser.AliasedTableExpr:
			tableNodes = append(tableNodes, node)
		}
		return true, nil
	}, node)
	// fmt.Println(err)
	return tableNodes, err
}

func FindLimitNodes(node sqlparser.SQLNode) ([]interface{}, error) {
	var limitNodes []interface{}
	err := sqlparser.Walk(func(node sqlparser.SQLNode) (kontinue bool, err error) {
		switch node := node.(type) {
		case *sqlparser.Limit:
			if node != nil {
				limitNodes = append(limitNodes, node)
			}
		}
		return true, nil
	}, node)
	return limitNodes, err
}


