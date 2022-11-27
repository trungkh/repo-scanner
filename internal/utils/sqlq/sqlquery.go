package sqlq

import (
	"fmt"
	"strings"
	"time"

	"github.com/gearintellix/u2"

	"repo-scanner/internal/utils/serror"
	"repo-scanner/internal/utils/utarray"
	"repo-scanner/internal/utils/utinterface"
	"repo-scanner/internal/utils/utstring"
	"repo-scanner/internal/utils/uttime"
)

type EAVRole string

const (
	EAVRoleNone        EAVRole = "-"
	EAVRoleKey         EAVRole = "key"
	EAVRoleGroup       EAVRole = "group"
	EAVRoleDescription EAVRole = "description"
	EAVRoleValue       EAVRole = "value"
)

type Arguments struct {
	Offset     int64                  `json:"offset"`
	Limit      int                    `json:"limit"`
	Sorting    []string               `json:"sorting"`
	Conditions map[string]interface{} `json:"conditions"`
	Fields     []string               `json:"fields"`
}

type EAVData struct {
	Meta map[string]interface{}
	Data map[string]EAVItem
}

type EAVItem struct {
	ID          int64
	key         string
	group       string
	Description string
	Value       string
}

type TableData struct {
	Conditions map[string]interface{}
	Datas      []AttributeRow
	Returns    map[string]interface{}
}

type AttributeRow map[string]interface{}

type SQLQuery interface {
	Driver() SQLDriver
	Tables() Tables
	TableByKey(key string) *Table

	Select(q string, args Arguments, allow []string) (string, serror.SError)
	Insert(table string, data TableData) (qo string, errx serror.SError)
	Update(table string, data TableData) (qo string, errx serror.SError)
	Set(table string, data TableData) (qo string, errx serror.SError)
	SetV2(table string, data TableData) (qo string, errx serror.SError)

	PrepareInsert(table string, data TableData) (prp QPrepare, pars [][]interface{}, errx serror.SError)

	InsertEAV(table string, data EAVData) (qo string, errx serror.SError)
	UpdateEAV(table string, data EAVData) (qo string, errx serror.SError)
	SetEAV(table string, data EAVData, dels map[string]interface{}) (qo string, errx serror.SError)
}

type sqlQuery struct {
	driver SQLDriver
	tables Tables
}

func (ox *EAVData) solving() {
	for k, v := range ox.Data {
		v.key = k
		if i := strings.Index(k, "."); i > 0 {
			v.group = utstring.Sub(k, 0, i)
			v.key = utstring.Sub(k, i+1, 0)
		}

		ox.Data[k] = v
	}
}

func (ox sqlQuery) Driver() SQLDriver {
	return ox.driver
}

func (ox sqlQuery) Tables() Tables {
	return ox.tables
}

func (ox sqlQuery) TableByKey(key string) *Table {
	if cur, ok := ox.tables[key]; ok {
		return cur
	}
	return nil
}

func (ox sqlQuery) Insert(table string, data TableData) (qo string, errx serror.SError) {
	tb, ok := ox.tables[table]
	if ok {
		if _, ok := tb.Columns[tb.Primary]; !ok || tb.Primary == "" {
			errx = serror.Newc(fmt.Sprintf("No primary key on table %s", table), "while sqlq_insert")
			return qo, errx
		}

		q := `
			INSERT INTO __table__ (__columns__)
			VALUES __values__
			__returning__
		`

		cols := []string{}
		colNames := []string{}

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
				return qo, errx
			}
		}

		pars := map[string]string{
			"table":     fmt.Sprintf("%s.%s", ox.driver.SQLColumnEscape(tb.Schema), ox.driver.SQLColumnEscape(tb.Table)),
			"columns":   strings.Join(colNames, ", "),
			"values":    "",
			"returning": "",
		}

		returns := []string{}
		if tb.Primary != "" {
			ccol := tb.Columns[tb.Primary]
			returns = append(returns, fmt.Sprintf("%s.%s", ox.driver.SQLColumnEscape(tb.Table), ox.driver.SQLColumnEscape(ccol.Name)))
		}

		if data.Returns != nil && len(data.Returns) > 0 {
			for k, v := range data.Returns {
				vx := v
				switch vy := v.(type) {
				case string:
					if !strings.HasPrefix(vy, "!") {
						if ccol, ok := tb.Columns[vy]; ok {
							if ccol.IsPrimary && tb.Primary != "" {
								returns[0] = fmt.Sprintf("%s %s", returns[0], ox.driver.SQLColumnEscape(k))
							}
							returns = append(returns, fmt.Sprintf("%s.%s %s", ox.driver.SQLColumnEscape(tb.Table), ox.driver.SQLColumnEscape(ccol.Name), ox.driver.SQLColumnEscape(k)))
							continue
						}

						errx = serror.Newc(fmt.Sprintf("Cannot find column %s", vy), "while sqlq_insert")
						return qo, errx
					}
					vx = utstring.Sub(vy, 1, 0)
				}

				if qq, ok := ox.driver.ToSQLValueQuery(vx); ok {
					returns = append(returns, fmt.Sprintf("%s %s", qq, ox.driver.SQLColumnEscape(k)))
					continue
				}

				errx = serror.Newc(fmt.Sprintf("Cannot parsing for %s", k), "while sqlq_insert")
				return qo, errx
			}
		}

		if len(returns) > 0 {
			pars["returning"] = fmt.Sprintf("RETURNING %s", strings.Join(returns, ", "))
		}

		vals := []string{}
		for _, v := range data.Datas {
			val := []string{}

			for _, v2 := range cols {
				if cur, ok := v[v2]; ok {
					if res, ok := ox.driver.ToSQLValueQuery(cur); ok {
						val = append(val, res)
						continue
					}
				}
				val = append(val, "NULL")
			}

			vals = append(vals, fmt.Sprintf("(%s)", strings.Join(val, ", ")))
		}

		pars["values"] = strings.Join(vals, ", ")

		qo = u2.Binding(q, pars)
		return qo, nil
	}

	return qo, serror.Newc(fmt.Sprintf("Table %s not found", table), "while sqlq_insert")
}

