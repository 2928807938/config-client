package queryutil

import (
	"fmt"

	"gorm.io/gorm"

	"config-client/share/repository"
	gormRepo "config-client/share/repository/gorm"
)

// FieldQuery 字段查询构建器，提供类型安全的查询条件构建
type FieldQuery struct {
	field string
}

// Field 创建字段查询构建器
func Field(fieldName string) *FieldQuery {
	return &FieldQuery{field: fieldName}
}

// GetColumnName 获取列名（用于直接传递给 GORM）
func (f *FieldQuery) GetColumnName() string {
	return f.field
}

// Eq 等于条件
func (f *FieldQuery) Eq(value interface{}) *repository.Condition {
	return repository.Eq(f.field, value)
}

// NotEq 不等于条件
func (f *FieldQuery) NotEq(value interface{}) *repository.Condition {
	return repository.NotEq(f.field, value)
}

// Gt 大于条件
func (f *FieldQuery) Gt(value interface{}) *repository.Condition {
	return repository.Gt(f.field, value)
}

// Gte 大于等于条件
func (f *FieldQuery) Gte(value interface{}) *repository.Condition {
	return repository.Gte(f.field, value)
}

// Lt 小于条件
func (f *FieldQuery) Lt(value interface{}) *repository.Condition {
	return repository.Lt(f.field, value)
}

// Lte 小于等于条件
func (f *FieldQuery) Lte(value interface{}) *repository.Condition {
	return repository.Lte(f.field, value)
}

// Like 模糊匹配条件
func (f *FieldQuery) Like(value string) *repository.Condition {
	return repository.Like(f.field, value)
}

// In 包含条件
func (f *FieldQuery) In(values interface{}) *repository.Condition {
	return repository.In(f.field, values)
}

// NotIn 不包含条件
func (f *FieldQuery) NotIn(values interface{}) *repository.Condition {
	return repository.NotIn(f.field, values)
}

// Between 区间条件
func (f *FieldQuery) Between(start, end interface{}) *repository.Condition {
	return repository.Between(f.field, start, end)
}

// IsNull 为空条件
func (f *FieldQuery) IsNull() *repository.Condition {
	return repository.IsNull(f.field)
}

// IsNotNull 不为空条件
func (f *FieldQuery) IsNotNull() *repository.Condition {
	return repository.IsNotNull(f.field)
}

// ==================== 便捷查询方法 ====================

// Eq 快捷等于条件
func Eq(field string, value interface{}) *repository.Condition {
	return repository.Eq(field, value)
}

// NotEq 快捷不等于条件
func NotEq(field string, value interface{}) *repository.Condition {
	return repository.NotEq(field, value)
}

// Gt 快捷大于条件
func Gt(field string, value interface{}) *repository.Condition {
	return repository.Gt(field, value)
}

// Gte 快捷大于等于条件
func Gte(field string, value interface{}) *repository.Condition {
	return repository.Gte(field, value)
}

// Lt 快捷小于条件
func Lt(field string, value interface{}) *repository.Condition {
	return repository.Lt(field, value)
}

// Lte 快捷小于等于条件
func Lte(field string, value interface{}) *repository.Condition {
	return repository.Lte(field, value)
}

// Like 快捷模糊匹配条件
func Like(field string, value string) *repository.Condition {
	return repository.Like(field, value)
}

// In 快捷包含条件
func In(field string, values interface{}) *repository.Condition {
	return repository.In(field, values)
}

// NotIn 快捷不包含条件
func NotIn(field string, values interface{}) *repository.Condition {
	return repository.NotIn(field, values)
}

// Between 快捷区间条件
func Between(field string, start, end interface{}) *repository.Condition {
	return repository.Between(field, start, end)
}

// IsNull 快捷为空条件
func IsNull(field string) *repository.Condition {
	return repository.IsNull(field)
}

// IsNotNull 快捷不为空条件
func IsNotNull(field string) *repository.Condition {
	return repository.IsNotNull(field)
}

