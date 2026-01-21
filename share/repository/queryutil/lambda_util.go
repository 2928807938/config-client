package queryutil

import (
	"reflect"
	"strings"
	"sync"
)

// EntityFields 实体字段查询构建器（通用类型）
type EntityFields[T any] struct {
	fields        map[string]*FieldQuery // 字段名 -> FieldQuery
	offsetToField map[uintptr]string     // 字段偏移量 -> 字段名
	mu            sync.RWMutex
}

// Lambda 创建实体字段查询构建器（自动反射解析字段）
func Lambda[T any]() *EntityFields[T] {
	ef := &EntityFields[T]{
		fields:        make(map[string]*FieldQuery),
		offsetToField: make(map[uintptr]string),
	}
	ef.initialize()
	return ef
}

// initialize 初始化字段映射（自动解析 struct tag）
func (ef *EntityFields[T]) initialize() {
	var zero T
	t := reflect.TypeOf(zero)

	// 如果是指针类型，获取其元素类型
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return
	}

	// 遍历所有字段
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// 跳过匿名字段和未导出字段
		if field.Anonymous || !field.IsExported() {
			continue
		}

		// 从 gorm tag 中获取列名
		columnName := ef.getColumnName(field)
		if columnName == "" {
			// 如果没有 gorm tag，使用字段名的 snake_case 形式
			columnName = toSnakeCase(field.Name)
		}

		// 创建字段查询构建器
		ef.fields[field.Name] = Field(columnName)

		// 存储字段偏移量映射
		ef.offsetToField[field.Offset] = field.Name
	}
}

// getColumnName 从 struct tag 中获取列名
func (ef *EntityFields[T]) getColumnName(field reflect.StructField) string {
	gormTag := field.Tag.Get("gorm")
	if gormTag == "" {
		return ""
	}

	// 解析 gorm tag，查找 column: 部分
	tags := strings.Split(gormTag, ";")
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if strings.HasPrefix(tag, "column:") {
			return strings.TrimPrefix(tag, "column:")
		}
	}

	return ""
}

// Of 通过字段指针获取 FieldQuery（完全类型安全，无魔法字符串）
// 使用示例：
//
//	var model infraEntity.ConfigPO
//	fields := queryutil.Lambda[infraEntity.ConfigPO]()
//	condition := fields.Of(&model.NamespaceID).Eq(1)
func (ef *EntityFields[T]) Of(fieldPtr interface{}) *FieldQuery {
	// 获取字段指针
	fieldValue := reflect.ValueOf(fieldPtr)
	if fieldValue.Kind() != reflect.Ptr {
		panic("Of() requires a pointer to a field")
	}

	// 计算字段相对于结构体起始位置的偏移量
	fieldAddr := fieldValue.Pointer()

	// 创建一个零值实例来获取结构体基址
	var zero T
	zeroValue := reflect.ValueOf(&zero)
	if zeroValue.Kind() == reflect.Ptr {
		zeroValue = zeroValue.Elem()
	}

	structAddr := zeroValue.Addr().Pointer()

	// 计算偏移量
	offset := uintptr(fieldAddr) - uintptr(structAddr)

	ef.mu.RLock()
	defer ef.mu.RUnlock()

	// 通过偏移量查找字段名
	if fieldName, ok := ef.offsetToField[offset]; ok {
		if fq, ok := ef.fields[fieldName]; ok {
			return fq
		}
	}

	panic("field not found in entity")
}

// Get 通过字段名获取 FieldQuery(类型安全的字符串查询)
// 使用示例:
//
//	fields := queryutil.Lambda[infraEntity.ConfigPO]()
//	condition := fields.Get("NamespaceID").Eq(1)
func (ef *EntityFields[T]) Get(fieldName string) *FieldQuery {
	ef.mu.RLock()
	defer ef.mu.RUnlock()

	if fq, ok := ef.fields[fieldName]; ok {
		return fq
	}

	panic("field " + fieldName + " not found in entity")
}

// toSnakeCase 将驼峰命名转换为下划线命名
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}