func (ox sqlQuery) Update(table string, data TableData) (qo string, errx serror.SError) {
	tb, ok := ox.tables[table]
	if ok {
		if len(data.Datas) != 1 {
			errx = serror.Newc("Data empty or more than one", "while sqlq_update")
			return qo, errx
		}

		q := `
			UPDATE __table__
			SET __updates__
			WHERE __conditions__
		`

		conds := []string{}
		upds := []string{}

		// process conditions
		{
			for k, v := range data.Conditions {
				var operator Operator
				key := k

				operator = OperatorEqual

				if i := strings.LastIndex(k, ":"); i > 0 {
					operator = ToOperator(utstring.Sub(k, i+1, 0))
					key = utstring.Sub(k, 0, i)
				}

				if col, ok := tb.Columns[key]; ok {
					qcond := ""
					qcond, ok = ox.driver.ToSQLConditionQuery(
						QColumn{col.Name},
						operator,
						v,
					)
					if !ok {
						errx = serror.Newc(fmt.Sprintf("Cannot parsing condition data on %s", key), "[sqlq][query][Update] while parsing conditions")
						return qo, errx
					}

					conds = append(conds, qcond)
					continue
				}

				errx = serror.Newc(fmt.Sprintf("Column %s not found", key), "[sqlq][query][Update] while parsing conditions")
				return qo, errx
			}
		}

		// process values
		{
			for k, v := range data.Datas[0] {
				if col, ok := tb.Columns[k]; ok {
					if val, ok := ox.driver.ToSQLValueQuery(v); ok {
						upds = append(upds, fmt.Sprintf("%s = %s", ox.driver.SQLColumnEscape(col.Name), val))
					}

					continue
				}

				errx = serror.Newc(fmt.Sprintf("Cannot find column %s", k), "while sqlq_update")
				return qo, errx
			}
		}

		pars := map[string]string{
			"table":      fmt.Sprintf("%s.%s", ox.driver.SQLColumnEscape(tb.Schema), ox.driver.SQLColumnEscape(tb.Table)),
			"updates":    strings.Join(upds, ", "),
			"conditions": strings.Join(conds, " AND "),
		}

		qo = u2.Binding(q, pars)
		return qo, nil
	}

	return qo, serror.Newc(fmt.Sprintf("Table %s not found", table), "while sqlq_update")
}

