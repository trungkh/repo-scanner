package sqlq

import (
	"fmt"
	"strings"

	"repo-scanner/internal/utils/serror"
	"repo-scanner/internal/utils/utarray"

	"github.com/gearintellix/u2"
)

type QPrepare interface {
	Syntax() string
	Table() string
	Type() string
	Insert(data TableData) (pars [][]interface{}, errx serror.SError)
}

type qPrepare struct {
	_syntax  string
	_table   string
	_type    string
	_columns []string
}

func (ox sqlQuery) PrepareInsert(table string, data TableData) (prp QPrepare, pars [][]interface{}, errx serror.SError) {
	tb, ok := ox.tables[table]
	if ok {
		q := `
			INSERT INTO __table__ (__columns__)
			SELECT __values__
		`

		colNames := []string{}
		cols := []string{}

		for _, v := range data.Datas {
			for k2 := range v {
				if col, ok := tb.Columns[k2]; ok {
					if !utarray.IsExist(k2, cols) {
						cols = append(cols, k2)
						colNames = append(colNames, ox.driver.SQLColumnEscape(col.Name))
					}
					continue
				}

				errx = serror.Newc(fmt.Sprintf("Cannot find column %s", k2), "while sqlq_insert")
				return prp, pars, errx
			}
		}

		vals := []string{}
		for k := range colNames {
			vals = append(vals, fmt.Sprintf("$%d", k+1))
		}

		syntx := u2.Binding(q, map[string]string{
			"table":   fmt.Sprintf("%s.%s", ox.driver.SQLColumnEscape(tb.Schema), ox.driver.SQLColumnEscape(tb.Table)),
			"columns": strings.Join(colNames, ", "),
			"values":  strings.Join(vals, ", "),
		})

		prp = NewQPrepare(syntx, table, "insert", cols)

		pars, errx = prp.Insert(data)
		return prp, pars, errx
	}

	return prp, pars, serror.Newc(fmt.Sprintf("Table %s not found", table), "while sqlq_insert")
}

func NewQPrepare(syntax string, table string, typ string, columns []string) QPrepare {
	return &qPrepare{
		_syntax:  syntax,
		_table:   table,
		_type:    typ,
		_columns: columns,
	}
}

func (ox qPrepare) Syntax() string {
	return ox._syntax
}

func (ox qPrepare) Table() string {
	return ox._table
}

func (ox qPrepare) Type() string {
	return ox._type
}

func (ox qPrepare) Insert(data TableData) (pars [][]interface{}, errx serror.SError) {
	pars = [][]interface{}{}

	if ox.Type() == "insert" {
		for _, v := range data.Datas {
			par := []interface{}{}

			for _, v2 := range ox._columns {
				if cur, ok := v[v2]; ok {
					par = append(par, cur)
					continue
				}
				par = append(par, nil)
			}

			pars = append(pars, par)
		}

		return pars, errx
	}

	errx = serror.Newc(fmt.Sprintf("Wrong prepare type of %s", ox.Type()), "while insert_on_sqlq_prepare")
	return pars, errx
}
