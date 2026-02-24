package main

// Задача: Query Builder — fluent interface для построения SQL-запросов.

import (
	"fmt"
	"strings"
)

type QueryBuilder interface {
	Select(fields ...string) QueryBuilder
	From(table string) QueryBuilder
	Where(condition string, args ...interface{}) QueryBuilder
	Join(table string, condition string) QueryBuilder
	OrderBy(field string, desc bool) QueryBuilder
	Limit(limit int) QueryBuilder
	Offset(offset int) QueryBuilder
	Build() (query string, args []interface{}, err error)
}

type InsertBuilder interface {
	Into(table string) InsertBuilder
	Values(values map[string]interface{}) InsertBuilder
	Build() (query string, args []interface{}, err error)
}

type UpdateBuilder interface {
	Table(table string) UpdateBuilder
	Set(field string, value interface{}) UpdateBuilder
	Where(condition string, args ...interface{}) UpdateBuilder
	Build() (query string, args []interface{}, err error)
}

// --- SQLQueryBuilder ---

type joinClause struct{ table, condition string }
type whereClause struct{ condition string; args []interface{} }
type orderClause struct{ field string; desc bool }

type SQLQueryBuilder struct {
	fields  []string
	table   string
	joins   []joinClause
	wheres  []whereClause
	orders  []orderClause
	limit   int
	offset  int
	hasLim  bool
	hasOff  bool
}

func NewSelect() *SQLQueryBuilder { return &SQLQueryBuilder{} }

func (b *SQLQueryBuilder) Select(fields ...string) QueryBuilder {
	b.fields = append(b.fields, fields...)
	return b
}
func (b *SQLQueryBuilder) From(table string) QueryBuilder { b.table = table; return b }
func (b *SQLQueryBuilder) Where(cond string, args ...interface{}) QueryBuilder {
	b.wheres = append(b.wheres, whereClause{cond, args})
	return b
}
func (b *SQLQueryBuilder) Join(table, cond string) QueryBuilder {
	b.joins = append(b.joins, joinClause{table, cond})
	return b
}
func (b *SQLQueryBuilder) OrderBy(field string, desc bool) QueryBuilder {
	b.orders = append(b.orders, orderClause{field, desc})
	return b
}
func (b *SQLQueryBuilder) Limit(n int) QueryBuilder  { b.limit = n; b.hasLim = true; return b }
func (b *SQLQueryBuilder) Offset(n int) QueryBuilder { b.offset = n; b.hasOff = true; return b }

func (b *SQLQueryBuilder) Build() (string, []interface{}, error) {
	if b.table == "" {
		return "", nil, fmt.Errorf("FROM clause is required")
	}
	cols := "*"
	if len(b.fields) > 0 {
		cols = strings.Join(b.fields, ", ")
	}
	sb := &strings.Builder{}
	sb.WriteString("SELECT ")
	sb.WriteString(cols)
	sb.WriteString(" FROM ")
	sb.WriteString(b.table)

	for _, j := range b.joins {
		sb.WriteString(" JOIN ")
		sb.WriteString(j.table)
		sb.WriteString(" ON ")
		sb.WriteString(j.condition)
	}

	var allArgs []interface{}
	for i, w := range b.wheres {
		if i == 0 {
			sb.WriteString(" WHERE ")
		} else {
			sb.WriteString(" AND ")
		}
		sb.WriteString(w.condition)
		allArgs = append(allArgs, w.args...)
	}

	for i, o := range b.orders {
		if i == 0 {
			sb.WriteString(" ORDER BY ")
		} else {
			sb.WriteString(", ")
		}
		sb.WriteString(o.field)
		if o.desc {
			sb.WriteString(" DESC")
		}
	}

	if b.hasLim {
		sb.WriteString(fmt.Sprintf(" LIMIT %d", b.limit))
	}
	if b.hasOff {
		sb.WriteString(fmt.Sprintf(" OFFSET %d", b.offset))
	}
	return sb.String(), allArgs, nil
}

// --- SQLInsertBuilder ---

type SQLInsertBuilder struct {
	table  string
	values map[string]interface{}
}

func NewInsert() *SQLInsertBuilder { return &SQLInsertBuilder{} }

func (b *SQLInsertBuilder) Into(table string) InsertBuilder { b.table = table; return b }
func (b *SQLInsertBuilder) Values(v map[string]interface{}) InsertBuilder { b.values = v; return b }

func (b *SQLInsertBuilder) Build() (string, []interface{}, error) {
	if b.table == "" {
		return "", nil, fmt.Errorf("INTO clause is required")
	}
	if len(b.values) == 0 {
		return "", nil, fmt.Errorf("VALUES is empty")
	}
	var cols, placeholders []string
	var args []interface{}
	for col, val := range b.values {
		cols = append(cols, col)
		placeholders = append(placeholders, "?")
		args = append(args, val)
	}
	q := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		b.table, strings.Join(cols, ", "), strings.Join(placeholders, ", "))
	return q, args, nil
}

// --- SQLUpdateBuilder ---

type SQLUpdateBuilder struct {
	table  string
	sets   []struct{ field string; value interface{} }
	wheres []whereClause
}

func NewUpdate() *SQLUpdateBuilder { return &SQLUpdateBuilder{} }

func (b *SQLUpdateBuilder) Table(t string) UpdateBuilder { b.table = t; return b }
func (b *SQLUpdateBuilder) Set(field string, value interface{}) UpdateBuilder {
	b.sets = append(b.sets, struct{ field string; value interface{} }{field, value})
	return b
}
func (b *SQLUpdateBuilder) Where(cond string, args ...interface{}) UpdateBuilder {
	b.wheres = append(b.wheres, whereClause{cond, args})
	return b
}

func (b *SQLUpdateBuilder) Build() (string, []interface{}, error) {
	if b.table == "" || len(b.sets) == 0 {
		return "", nil, fmt.Errorf("TABLE and SET clauses are required")
	}
	var setClauses []string
	var args []interface{}
	for _, s := range b.sets {
		setClauses = append(setClauses, s.field+" = ?")
		args = append(args, s.value)
	}
	sb := &strings.Builder{}
	sb.WriteString("UPDATE ")
	sb.WriteString(b.table)
	sb.WriteString(" SET ")
	sb.WriteString(strings.Join(setClauses, ", "))

	for i, w := range b.wheres {
		if i == 0 {
			sb.WriteString(" WHERE ")
		} else {
			sb.WriteString(" AND ")
		}
		sb.WriteString(w.condition)
		args = append(args, w.args...)
	}
	return sb.String(), args, nil
}

func main() {
	q, args, _ := NewSelect().
		Select("id", "name", "email").
		From("users").
		Join("orders", "users.id = orders.user_id").
		Where("age > ?", 18).
		Where("active = ?", true).
		OrderBy("name", false).
		Limit(10).
		Offset(20).
		Build()
	fmt.Println("SELECT:", q, args)

	q, args, _ = NewInsert().
		Into("users").
		Values(map[string]interface{}{"name": "Alice", "age": 25}).
		Build()
	fmt.Println("INSERT:", q, args)

	q, args, _ = NewUpdate().
		Table("users").
		Set("name", "Bob").
		Set("age", 30).
		Where("id = ?", 1).
		Build()
	fmt.Println("UPDATE:", q, args)
}