func (ox sqlQuery) Set(table string, data TableData) (qo string, errx serror.SError) {
	tb, ok := ox.tables[table]
	if ok {
		if len(data.Datas) <= 0 {
			errx = serror.Newc("Data is empty", "while sqlq_set")
			return qo, errx
		}

		q := `
			INSERT INTO __table__ (__columns__)
			VALUES __values__
			ON CONFLICT (__foreigns__)
			DO UPDATE SET __updates__
			__returning__
		`

		keys := []string{}
		columns := []string{}
		values := []string{}
		updates := []string{}
		foreigns := []string{}
		returns := []string{}

		// condKeys := []string{}
		mostValue := map[int]string{}

		// preconditions
		{
			for k, v := range data.Conditions {
				kx := k

				isDynamic := false
				if strings.HasPrefix(kx, "~") {
					kx = utstring.Sub(kx, 1, 0)
					isDynamic = true
				}

				isNotToUpdate := false
				if strings.HasPrefix(kx, "!") {
					kx = utstring.Sub(kx, 1, 0)
					isNotToUpdate = true
				}

				if utarray.IsExist(kx, keys) {
					errx = serror.Newc(fmt.Sprintf("Condition %s has declared multiple times", kx), "@")
					return qo, errx
				}

				if cur, ok := tb.Columns[kx]; ok {
					if !isNotToUpdate {
						foreigns = append(foreigns, ox.driver.SQLColumnEscape(cur.Name))
					}

					keys = append(keys, kx)
					if !isDynamic && !isNotToUpdate {
						val, _ := ox.driver.ToSQLValueQuery(v)
						mostValue[len(columns)] = val
					}

					columns = append(columns, ox.driver.SQLColumnEscape(cur.Name))
					continue
				}

				errx = serror.Newc(fmt.Sprintf("Cannot find column %s", k), "while sqlq_set")
				return qo, errx
			}

			for _, v := range data.Datas {
				for k2 := range v {
					if !utarray.IsExist(k2, keys) {
						if cur, ok := tb.Columns[k2]; ok {
							keys = append(keys, k2)

							cn := ox.driver.SQLColumnEscape(cur.Name)
							columns = append(columns, cn)

							updates = append(updates, fmt.Sprintf("%s = %s.%s", cn, ox.driver.SQLColumnEscape("excluded"), cn))
							continue
						}

						errx = serror.Newc(fmt.Sprintf("Cannot find column %s", k2), "while sqlq_set")
						return qo, errx
					}
				}
			}
		}

		// process rows
		{
			for _, v := range data.Datas {
				vx := []string{}
				for k2, v2 := range keys {
					if cur, ok := mostValue[k2]; ok {
						vx = append(vx, cur)
						continue
					}

					if cur, ok := v[v2]; ok {
						if val, ok := ox.driver.ToSQLValueQuery(cur); ok {
							vx = append(vx, val)
							continue
						}

						vx = append(vx, "NULL")
						continue
					}

					vx = append(vx, "DEFAULT")
				}

				values = append(values, fmt.Sprintf("(%s)", strings.Join(vx, ", ")))
			}
		}

		// process returning
		{
			if tb.Primary != "" {
				ccol := tb.Columns[tb.Primary]
				returns = append(returns, fmt.Sprintf("%s.%s", ox.driver.SQLColumnEscape(tb.Table), ox.driver.SQLColumnEscape(ccol.Name)))
			}

			if data.Returns != nil && len(data.Returns) > 0 {
				for k, v := range data.Returns {
					vx := v
					switch vy := v.(type) {
					case string:
						if !strings.HasPrefix(vy, "!") {
							if ccol, ok := tb.Columns[vy]; ok {
								if ccol.IsPrimary && tb.Primary != "" {
									returns[0] = fmt.Sprintf("%s %s", returns[0], ox.driver.SQLColumnEscape(k))
								}
								returns = append(returns, fmt.Sprintf("%s.%s %s", ox.driver.SQLColumnEscape(tb.Table), ox.driver.SQLColumnEscape(ccol.Name), ox.driver.SQLColumnEscape(k)))
								continue
							}

							errx = serror.Newc(fmt.Sprintf("Cannot find column %s", vy), "while sqlq_insert")
							return qo, errx
						}
						vx = utstring.Sub(vy, 1, 0)
					}

					if qq, ok := ox.driver.ToSQLValueQuery(vx); ok {
						returns = append(returns, fmt.Sprintf("%s %s", qq, ox.driver.SQLColumnEscape(k)))
						continue
					}

					errx = serror.Newc(fmt.Sprintf("Cannot parsing for %s", k), "while sqlq_insert")
					return qo, errx
				}
			}
		}

		pars := map[string]string{
			"table":     fmt.Sprintf("%s.%s", ox.driver.SQLColumnEscape(tb.Schema), ox.driver.SQLColumnEscape(tb.Table)),
			"columns":   strings.Join(columns, ", "),
			"values":    strings.Join(values, ", "),
			"foreigns":  strings.Join(foreigns, ", "),
			"updates":   strings.Join(updates, ", "),
			"returning": "",
		}

		if len(returns) > 0 {
			pars["returning"] = fmt.Sprintf("RETURNING %s", strings.Join(returns, ", "))
		}

		qo = u2.Binding(q, pars)
		return qo, nil
	}

	return qo, serror.Newc(fmt.Sprintf("Table %s not found", table), "while sqlq_set")
}

