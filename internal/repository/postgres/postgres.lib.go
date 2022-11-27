package postgres

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	"github.com/gearintellix/structs"
	"github.com/gearintellix/u2"
	"github.com/jmoiron/sqlx"

	"repo-scanner/internal"
	"repo-scanner/internal/model"
	"repo-scanner/internal/utils/serror"
	"repo-scanner/internal/utils/sqlq"
	"repo-scanner/internal/utils/utarray"
	"repo-scanner/internal/utils/utinterface"
	"repo-scanner/internal/utils/utsql"
	"repo-scanner/internal/utils/utstring"

	log "github.com/sirupsen/logrus"
)

type psql struct {
	NumberHelp internal.INumberHelperRepository
	TrxRepo    internal.ITrxRepository
	DB         *sqlx.DB
	Q          sqlq.SQLQuery
}

type record struct {
	Fields []string
	Datas  map[string]interface{}
}

type ICopyToStruct interface {
	Cast(val interface{}) (res interface{}, err error)
}

type (
	nextCallback      func(row record) (next bool)
	preDetailCallback func(row model.WriteItem) (res sqlq.EAVData, errx serror.SError)
)

func (ox psql) ScanRecords(rows *sql.Rows, nextFN nextCallback) (resp []record, errx serror.SError) {
	resp = []record{}

	columns, err := rows.Columns()
	if err != nil {
		errx = serror.NewFromErrorc(err, "Failed to fetching columns")
		return resp, errx
	}

	typs, err := rows.ColumnTypes()
	if err != nil {
		errx = serror.NewFromErrorc(err, "Failed to fetching column types")
		return resp, errx
	}

	for {
		if !rows.Next() {
			break
		}

		row := make([]interface{}, len(columns))
		errx = utsql.ScanToArrayInterface(rows, row)
		if errx != nil {
			errx.AddComments("[repository][ScanRecords] while ScanToArrayInterface")
			return resp, errx
		}

		errx = utsql.PostgresRowTypeSolve(typs, &row)
		if errx != nil {
			errx.AddComments("[repository][ScanRecords] while PostgresRowTypeSolve")
			return resp, errx
		}

		for k, v := range row {
			switch vx := v.(type) {
			case float64:
				row[k] = ox.NumberHelp.Round(vx)
			}
		}

		item := record{
			Fields: []string{},
			Datas:  make(map[string]interface{}),
		}

		for k, v := range columns {
			item.Fields = append(item.Fields, v)
			item.Datas[v] = row[k]
		}

		next := true
		if nextFN != nil {
			next = nextFN(item)
		}

		resp = append(resp, item)

		if !next {
			break
		}
	}

	return resp, errx
}

func (ox psql) InsertTable(name string, tx *model.Trx, args sqlq.TableData) (resp model.WriteResult, errx serror.SError) {
	var err error

	auto := false
	if tx == nil {
		auto = true
	}

	var (
		q       string
		results = []model.WriteItem{}
	)

	q, errx = ox.Q.Insert(name, args)
	if errx != nil {
		errx.AddComments("[repository][InsertTable] while build insert query")
		return resp, errx
	}

	if auto {
		tx, errx = ox.TrxRepo.Create()
		if errx != nil {
			errx.AddComments("[repository][InsertTable] while Create transaction")
			return resp, errx
		}
	}

	defer func() {
		if errStr := recover(); errStr != nil {
			errx = serror.Newc(fmt.Sprintf("%+v", errStr), "Something when wrong")
		}

		if auto && errx != nil && tx != nil {
			errs := tx.Abort()
			if errs != nil {
				errs.AddComments("[repository][InsertTable] while Abort transaction")
				log.Error(errs)
			}
		}
	}()

	rowx, err := tx.Query(q)
	if err != nil {
		errx = serror.NewFromErrorc(err, "Failed to querying")
		return resp, errx
	}

	defer func() {
		if rowx != nil {
			rowx.Close()
		}
	}()

	var rowb []record
	rowb, errx = ox.ScanRecords(rowx, nil)
	if errx != nil {
		errx.AddComments("[repository][InsertTable] while ScanRecords")
		return resp, errx
	}

	if len(rowb) != len(args.Datas) {
		errx = serror.Newc("Length not same", "Something when wrong")
		return resp, errx
	}

	for k, v := range rowb {
		item := model.WriteItem{
			Key:   k,
			ID:    -1,
			Metas: make(model.Metas),
		}
		for k2, v2 := range v.Fields {
			vx := v.Datas[v2]
			if k2 == 0 {
				item.ID = utinterface.ToInt(vx, -1)
				item.RawID = vx
			}

			item.Metas[v2] = vx
		}
		results = append(results, item)
	}

	if auto {
		errx = tx.Admit()
		if errx != nil {
			errx.AddComments("[repository][InsertTable] while Admit transaction")
			return resp, errx
		}
	}

	return model.WriteResult{
		Items: results,
	}, errx
}

