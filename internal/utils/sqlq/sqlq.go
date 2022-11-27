package sqlq

import "strings"

type BuilderOption struct {
	Driver SQLDriver
	Tables Tables
}

func NewBuilder(opt BuilderOption) SQLQuery {
	return &sqlQuery{
		driver: opt.Driver,
		tables: opt.Tables,
	}
}

func ToOperator(value string) Operator {
	d := map[string]Operator{
		"*":           OperatorAll,
		"=":           OperatorEqual,
		"!=":          OperatorNotEqual,
		"<>":          OperatorNotEqual,
		">":           OperatorGreater,
		">=":          OperatorGreaterThen,
		"<":           OperatorLess,
		"<=":          OperatorLessThen,
		"LIKE":        OperatorLike,
		"NOT-LIKE":    OperatorNotLike,
		"ILIKE":       OperatorILike,
		"NOT-ILIKE":   OperatorNotILike,
		"BETWEEN":     OperatorBetween,
		"IN":          OperatorIn,
		"NOT-IN":      OperatorNotIn,
		"IS-NOT-NULL": OperatorIsNotNull,
		"IS-NULL":     OperatorIsNull,
	}

	value = strings.ToUpper(strings.ReplaceAll(value, " ", "-"))
	if opr, ok := d[value]; ok {
		return opr
	}

	return OperatorEmpty
}