func (ox sqlQuery) SetV2(table string, data TableData) (qo string, errx serror.SError) {
	if len(data.Datas) <= 0 {
		errx = serror.New("Data is empty")
		return qo, errx
	}

	tb, ok := ox.tables[table]
	if ok {
		if _, ok := tb.Columns[tb.Primary]; !ok || tb.Primary == "" {
			errx = serror.Newf("No primary key on table %s", table)
			return qo, errx
		}

		var q string

		parameters := map[string]string{
			"table": fmt.Sprintf("%s.%s", ox.driver.SQLColumnEscape(tb.Schema), ox.driver.SQLColumnEscape(tb.Table)),
		}

		switch ox.driver {
		case DriverPostgreSQL:
			q = `
				WITH "A"("_idx", __columns__) AS (
					VALUES __values__
				),
				"B"("_idx", __x.returningColumns__) AS (
					__!updateStatement__
				),
				"C"(__x.returningColumns__) AS (
					INSERT INTO __table__ (__columns__)
					SELECT
						__columns__
					FROM "A"
					WHERE
						"A"."_idx" NOT IN (SELECT "_idx" FROM "B")
					RETURNING __returnings__
				)
				SELECT __returningColumns__
				FROM (
					SELECT "_idx", __returningColumns__ FROM "B"
					UNION ALL
					SELECT "A"."_idx", __y.returningColumns__ FROM "C"
					JOIN "A" ON __x.conditions__
				) "D"
				ORDER BY "_idx" ASC;
			`

			var (
				keys           = []string{}
				columns        = []string{}
				values         = []string{}
				updates        = []string{}
				conditions     = []string{}
				xConditions    = []string{}
				returns        = []string{}
				returnColumns  = []string{}
				xReturnColumns = []string{}
				yReturnColumns = []string{}

				castValue = make(map[string]string)
				mostValue = make(map[int]string)
			)

			// process options
			{
				optionBinds := map[string]string{
					"!updateStatement": `
						UPDATE __table__
						SET __updates__
						FROM "A"
						WHERE __conditions__
						RETURNING "A"."_idx", __returnings__
					`,
				}

				if val, ok := data.Conditions["$noUpdate"]; ok {
					if utinterface.ToBool(val, false) {
						optionBinds["!updateStatement"] = `
							SELECT "A"."_idx", __returnings__
							FROM "A"
							JOIN __table__ ON __conditions__
						`
					}
				}

				q = u2.Binding(q, optionBinds)
			}

			// process conditions
			{
				for k, v := range data.Conditions {
					kx := k

					if len(k) <= 0 {
						continue
					}

					var (
						isConditionOnly = false
						isMostValue     = false
						isDynamic       = false
					)

					switch string(k[0]) {
					case "$":
						continue

					case "~":
						kx = utstring.Sub(kx, 1, 0)
						isDynamic = true

					case "=":
						kx = utstring.Sub(kx, 1, 0)
						isMostValue = true

					case "!", "-":
						kx = utstring.Sub(kx, 1, 0)

					case "#":
						kx = utstring.Sub(kx, 1, 0)
						castValue[kx] = fmt.Sprintf("%v", v)
						continue

					case "@":
						conq, ok := ox.driver.ToSQLValueQuery(v)
						if !ok {
							errx = serror.Newf("Failed to parsing custom condition %s", kx)
							return qo, errx
						}

						conditions = append(conditions, conq)
						continue

					default:
						isConditionOnly = true
					}

					var (
						opr = OperatorEqual
						kxs = utstring.CleanSpit(kx, ":")
					)
					if len(kxs) > 1 {
						kx = kxs[0]
						opr = ToOperator(kxs[1])
					}

					if utarray.IsExist(kx, keys) {
						errx = serror.Newf("Condition %s has declared multiple times", kx)
						return qo, errx
					}

					if cur, ok := tb.Columns[kx]; ok {
						if isConditionOnly || isDynamic {
							ccol := QColumn{tb.Schema, tb.Table, cur.Name}

							if isConditionOnly {
								conq, ok := ox.driver.ToSQLConditionQuery(ccol, opr, v)
								if !ok {
									errx = serror.Newf("Failed to parsing condition %s", kx)
									return qo, errx
								}

								conditions = append(conditions, conq)
								continue
							}

							if isDynamic {
								conq, ok := ox.driver.ToSQLConditionQuery(ccol, opr, QColumn{"A", cur.Name})
								if !ok {
									errx = serror.Newf("Failed to parsing condition %s", kx)
									return qo, errx
								}

								conditions = append(conditions, conq)
							}

							// collect foreign keys
							{
								conq, ok := ox.driver.ToSQLConditionQuery(QColumn{"A", cur.Name}, opr, QColumn{"C", cur.Name})
								if !ok {
									errx = serror.Newf("Failed to parsing condition %s", kx)
									return qo, errx
								}

								xConditions = append(xConditions, conq)

								returns = append(returns, fmt.Sprintf(
									"%s.%s %s",
									ox.driver.SQLColumnEscape(tb.Table),
									ox.driver.SQLColumnEscape(cur.Name),
									ox.driver.SQLColumnEscape(cur.Name),
								))
								xReturnColumns = append(xReturnColumns, ox.driver.SQLColumnEscape(cur.Name))
							}
						}

						keys = append(keys, kx)
						if isMostValue {
							val, _ := ox.driver.ToSQLValueQuery(v)
							mostValue[len(columns)] = val
						}

						columns = append(columns, ox.driver.SQLColumnEscape(cur.Name))
						continue
					}

					errx = serror.Newf("Cannot find column %s", kx)
					return qo, errx
				}

				for _, v := range data.Datas {
					for k2 := range v {
						if !utarray.IsExist(k2, keys) {
							if cur, ok := tb.Columns[k2]; ok {
								keys = append(keys, k2)

								cn := ox.driver.SQLColumnEscape(cur.Name)
								columns = append(columns, cn)

								updates = append(updates, fmt.Sprintf("%s = %s.%s", cn, ox.driver.SQLColumnEscape("A"), cn))
								continue
							}

							errx = serror.Newf("Cannot find column %s", k2)
							return qo, errx
						}
					}
				}
			}

			// process datas
			{
				for k, v := range data.Datas {
					vx := []string{
						utstring.IntToString(k),
					}
					for k2, v2 := range keys {
						var cst string
						if cur, ok := castValue[v2]; ok {
							cst = fmt.Sprintf("::%s", cur)
						}

						if cur, ok := mostValue[k2]; ok {
							vx = append(vx, cur+cst)
							continue
						}

						if cur, ok := v[v2]; ok {
							if val, ok := ox.driver.ToSQLValueQuery(cur); ok {
								vx = append(vx, val+cst)
								continue
							}

							vx = append(vx, "NULL"+cst)
							continue
						}

						vx = append(vx, "NULL"+cst)
					}

					values = append(values, fmt.Sprintf("(%s)", strings.Join(vx, ", ")))
				}
			}

			// process returnings
			{
				if cur, ok := tb.Columns[tb.Primary]; ok {
					returns = append(returns,
						fmt.Sprintf(
							"%s.%s %s",
							ox.driver.SQLColumnEscape(tb.Table),
							ox.driver.SQLColumnEscape(cur.Name),
							ox.driver.SQLColumnEscape(cur.Alias),
						),
					)

					returnColumns = append(returnColumns, ox.driver.SQLColumnEscape(cur.Alias))
					yReturnColumns = append(
						yReturnColumns,
						fmt.Sprintf(
							"%s.%s",
							ox.driver.SQLColumnEscape("C"),
							ox.driver.SQLColumnEscape(cur.Alias),
						),
					)
				}

				if len(data.Returns) > 0 {
					for k, v := range data.Returns {
						if k == tb.Primary {
							continue
						}

						vx := v
						switch vy := v.(type) {
						case string:
							if !strings.HasPrefix(vy, "!") {
								if ccol, ok := tb.Columns[vy]; ok {
									if ccol.IsPrimary || tb.Primary == vy {
										continue
									}

									if nm := fmt.Sprintf(
										"%s.%s %s",
										ox.driver.SQLColumnEscape(tb.Table),
										ox.driver.SQLColumnEscape(ccol.Name),
										ox.driver.SQLColumnEscape(k),
									); !utarray.IsExist(nm, returns) {
										returns = append(returns, nm)
									}

									returnColumns = append(returnColumns, ox.driver.SQLColumnEscape(k))
									yReturnColumns = append(
										yReturnColumns,
										fmt.Sprintf(
											"%s.%s",
											ox.driver.SQLColumnEscape("C"),
											ox.driver.SQLColumnEscape(k),
										),
									)
									continue
								}

								errx = serror.Newf("Cannot find column %s", vy)
								return qo, errx
							}

							vx = utstring.Sub(vy, 1, 0)
						}

						if qq, ok := ox.driver.ToSQLValueQuery(vx); ok {
							returns = append(
								returns,
								fmt.Sprintf("%s %s", qq, ox.driver.SQLColumnEscape(k)),
							)

							returnColumns = append(returnColumns, ox.driver.SQLColumnEscape(k))
							yReturnColumns = append(
								yReturnColumns,
								fmt.Sprintf(
									"%s.%s",
									ox.driver.SQLColumnEscape("C"),
									ox.driver.SQLColumnEscape(k),
								),
							)
							continue
						}

						errx = serror.Newf("Cannot parsing for %s", k)
						return qo, errx
					}
				}
			}

			parameters["columns"] = strings.Join(columns, ", ")
			parameters["values"] = strings.Join(values, ", ")
			parameters["updates"] = strings.Join(updates, ", ")
			parameters["conditions"] = strings.Join(conditions, " AND ")
			parameters["x.conditions"] = strings.Join(xConditions, " AND ")
			parameters["returnings"] = ""
			parameters["y.returningColumns"] = ""
			parameters["returningColumns"] = ""
			parameters["y.returningColumns"] = ""

			if len(returns) > 0 {
				parameters["returnings"] = strings.Join(returns, ", ")
				parameters["returningColumns"] = strings.Join(returnColumns, ", ")
				parameters["y.returningColumns"] = strings.Join(yReturnColumns, ", ")
			}

			if len(xReturnColumns) > 0 {
				parameters["x.returningColumns"] = strings.Join(append(xReturnColumns, returnColumns...), ", ")
			}

		default:
			errx = serror.New("Driver not supported")
			return qo, errx
		}

		qo = u2.Binding(q, parameters)
		return qo, nil
	}

	errx = serror.Newf("Table %s not found", table)
	return qo, errx
}