func (ox psql) SetTable(name string, tx *model.Trx, args sqlq.TableData) (resp model.WriteResult, errx serror.SError) {
	var err error

	auto := false
	if tx == nil {
		auto = true
	}

	var (
		q       string
		results = []model.WriteItem{}
	)

	q, errx = ox.Q.Set(name, args)
	if errx != nil {
		errx.AddComments("[repository][SetTable] while build set query")
		return resp, errx
	}

	if auto {
		tx, errx = ox.TrxRepo.Create()
		if errx != nil {
			errx.AddComments("[repository][SetTable] while Create transaction")
			return resp, errx
		}
	}

	defer func() {
		if errStr := recover(); errStr != nil {
			errx = serror.Newc(fmt.Sprintf("%+v", errStr), "Something when wrong")
		}

		if auto && errx != nil && tx != nil {
			errs := tx.Abort()
			if errs != nil {
				errx.AddComments("[repository][SetTable] while Abort transaction")
				log.Error(errs)
			}
		}
	}()

	rowx, err := tx.Query(q)
	if err != nil {
		errx = serror.NewFromErrorc(err, "Failed to querying")
		return resp, errx
	}

	defer func() {
		if rowx != nil {
			rowx.Close()
		}
	}()

	var rowb []record
	rowb, errx = ox.ScanRecords(rowx, nil)
	if errx != nil {
		errx.AddComments("[repository][SetTable] while ScanRecords")
		return resp, errx
	}

	if len(rowb) != len(args.Datas) {
		errx = serror.Newc("Length not same", "Something when wrong")
		return resp, errx
	}

	for k, v := range rowb {
		item := model.WriteItem{
			Key:   k,
			ID:    -1,
			Metas: make(model.Metas),
		}
		for k2, v2 := range v.Fields {
			vx := v.Datas[v2]
			if k2 == 0 {
				item.ID = utinterface.ToInt(vx, -1)
				item.RawID = vx
			}

			item.Metas[v2] = vx
		}
		results = append(results, item)
	}

	if auto {
		errx = tx.Admit()
		if errx != nil {
			errx.AddComments("[repository][SetTable] while Admit transaction")
			return resp, errx
		}
	}

	return model.WriteResult{
		Items: results,
	}, errx
}

func (ox psql) SetV2Table(name string, tx *model.Trx, args sqlq.TableData) (resp model.WriteResult, errx serror.SError) {
	var err error

	auto := false
	if tx == nil {
		auto = true
	}

	var (
		q       string
		results = []model.WriteItem{}
	)

	q, errx = ox.Q.SetV2(name, args)
	if errx != nil {
		errx.AddComments("[repository][SetV2Table] while SetV2")
		return resp, errx
	}

	if auto {
		tx, errx = ox.TrxRepo.Create()
		if errx != nil {
			errx.AddComments("[repository][SetV2Table] while Create transaction")
			return resp, errx
		}
	}

	defer func() {
		if errStr := recover(); errStr != nil {
			errx = serror.Newc(fmt.Sprintf("%+v", errStr), "Something when wrong")
		}

		if auto && errx != nil && tx != nil {
			errs := tx.Abort()
			if errs != nil {
				errx.AddComments("[repository][SetV2Table] while Abort transaction")
				log.Error(errs)
			}
		}
	}()

	rowx, err := tx.Query(q)
	if err != nil {
		errx = serror.NewFromErrorc(err, "Failed to querying")
		return resp, errx
	}

	defer func() {
		if rowx != nil {
			rowx.Close()
		}
	}()

	var rowb []record
	rowb, errx = ox.ScanRecords(rowx, nil)
	if errx != nil {
		errx.AddComments("[repository][SetV2Table] while ScanRecords")
		return resp, errx
	}

	if len(rowb) != len(args.Datas) {
		errx = serror.Newc("Length not same", "Something when wrong")
		return resp, errx
	}

	for k, v := range rowb {
		item := model.WriteItem{
			Key:   k,
			ID:    -1,
			Metas: make(model.Metas),
		}
		for k2, v2 := range v.Fields {
			vx := v.Datas[v2]
			if k2 == 0 {
				item.ID = utinterface.ToInt(vx, -1)
				item.RawID = vx
			}

			item.Metas[v2] = vx
		}
		results = append(results, item)
	}

	if auto {
		errx = tx.Admit()
		if errx != nil {
			errx.AddComments("[repository][SetV2Table] while Admit transaction")
			return resp, errx
		}
	}

	return model.WriteResult{
		Items: results,
	}, errx
}

