package vector

import (
	"reflect"
)

// 小写 只能通过工厂函数创建
type Vector struct {
	values []interface{}
}

// 创建工厂函数
func New() *Vector {
	vector := new(Vector)
	vector.values = make([]interface{}, 0)

	return vector
}

func (vec *Vector) IsEmpty() bool {
	return len(vec.values) == 0
}

// 元素数量
func (vector *Vector) Size() int {
	return len(vector.values)
}

// 追加单个元素
func (vector *Vector) Append(value interface{}) bool {
	vector.values = append(vector.values, value)
	return true
}

// 追加元素切片
func (vector *Vector) AppendAll(values []interface{}) bool {
	if values == nil || len(values) < 1 {
		return false
	}
	vector.values = append(vector.values, values...)
	return true
}

// 插入单个元素
func (vector *Vector) Insert(index int, value interface{}) bool {
	if index < 0 || index >= len(vector.values) {
		return false
	}
	vector.values = append(vector.values[:index], append([]interface{}{value}, vector.values[index:]...)...)
	return true
}

// 插入元素切片
func (vector *Vector) InsertAll(index int, values []interface{}) bool {
	if index < 0 || index >= len(vector.values) || values == nil || len(values) < 1 {
		return false
	}
	vector.values = append(vector.values[:index], append(values, vector.values[index:]...)...)
	return true
}

// 移除
func (vector *Vector) Remove(index int) bool {
	if index < 0 || index >= len(vector.values) {
		return false
	}
	// 重置为 nil 防止内存泄漏
	vector.values[index] = nil
	vector.values = append(vector.values[:index], vector.values[index+1:]...)
	return true
}

// 范围移除 从 fromIndex(包含) 到 toIndex(不包含) 之间的元素
func (vector *Vector) RemoveRange(fromIndex, toIndex int) bool {
	if fromIndex < 0 || fromIndex >= len(vector.values) || toIndex > len(vector.values) || fromIndex > toIndex {
		return false
	}
	// 重置为 nil 防止内存泄漏
	for i := fromIndex; i < toIndex; i++ {
		vector.values[i] = nil
	}
	vector.values = append(vector.values[:fromIndex], vector.values[toIndex:]...)
	return true
}

// 全部移除
func (vector *Vector) RemoveAll() {
	// 重置为 nil 防止内存泄漏
	for i := 0; i < vector.Size(); i++ {
		vector.values[i] = nil
	}
	vector.values = vector.values[0:0]
}

func (vector *Vector) getIndex(value interface{}) int {
	for i := 0; i < len(vector.values); i++ {
		if reflect.DeepEqual(vector.values[i], value) {
			return i
		}
	}
	return -1
}

// 是否存在该元素值
func (vector *Vector) Contains(value interface{}) bool {
	return vector.getIndex(value) >= 0
}

// 获取元素值第一次出现的索引
func (vector *Vector) IndexOf(value interface{}) int {
	return vector.getIndex(value)
}

// 获取元素值最后一次出现的索引
func (vector *Vector) LastIndexOf(value interface{}) int {
	for i := len(vector.values) - 1; i >= 0; i-- {
		if reflect.DeepEqual(vector.values[i], value) {
			return i
		}
	}
	return -1
}

// 得到索引对应的元素值
func (vector *Vector) GetValue(index int) interface{} {
	if index < 0 || index >= len(vector.values) {
		return nil
	}
	return vector.values[index]
}

// 设置值
func (vector *Vector) SetValue(index int, value interface{}) bool {
	if index < 0 || index >= len(vector.values) {
		return false
	}
	vector.values[index] = value
	return true
}

func (vector *Vector) ToArray() []interface{} {
	dst := make([]interface{}, vector.Size())
	copy(dst, vector.values)
	return dst
}