// ==================== GORM 直接查询方法 ====================

// WhereEq 直接应用等于条件到 GORM DB
func WhereEq(db *gorm.DB, field string, value interface{}) *gorm.DB {
	return db.Where(fmt.Sprintf("%s = ?", field), value)
}

// WhereNotEq 直接应用不等于条件到 GORM DB
func WhereNotEq(db *gorm.DB, field string, value interface{}) *gorm.DB {
	return db.Where(fmt.Sprintf("%s != ?", field), value)
}

// WhereGt 直接应用大于条件到 GORM DB
func WhereGt(db *gorm.DB, field string, value interface{}) *gorm.DB {
	return db.Where(fmt.Sprintf("%s > ?", field), value)
}

// WhereGte 直接应用大于等于条件到 GORM DB
func WhereGte(db *gorm.DB, field string, value interface{}) *gorm.DB {
	return db.Where(fmt.Sprintf("%s >= ?", field), value)
}

// WhereLt 直接应用小于条件到 GORM DB
func WhereLt(db *gorm.DB, field string, value interface{}) *gorm.DB {
	return db.Where(fmt.Sprintf("%s < ?", field), value)
}

// WhereLte 直接应用小于等于条件到 GORM DB
func WhereLte(db *gorm.DB, field string, value interface{}) *gorm.DB {
	return db.Where(fmt.Sprintf("%s <= ?", field), value)
}

// WhereLike 直接应用模糊匹配条件到 GORM DB
func WhereLike(db *gorm.DB, field string, value string) *gorm.DB {
	return db.Where(fmt.Sprintf("%s LIKE ?", field), value)
}

// WhereIn 直接应用包含条件到 GORM DB
func WhereIn(db *gorm.DB, field string, values interface{}) *gorm.DB {
	return db.Where(fmt.Sprintf("%s IN ?", field), values)
}

// WhereNotIn 直接应用不包含条件到 GORM DB
func WhereNotIn(db *gorm.DB, field string, values interface{}) *gorm.DB {
	return db.Where(fmt.Sprintf("%s NOT IN ?", field), values)
}

// WhereBetween 直接应用区间条件到 GORM DB
func WhereBetween(db *gorm.DB, field string, start, end interface{}) *gorm.DB {
	return db.Where(fmt.Sprintf("%s BETWEEN ? AND ?", field), start, end)
}

// WhereIsNull 直接应用为空条件到 GORM DB
func WhereIsNull(db *gorm.DB, field string) *gorm.DB {
	return db.Where(fmt.Sprintf("%s IS NULL", field))
}

// WhereIsNotNull 直接应用不为空条件到 GORM DB
func WhereIsNotNull(db *gorm.DB, field string) *gorm.DB {
	return db.Where(fmt.Sprintf("%s IS NOT NULL", field))
}

// ==================== 排序方法 ====================

// OrderBy 添加排序（升序）
func OrderBy(db *gorm.DB, field string) *gorm.DB {
	return db.Order(field + string(repository.OrderAsc))
}

// OrderByDesc 添加排序（降序）
func OrderByDesc(db *gorm.DB, field string) *gorm.DB {
	return db.Order(field + string(repository.OrderDesc))
}

// ==================== 批量条件应用 ====================

// ApplyConditions 批量应用查询条件
func ApplyConditions(db *gorm.DB, conditions ...*repository.Condition) *gorm.DB {
	return gormRepo.ApplyConditions(db, conditions...)
}

// ApplyCondition 应用单个查询条件
func ApplyCondition(db *gorm.DB, condition *repository.Condition) *gorm.DB {
	return gormRepo.ApplyCondition(db, condition)
}

// ApplyOrderBys 批量应用排序规则
func ApplyOrderBys(db *gorm.DB, orderBys []repository.OrderBy) *gorm.DB {
	return gormRepo.ApplyOrderBys(db, orderBys)
}