func (ox psql) InsertHeaderEAVTable(header string, detail string, tx *model.Trx, args []sqlq.EAVData, preDetailFN preDetailCallback) (resp model.WriteResult, errx serror.SError) {
	var err error

	auto := false
	if tx == nil {
		auto = true
	}

	if auto {
		tx, errx = ox.TrxRepo.Create()
		if errx != nil {
			errx.AddComments("[repository][InsertHeaderEAVTable] while Create transaction")
			return resp, errx
		}
	}

	defer func() {
		if errStr := recover(); errStr != nil {
			errx = serror.Newc(fmt.Sprintf("%+v", errStr), "Something when wrong")
		}

		if auto && errx != nil && tx != nil {
			errs := tx.Abort()
			if errs != nil {
				errx.AddComments("[repository][InsertHeaderEAVTable] while Abort transaction")
				log.Error(errs)
			}
		}
	}()

	headers := sqlq.TableData{
		Datas: []sqlq.AttributeRow{},
	}
	for _, v := range args {
		headers.Datas = append(headers.Datas, v.Meta)
	}

	var headerr model.WriteResult
	headerr, errx = ox.InsertTable(header, tx, headers)
	if errx != nil {
		errx.AddComments("[repository][InsertHeaderEAVTable] while InsertTable")
		return resp, errx
	}

	if !headerr.IsSuccess() || headerr.AffectedCount() != int64(len(headers.Datas)) {
		errx = serror.Newc("Length not same", "Something when wrong")
		return resp, errx
	}

	results := []model.WriteItem{}
	for k, v := range headerr.Items {
		if v.Error != nil {
			return resp, v.Error
		}

		var (
			q    string
			eavd sqlq.EAVData
		)

		eavd, errx = preDetailFN(v)
		if errx != nil {
			errx.AddComments("[repository][InsertHeaderEAVTable] while preDetailFN")
			return resp, errx
		}

		childs := sqlq.EAVData{
			Meta: eavd.Meta,
			Data: args[k].Data,
		}
		for k, v := range eavd.Data {
			childs.Data[k] = v
		}

		q, errx = ox.Q.InsertEAV(detail, childs)
		if errx != nil {
			errx.AddComments("[repository][InsertHeaderEAVTable] while InsertEAV detail")
			return resp, errx
		}

		_, err = tx.Exec(q)
		if err != nil {
			errx = serror.NewFromErrorc(err, "Failed to exec query")
			return resp, errx
		}

		results = append(results, v)
	}

	if auto {
		errx = tx.Admit()
		if errx != nil {
			errx.AddComments("[repository][InsertHeaderEAVTable] while Admit transaction")
			return resp, errx
		}
	}

	return model.WriteResult{
		Items: results,
	}, errx
}

func (ox psql) UpdateHeaderEAVTable(header string, detail string, tx *model.Trx, args model.HeaderDetailTable) (errx serror.SError) {
	var err error

	auto := false
	if tx == nil {
		auto = true
	}

	var (
		q  string
		q2 string
	)

	q, errx = ox.Q.Update(header, args.Header)
	if errx != nil {
		errx.AddComments("[repository][UpdateHeaderEAVTable] while Update")
		return errx
	}

	q2, errx = ox.Q.SetEAV(detail, args.Detail, args.DeleteMetas)
	if errx != nil {
		errx.AddCommentf("[repository][UpdateHeaderEAVTable] while SetEav (detail:%s)", detail)
		return errx
	}

	if auto {
		tx, errx = ox.TrxRepo.Create()
		if errx != nil {
			errx.AddComments("[repository][UpdateHeaderEAVTable] while Create transaction")
			return errx
		}
	}

	defer func() {
		if errStr := recover(); errStr != nil {
			errx = serror.Newc(fmt.Sprintf("%+v", errStr), "Something when wrong")
		}

		if auto && errx != nil && tx != nil {
			errs := tx.Abort()
			if errs != nil {
				errs.AddComments("[repository][UpdateHeaderEAVTable] while Abort transaction")
				log.Error(errs)
			}
		}
	}()

	if q != "" {
		_, err = tx.Exec(q)
		if err != nil {
			errx = serror.NewFromErrorc(err, "Failed to exec query header")
			return errx
		}
	}

	if q2 != "" {
		_, err = tx.Exec(q2)
		if err != nil {
			errx = serror.NewFromErrorc(err, "Failed to exec query detail")
			return errx
		}
	}

	if auto {
		errx = tx.Admit()
		if errx != nil {
			errx.AddComments("[repository][UpdateHeaderEAVTable] while Admit transaction")
			return errx
		}
	}

	return errx
}

