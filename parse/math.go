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

// 通过组合，得到动态sql所有可能的情况

/*
定义组合方法start
*/

func combine(columnList []string, fixedList []string) [][]string {
	tmp := make([]string, len(columnList))
	copy(tmp, columnList)
	for _, e := range fixedList {
		tmp = remove(tmp, e)
	}
	var combinedList [][]string
	if len(tmp) == 0 {
		combinedList = append(combinedList, fixedList)
		return combinedList
	}
	for i := 1; i < len(tmp)+1; i++ {
		combinedList = append(combinedList, combinations(tmp, i)...)
	}
	for i := 0; i < len(combinedList); i++ {
		combinedList[i] = append(combinedList[i], fixedList...)
	}
	combinedList = append(combinedList, fixedList)
	return combinedList
}

func combinations(s []string, m int) [][]string {
	n := len(s)
	indexs := zuheResult(n, m)
	result := findStrsByIndexs(s, indexs)
	return result
}

// 组合算法(从nums中取出m个数)
func zuheResult(n int, m int) [][]int {

	// 保存最终结果的数组，总数直接通过数学公式计算
	result := make([][]int, 0, mathZuhe(n, m))
	// 保存每一个组合的索引的数组，1表示选中，0表示未选中
	indexs := make([]int, n)
	for i := 0; i < n; i++ {
		if i < m {
			indexs[i] = 1
		} else {
			indexs[i] = 0
		}
	}

	// 第一个结果
	result = addTo(result, indexs)
	for {
		find := false
		// 每次循环将第一次出现的 1 0 改为 0 1，同时将左侧的1移动到最左侧
		for i := 0; i < n-1; i++ {
			if indexs[i] == 1 && indexs[i+1] == 0 {
				find = true
				indexs[i], indexs[i+1] = 0, 1
				if i > 1 {
					moveOneToLeft(indexs[:i])
				}
				result = addTo(result, indexs)
				break
			}
		}
		// 本次循环没有找到 1 0 ，说明已经取到了最后一种情况
		if !find {
			break
		}
	}
	return result
}

// 将ele复制后添加到arr中，返回新的数组
func addTo(arr [][]int, ele []int) [][]int {
	newEle := make([]int, len(ele))
	copy(newEle, ele)
	arr = append(arr, newEle)
	return arr
}

func moveOneToLeft(leftNums []int) {
	// 计算有几个1
	sum := 0
	for i := 0; i < len(leftNums); i++ {
		if leftNums[i] == 1 {
			sum++
		}
	}
	// 将前sum个改为1，之后的改为0
	for i := 0; i < len(leftNums); i++ {
		if i < sum {
			leftNums[i] = 1
		} else {
			leftNums[i] = 0
		}
	}
}

// 根据索引号数组得到元素数组
func findStrsByIndexs(nums []string, indexs [][]int) [][]string {
	if len(indexs) == 0 {
		return [][]string{}
	}
	result := make([][]string, len(indexs))
	for i, v := range indexs {
		line := make([]string, 0)
		for j, v2 := range v {
			if v2 == 1 {
				line = append(line, nums[j])
			}
		}
		result[i] = line
	}
	return result
}

// 数学方法计算组合数(从n中取m个数)
func mathZuhe(n int, m int) int {
	return jieCheng(n) / (jieCheng(n-m) * jieCheng(m))
}

// 阶乘
func jieCheng(n int) int {
	result := 1
	for i := 2; i <= n; i++ {
		result *= i
	}
	return result
}

/*
end
*/