func (ox sqlQuery) InsertEAV(table string, data EAVData) (qo string, errx serror.SError) {
	if len(data.Data) <= 0 {
		return qo, errx
	}

	tb, ok := ox.tables[table]
	if ok {
		if !tb.IsEAV {
			return qo, serror.Newc(fmt.Sprintf("Table %s is not EAV model", table), "while insert_eav")
		}

		data.solving()

		defs := tb.GetAvailableEAVColumns()

		q := ""
		switch ox.driver {
		case DriverPostgreSQL:
			q = `
				WITH V (__columns__) AS (
					VALUES __values__
				)
				INSERT INTO __table__ (__columns__)
				SELECT *
				FROM V
				__returning__
			`

		default:
			return qo, serror.Newc("Driver not supported", "while insert_eav")
		}

		pars := map[string]string{
			"table":     fmt.Sprintf("%s.%s", ox.driver.SQLColumnEscape(tb.Schema), ox.driver.SQLColumnEscape(tb.Table)),
			"returning": "",
		}

		keys := []string{}
		cols := []string{}

		if cur, ok := tb.Columns[tb.Primary]; ok {
			pars["returning"] = fmt.Sprintf("RETURNING %s", ox.driver.SQLColumnEscape(cur.Name))
		}

		for k2 := range data.Meta {
			if col, ok := tb.Columns[k2]; ok {
				if !utarray.IsExist(col.Name, cols) {
					keys = append(keys, k2)
					cols = append(cols, ox.driver.SQLColumnEscape(col.Name))
				}

			} else {
				return qo, serror.Newc(fmt.Sprintf("Field %s not exists", k2), "while insert_eav")
			}
		}

		eavs := []EAVRole{}
		for k, v := range defs {
			cols = append(cols, ox.driver.SQLColumnEscape(v.Name))
			eavs = append(eavs, k)
		}

		vals := []string{}
		for _, v := range data.Data {
			val := []string{}

			for _, v2 := range keys {
				if cur, ok := data.Meta[v2]; ok {
					if v3, ok := ox.driver.ToSQLValueQuery(cur); ok {
						val = append(val, v3)
						continue
					}
				}

				val = append(val, "NULL")
			}

			for _, v2 := range eavs {
				vc := ""
				switch v2 {
				case EAVRoleKey:
					vc = v.key

				case EAVRoleGroup:
					vc = v.group

				case EAVRoleDescription:
					vc = v.Description

				case EAVRoleValue:
					vc = v.Value
				}

				if v3, ok := ox.driver.ToSQLValueQuery(vc); ok {
					val = append(val, v3)
					continue
				}
				val = append(val, "")
			}

			vals = append(vals, fmt.Sprintf("(%s)", strings.Join(val, ", ")))
		}

		pars["values"] = strings.Join(vals, ", ")
		pars["columns"] = strings.Join(cols, ", ")

		qo = u2.Binding(q, pars)
		return qo, nil
	}
	return qo, serror.Newc(fmt.Sprintf("Table %s not found", table), "while insert_eav")
}