func (ox psql) UpdateTable(name string, tx *model.Trx, args sqlq.TableData) (errx serror.SError) {
	var err error

	auto := false
	if tx == nil {
		auto = true
	}

	var q string
	q, errx = ox.Q.Update(name, args)
	if errx != nil {
		errx.AddComments("[repository][UpdateTable] while buid update query")
		return errx
	}

	if auto {
		tx, errx = ox.TrxRepo.Create()
		if errx != nil {
			errx.AddComments("[repository][UpdateTable] while Create transaction")
			return errx
		}
	}

	defer func() {
		if errStr := recover(); errStr != nil {
			errx = serror.Newc(fmt.Sprintf("%+v", errStr), "Something when wrong")
		}

		if auto && errx != nil && tx != nil {
			errs := tx.Abort()
			if errs != nil {
				errs.AddComments("[repository][UpdateTable] while Abort transaction")
				log.Error(errs)
			}
		}
	}()
	_, err = tx.Exec(q)
	if err != nil {
		errx = serror.NewFromErrorc(err, "Failed to exec the query")
		return errx
	}

	if auto {
		errx = tx.Admit()
		if errx != nil {
			errx.AddComments("[repository][UpdateTable] while Admit transaction")
			return errx
		}
	}

	return errx
}

func (ox psql) BatchUpdateTable(name string, tx *model.Trx, args sqlq.TableData) (errx serror.SError) {
	if len(args.Datas) <= 0 {
		return errx
	}

	bindingArgs := map[string]string{
		"!table":      fmt.Sprintf("__:~%s__", name),
		"!columns":    "",
		"!values":     "",
		"!sets":       "",
		"!conditions": "",
	}

	{
		var (
			tb     = ox.Q.TableByKey(name)
			driver = ox.Q.Driver()

			cols  []string
			sets  []string
			vals  []string
			skips []string
			casts = make(map[string]string)
		)

		// conditions
		{
			vals = []string{}
			for k, v := range args.Conditions {
				var (
					opr = sqlq.OperatorEqual
					kx  = k
				)

				switch {
				case strings.HasPrefix(k, "#"):
					casts[utstring.Sub(k, 1, 0)] = utinterface.ToString(v)
					continue
				}

				ks := strings.Split(k, ":")
				if len(ks) > 1 {
					opr = sqlq.ToOperator(ks[1])
					if opr == sqlq.OperatorEmpty {
						opr = sqlq.OperatorEqual
					}
					kx = ks[0]
				}

				if cur, ok := tb.Columns[kx]; ok {
					stx, ok := driver.ToSQLConditionQuery(sqlq.QColumn([]string{cur.Name}), opr, v)
					if !ok {
						errx = serror.Newf("Cannot converting condition %s", kx)
						return errx
					}

					vals = append(vals, stx)
					continue
				}

				errx = serror.Newf("Condition column %s is doesn't exists", k)
				return errx
			}
			bindingArgs["!conditions"] = strings.Join(vals, " AND ")
		}

		// columns
		{
			vals = []string{}
			for _, v := range args.Datas {
				for k := range v {
					var isCustomSet bool

					switch {
					case strings.HasPrefix(k, "!"):
						k = utstring.Sub(k, 1, 0)
						isCustomSet = true
					}

					if utarray.IsExist(k, cols) {
						continue
					}

					if cur, ok := tb.Columns[k]; ok {
						if !isCustomSet {
							vals = append(vals, driver.SQLColumnEscape(cur.Alias))
						}

						// collect sets
						{
							set := fmt.Sprintf(
								`%s = COALESCE("A".%s, %s)`,
								driver.SQLColumnEscape(cur.Name),
								driver.SQLColumnEscape(cur.Alias),
								driver.SQLColumnEscape(cur.Name),
							)
							if isCustomSet {
								set = "@"
							}
							sets = append(sets, set)
						}

						cols = append(cols, k)
						continue
					}

					errx = serror.Newc(fmt.Sprintf("Column %s is doesn't exists", k), "@")
					return errx
				}
			}
			bindingArgs["!columns"] = strings.Join(vals, ", ")
		}

		// values
		{
			vals = []string{}
			for k, v := range args.Datas {
				var subv []string
				for k2, v2 := range cols {
					var (
						onm  = v2
						cset = sets[k2]
					)
					if cset == "@" {
						v2 = fmt.Sprintf("!%s", v2)
					}

					if utarray.IsExist(onm, skips) {
						continue
					}

					cur, ok := driver.ToSQLValueQuery(v[v2])
					if !ok {
						errx = serror.Newf("Cannot converting row %d column %s", k, v2)
						return errx
					}

					if cset == "@" {
						col := tb.Columns[onm]
						sets[k2] = fmt.Sprintf("%s = %s", driver.SQLColumnEscape(col.Name), cur)

						skips = append(skips, onm)
						continue
					}

					if cst, ok := casts[onm]; ok {
						cur = fmt.Sprintf("%s::%s", cur, cst)
					}

					subv = append(subv, cur)
				}

				vals = append(vals, fmt.Sprintf("(%s)", strings.Join(subv, ", ")))
			}
			bindingArgs["!values"] = strings.Join(vals, ", ")
		}

		// sets
		{
			bindingArgs["!sets"] = strings.Join(sets, ", ")
		}
	}

	q := `
		WITH "A"(__!columns__) AS (
			VALUES __!values__
		)
		UPDATE __!table__
		SET
			__!sets__
		FROM "A"
		WHERE
			__!conditions__
	`
	q = u2.Binding(q, bindingArgs)

	q, errx = ox.Q.Select(q, sqlq.Arguments{}, []string{"*"})
	if errx != nil {
		errx.AddComments("[repository][BatchUpdateTable] while buid select query")
		return errx
	}

	auto := false
	if tx == nil {
		auto = true
	}

	var err error
	if auto {
		tx, errx = ox.TrxRepo.Create()
		if errx != nil {
			errx.AddComments("[repository][BatchUpdateTable] while Create transaction")
			return errx
		}
	}

	defer func() {
		if errStr := recover(); errStr != nil {
			errx = serror.Newc(fmt.Sprintf("%+v", errStr), "Something when wrong")
		}

		if auto && errx != nil && tx != nil {
			errs := tx.Abort()
			if errs != nil {
				errs.AddComments("[repository][BatchUpdateTable] while Abort transaction")
				log.Error(errs)
			}
		}
	}()

	_, err = tx.Exec(q)
	if err != nil {
		errx = serror.NewFromErrorc(err, "Failed to exec the query")
		return errx
	}

	if auto {
		errx = tx.Admit()
		if errx != nil {
			errx.AddComments("[repository][BatchUpdateTable] while Admit transaction")
			return errx
		}
	}

	return errx
}

