package repository

import "strings"

type whereBuilder struct {
	clauses []string
	args    []any
}

func newWhereBuilder(cap int) *whereBuilder {
	return &whereBuilder{
		clauses: make([]string, 0, cap),
		args:    make([]any, 0, cap),
	}
}

func (b *whereBuilder) add(clause string, arg any) {
	b.clauses = append(b.clauses, clause)
	b.args = append(b.args, arg)
}

func (b *whereBuilder) clause() string {
	return strings.Join(b.clauses, " AND ")
}