func (ox sqlQuery) SetEAV(table string, data EAVData, dels map[string]interface{}) (qo string, errx serror.SError) {
	if len(data.Data) <= 0 {
		return qo, errx
	}

	tb, ok := ox.tables[table]
	if ok {
		if !tb.IsEAV {
			return qo, serror.Newc(fmt.Sprintf("Table %s is not EAV model", table), "while sqlq EAV table check")
		}

		data.solving()

		defs := tb.GetAvailableEAVColumns()

		tableFullpath := fmt.Sprintf("%s.%s", ox.driver.SQLColumnEscape(tb.Schema), ox.driver.SQLColumnEscape(tb.Table))
		parameters := map[string]string{
			"table":     tableFullpath,
			"returning": "",
		}

		q := ""
		switch ox.driver {
		case DriverPostgreSQL:
			q = `
				WITH D (__columns__) AS (
					VALUES __values__
				),
				E AS (
					UPDATE __table__ SET
						__updates__
					FROM D
					WHERE
						__foreignsA__
				)
				INSERT INTO __table__ (
					__columns__
				)
				SELECT *
				FROM D
				__returning__
			`

			updates := []string{}
			values := []string{}
			foreignsA := []string{}
			foreignsB := []string{}

			if tb.Primary == "" {
				errx = serror.Newc(fmt.Sprintf("No primary key on table %s", table), "while sqlq primary table check")
				return qo, errx
			}

			if tb.SoftDelete != "" && dels == nil {
				errx = serror.Newc(fmt.Sprintf("Cannot soft delete on table %s", table), "while sqlq availability soft delete")
				return qo, errx
			}

			if cur, ok := tb.Columns[tb.Primary]; ok {
				parameters["returning"] = fmt.Sprintf("RETURNING %s", ox.driver.SQLColumnEscape(cur.Name))
			}

			now := time.Now()
			if cval, ok := dels["~"]; ok {
				cvtime, erry := uttime.Parse(cval)
				if erry != nil {
					now = cvtime
				}
			}

			if tb.SoftDelete != "" {
				isOk := false
				if ccol, ok := tb.Columns[tb.SoftDelete]; ok {
					if cval, ok := ox.driver.ToSQLValueQuery(now); ok {
						colName := ox.driver.SQLColumnEscape(ccol.Name)
						updates = append(updates, fmt.Sprintf("%s = %s", colName, cval))
						foreignsA = append(foreignsA, fmt.Sprintf("%s.%s IS NULL", tableFullpath, colName))
						isOk = true
					}
				}

				if !isOk && len(dels) <= 0 {
					return qo, errx
				}
			}

			for k, v := range dels {
				if k == "~" {
					continue
				}

				if ccol, ok := tb.Columns[k]; ok {
					if cval, ok := ox.driver.ToSQLValueQuery(v); ok {
						updates = append(updates, fmt.Sprintf("%s = %s", ox.driver.SQLColumnEscape(ccol.Name), cval))
						continue
					}

					errx = serror.Newc(fmt.Sprintf("Failed to processing value of %s", k), "while sqlq processing soft delete")
					return qo, errx
				}

				errx = serror.Newc(fmt.Sprintf("Field %s not exists", k), "while sqlq processing soft delete")
				return qo, errx
			}

			columns := []string{}
			constants := []string{}

			columnCasts := make(map[string]string)

			for k, v := range data.Meta {
				k2 := strings.ToLower(k)

				isForeign := true
				if strings.HasPrefix(k2, "~") {
					k2 = utstring.Sub(k2, 1, 0)
					isForeign = false
				}

				if strings.HasPrefix(k2, "#") {
					k2 = utstring.Sub(k2, 1, 0)
					columnCasts[k2] = utstring.Trim(fmt.Sprintf("%v", v))
					continue
				}

				if col, ok := tb.Columns[k2]; ok {
					if !utarray.IsExist(col.Name, columns) {
						k3 := ox.driver.SQLColumnEscape(col.Name)

						columns = append(columns, k3)

						if cval, ok := ox.driver.ToSQLValueQuery(v); ok {
							constants = append(constants, cval)
						}

						if isForeign {
							foreignsA = append(foreignsA, fmt.Sprintf("%s.%s = D.%s", tableFullpath, k3, k3))
							foreignsB = append(foreignsB, fmt.Sprintf("A.%s = D.%s", k3, k3))
						}
					}

				} else {
					return qo, serror.Newc(fmt.Sprintf("Field %s not exists", k), "while sqlq meta processing")
				}
			}

			eavs := []EAVRole{}
			eavCasts := []string{}

			for k, v := range defs {
				k2 := ox.driver.SQLColumnEscape(v.Name)

				eavCasts = append(eavCasts, columnCasts[strings.ToLower(v.Alias)])
				columns = append(columns, k2)
				eavs = append(eavs, k)

				switch k {
				case EAVRoleKey:
					fallthrough
				case EAVRoleGroup:
					foreignsA = append(foreignsA, fmt.Sprintf("%s.%s = D.%s", tableFullpath, k2, k2))
					foreignsB = append(foreignsB, fmt.Sprintf("A.%s = D.%s", k2, k2))
				}
			}

			for _, v := range data.Data {
				val := []string{}

				val = append(val, constants...)

				for k2, v2 := range eavs {
					vc := ""
					switch v2 {
					case EAVRoleKey:
						vc = v.key

					case EAVRoleGroup:
						vc = v.group

					case EAVRoleDescription:
						vc = v.Description

					case EAVRoleValue:
						vc = v.Value
					}

					cst := eavCasts[k2]
					if cst != "" {
						cst = fmt.Sprintf("::%s", cst)
					}

					if v3, ok := ox.driver.ToSQLValueQuery(vc); ok {
						val = append(val, v3+cst)
						continue
					}
					val = append(val, fmt.Sprintf("NULL%s", cst))
				}

				values = append(values, fmt.Sprintf("(%s)", strings.Join(val, ", ")))
			}

			parameters["values"] = strings.Join(values, ", ")
			parameters["columns"] = strings.Join(columns, ", ")
			parameters["updates"] = strings.Join(updates, ", ")
			parameters["foreignsA"] = strings.Join(foreignsA, " AND ")
			parameters["foreignsB"] = strings.Join(foreignsB, " AND ")

			if len(foreignsA) <= 0 {
				parameters["foreignsA"] = "FALSE"
			}
			if len(foreignsB) <= 0 {
				parameters["foreignsB"] = "FALSE"
			}

		default:
			return qo, serror.Newc("Driver not supported", "while sqlq SetEAV")
		}

		qo = u2.Binding(q, parameters)
		return qo, errx
	}
	return qo, serror.Newc(fmt.Sprintf("Table %s not found", table), "while insert_eav")
}