func (ox psql) InsertEAVTable(name string, tx *model.Trx, args sqlq.EAVData) (resp model.WriteResult, errx serror.SError) {
	var err error

	auto := false
	if tx == nil {
		auto = true
	}

	var (
		q       string
		results = []model.WriteItem{}
	)

	q, errx = ox.Q.InsertEAV(name, args)
	if errx != nil {
		errx.AddComments("[repository][InsertEAVTable] while build InsertEAV query")
		return resp, errx
	}

	if auto {
		tx, errx = ox.TrxRepo.Create()
		if errx != nil {
			errx.AddComments("[repository][InsertEAVTable] while Create transaction")
			return resp, errx
		}
	}

	defer func() {
		if errStr := recover(); errStr != nil {
			errx = serror.Newc(fmt.Sprintf("%+v", errStr), "Something when wrong")
		}

		if auto && errx != nil && tx != nil {
			errs := tx.Abort()
			if errs != nil {
				errs.AddComments("[repository][InsertEAVTable] while Abort transaction")
				log.Error(errs)
			}
		}
	}()

	rowx, err := tx.Query(q)
	if err != nil {
		errx = serror.NewFromErrorc(err, "Failed to querying")
		return resp, errx
	}

	defer func() {
		if rowx != nil {
			rowx.Close()
		}
	}()

	var rowb []record
	rowb, errx = ox.ScanRecords(rowx, nil)
	if errx != nil {
		errx.AddComments("[repository][InsertEAVTable] while ScanRecords")
		return resp, errx
	}

	if len(rowb) != len(args.Data) {
		errx = serror.Newc("Length not same", "Something when wrong")
		return resp, errx
	}

	for k, v := range rowb {
		item := model.WriteItem{
			Key:   k,
			ID:    -1,
			Metas: make(model.Metas),
		}
		for k2, v2 := range v.Fields {
			vx := v.Datas[v2]
			if k2 == 0 {
				item.ID = utinterface.ToInt(vx, -1)
			}

			item.Metas[v2] = vx
		}
		results = append(results, item)
	}

	if auto {
		errx = tx.Admit()
		if errx != nil {
			errx.AddComments("[repository][InsertEAVTable] while Admit transaction")
			return resp, errx
		}
	}

	resp = model.WriteResult{
		Items: results,
	}
	return resp, errx
}

