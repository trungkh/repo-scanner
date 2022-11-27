package utsql

import (
	"database/sql"
	"encoding/json"
	"strconv"

	"repo-scanner/internal/utils/serror"
	"repo-scanner/internal/utils/utinterface"
	"repo-scanner/internal/utils/uttime"
)

func ScanToArrayInterface(qo *sql.Rows, row []interface{}) (errx serror.SError) {
	valuePtrs := make([]interface{}, len(row))
	for k := range row {
		valuePtrs[k] = &row[k]
	}

	err := qo.Scan(valuePtrs...)
	if err != nil {
		errx = serror.NewFromErrorc(err, "while scan rows")
		return errx
	}

	return errx
}

func PostgresRowTypeSolve(typs []*sql.ColumnType, dst *[]interface{}) (errx serror.SError) {
	var err error

	if len(typs) != len(*dst) {
		errx = serror.Newc("Column not match with row destination", "while postgres_row_type_solve")
		return errx
	}

	for k, v := range *dst {
		typ := typs[k]

		if utinterface.IsNil(v) {
			continue
		}

		var cur interface{}
		ok := false

		switch typ.DatabaseTypeName() {
		case "NUMERIC", "DECIMAL":
			cur, err = strconv.ParseFloat(string(v.([]uint8)), 64)
			if err != nil {
				errx = serror.NewFromErrorc(err, "[utils][utsql][PostgresRowTypeSolve] while cast string to float64")
				return errx
			}
			ok = true

		case "BIT", "UUID":
			cur = string(v.([]uint8))
			ok = true

		case "TIMESTAMP", "DATE":
			cur, errx = uttime.ParseForceTimezone(v, "@")
			if errx != nil {
				return errx
			}
			ok = true

		case "JSONB":
			cur = nil
			err = json.Unmarshal([]byte(v.([]uint8)), &cur)
			if err != nil {
				errx = serror.NewFromErrorc(err, "[utils][utsql][PostgresRowTypeSolve] while cast bytes to interface")
				return errx
			}
			ok = true
		}

		if ok {
			(*dst)[k] = cur
		}
	}
	return errx
}