// TODO: update EAV must just update only, not insert too
func (ox sqlQuery) UpdateEAV(table string, data EAVData) (qo string, errx serror.SError) {
	if len(data.Data) <= 0 {
		return qo, errx
	}

	tb, ok := ox.tables[table]
	if ok {
		if !tb.IsEAV {
			return qo, serror.Newc(fmt.Sprintf("Table %s is not EAV model", table), "while update_eav")
		}

		data.solving()

		defs := tb.GetAvailableEAVColumns()

		q := ""
		switch ox.driver {
		case DriverPostgreSQL:
			q = `
				WITH D (__columns__) AS (
					VALUES __values__
				)
				INSERT INTO __table__ (
					__columns__
				)
				SELECT *
				FROM D
				ON CONFLICT (__foreigns__)
				DO UPDATE SET
					__updates__
			`

			pars := map[string]string{
				"table": fmt.Sprintf("%s.%s", ox.driver.SQLColumnEscape(tb.Schema), ox.driver.SQLColumnEscape(tb.Table)),
			}

			keys := []string{}
			columns := []string{}
			foreigns := []string{}
			updates := []string{}

			fields := []string{}
			for k := range data.Meta {
				k2 := strings.ToLower(k)

				isForeign := true
				if strings.HasPrefix(k2, "~") {
					k2 = utstring.Sub(k2, 1, 0)
					isForeign = false
				}

				if ccol, ok := tb.Columns[k2]; ok {
					cn := ox.driver.SQLColumnEscape(ccol.Name)

					keys = append(keys, k)
					fields = append(fields, k)
					columns = append(columns, cn)
					updates = append(updates, fmt.Sprintf("%s = %s.%s", cn, ox.driver.SQLColumnEscape("excluded"), cn))

					if isForeign {
						foreigns = append(foreigns, cn)
					}

					continue
				}

				errx = serror.Newc(fmt.Sprintf("Cannot find column %s", k), "while sqlq_insert_eav")
				return qo, errx
			}

			for k, v := range defs {
				v2 := ox.driver.SQLColumnEscape(v.Name)

				updates = append(updates, fmt.Sprintf("%s = %s.%s", v2, "excluded", v2))
				columns = append(columns, v2)
				keys = append(keys, string(k))

				switch k {
				case EAVRoleKey:
					fallthrough
				case EAVRoleGroup:
					foreigns = append(foreigns, v2)
				}
			}

			values := []string{}
			for _, v := range data.Data {
				vals := []string{}
				for _, v2 := range keys {
					if utarray.IsExist(v2, fields) {
						if res, ok := ox.driver.ToSQLValueQuery(data.Meta[v2]); ok {
							vals = append(vals, res)
							continue
						}

						errx = serror.Newc(fmt.Sprintf("Cannot parsing data %s", v2), "while sqlq_insert_eav")
						return qo, errx
					}

					var cv interface{}
					switch v2 {
					case string(EAVRoleKey):
						cv = v.key

					case string(EAVRoleGroup):
						cv = v.group

					case string(EAVRoleDescription):
						cv = v.Description

					case string(EAVRoleValue):
						cv = v.Value
					}

					if res, ok := ox.driver.ToSQLValueQuery(cv); ok {
						vals = append(vals, res)
						continue
					}
				}

				values = append(values, fmt.Sprintf("(%s)", strings.Join(vals, ", ")))
			}

			pars["columns"] = strings.Join(columns, ", ")
			pars["foreigns"] = strings.Join(foreigns, ", ")
			pars["updates"] = strings.Join(updates, ", ")
			pars["values"] = strings.Join(values, ", ")

			qo = u2.Binding(q, pars)

		default:
			return qo, serror.Newc("Driver not supported", "while insert_eav")
		}
	}
	return qo, errx
}