func (ox psql) SetEAVTable(name string, tx *model.Trx, args sqlq.EAVData, dels map[string]interface{}) (resp model.WriteResult, errx serror.SError) {
	var err error

	auto := false
	if tx == nil {
		auto = true
	}

	var (
		q       string
		results = []model.WriteItem{}
	)

	q, errx = ox.Q.SetEAV(name, args, dels)
	if errx != nil {
		errx.AddComments("[repository][SetEAVTable] while build SetEAV query")
		return resp, errx
	}

	if auto {
		tx, errx = ox.TrxRepo.Create()
		if errx != nil {
			errx.AddComments("[repository][SetEAVTable] while Create trasaction")
			return resp, errx
		}
	}

	defer func() {
		if errStr := recover(); errStr != nil {
			errx = serror.Newc(fmt.Sprintf("%+v", errStr), "Something when wrong")
		}

		if auto && errx != nil && tx != nil {
			errs := tx.Abort()
			if errs != nil {
				errs.AddComments("[repository][SetEAVTable] while Abort transaction")
				log.Error(errs)
			}
		}
	}()

	rowx, err := tx.Query(q)
	if err != nil {
		errx = serror.NewFromErrorc(err, "Failed to querying")
		return resp, errx
	}

	defer func() {
		if rowx != nil {
			rowx.Close()
		}
	}()

	var rowb []record
	rowb, errx = ox.ScanRecords(rowx, nil)
	if errx != nil {
		errx.AddComments("[repository][SetEAVTable] while ScanRecords")
		return resp, errx
	}

	if len(rowb) != len(args.Data) {
		errx = serror.Newc("Length not same", "Something when wrong")
		return resp, errx
	}

	for k, v := range rowb {
		item := model.WriteItem{
			Key:   k,
			ID:    -1,
			Metas: make(model.Metas),
		}
		for k2, v2 := range v.Fields {
			vx := v.Datas[v2]
			if k2 == 0 {
				item.ID = utinterface.ToInt(vx, -1)
			}

			item.Metas[v2] = vx
		}
		results = append(results, item)
	}

	if auto {
		errx = tx.Admit()
		if errx != nil {
			errx.AddComments("[repository][SetEAVTable] while Admit transaction")
			return resp, errx
		}
	}

	resp = model.WriteResult{
		Items: results,
	}
	return resp, errx
}

func (ox record) CopyToStruct(obj interface{}) (errx serror.SError) {
	if obj == nil {
		errx = serror.Newc("Object cannot be null", "@")
		return errx
	}

	if !structs.IsStruct(obj) {
		errx = serror.Newc("Object is not struct", "@")
		return errx
	}

	var (
		nms    = make(map[string]int)
		fields = structs.Fields(obj)
	)

	for k, v := range fields {
		nms[utstring.Chains(v.Tag("key"), v.Tag("json"), v.Name())] = k
	}

	setValue := func(idx int, val interface{}) {
		var err error
		defer func() {
			if err != nil {
				errx = serror.NewFromErrorc(err, "Failed to set value")
				return
			}

			if errRcv := recover(); errRcv != nil {
				errx = serror.Newc(fmt.Sprintf("%+v", errRcv), "Something when wrong")
				return
			}
		}()

		var (
			origin = fields[idx].Value()
			model  = reflect.TypeOf((*ICopyToStruct)(nil)).Elem()
		)

		if reflect.TypeOf(origin).Implements(model) {
			val, err = origin.(ICopyToStruct).Cast(val)
		}

		err = fields[idx].Set(val)
	}

	for k, v := range ox.Datas {
		if idx, ok := nms[k]; ok {
			setValue(idx, v)
			if errx != nil {
				errx.AddCommentf("[repository][CopyToStruct] while setValue %s", k)
				return errx
			}
		}
	}

	return errx
}
