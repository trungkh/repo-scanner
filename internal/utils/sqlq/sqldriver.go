package sqlq

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/gearintellix/u2"
	"github.com/lib/pq"

	log "github.com/sirupsen/logrus"

	"repo-scanner/internal/utils/utinterface"
	"repo-scanner/internal/utils/utstring"
	"repo-scanner/internal/utils/uttime"
)

// QColumn type
type QColumn []string

// QRaw type
type QRaw string

// QArray type
type QArray []interface{}

// QCast type
type QCast []interface{}

// SQLDriver type
type SQLDriver int

const (
	// DriverMySQL driver
	DriverMySQL SQLDriver = 1 + iota

	// DriverPostgreSQL driver
	DriverPostgreSQL

	// DriverBigQuery driver
	DriverBigQuery
)

type ConditionObj struct {
	Condition1 string
	Operator   string
	Condition2 string
}

func (QArray) FromStrings(v []string) QArray {
	var res QArray
	if len(v) > 0 {
		res = QArray{}
		for _, v := range v {
			res = append(res, v)
		}
	}
	return res
}

func (QArray) FromInt64s(v []int64) QArray {
	var res QArray
	if len(v) > 0 {
		res = QArray{}
		for _, v := range v {
			res = append(res, v)
		}
	}
	return res
}

func (QCast) Casting(val interface{}, cst string) QCast {
	return QCast{val, cst}
}

func (ox QCast) GetValue() interface{} {
	if len(ox) >= 0 {
		return ox[0]
	}

	return nil
}

func (ox QCast) GetType() string {
	if len(ox) >= 1 {
		return fmt.Sprintf("::%s", utinterface.ToString(ox[1]))
	}

	return ""
}

func (ox ConditionObj) Syntax() string {
	return fmt.Sprintf("%s %s %s", ox.Condition1, ox.Operator, ox.Condition2)
}

// ToSQLValueQuery function
func (ox SQLDriver) ToSQLValueQuery(value interface{}) (res string, ok bool) {
	opts := map[string]string{
		"null":       "NULL",
		"true":       "true",
		"false":      "false",
		"dateFormat": "Y-m-d H:i:s",
	}

	switch ox {
	case DriverPostgreSQL:
		opts["true"] = "TRUE"
		opts["false"] = "FALSE"
	}

	if value == nil {
		res, ok = opts["null"], true
		return res, ok
	}

	val := reflect.ValueOf(value)
	if val.Kind() == reflect.Ptr {
		valx := val.Elem()
		if valx.Kind() == reflect.Invalid {
			res, ok = opts["null"], true
			return res, ok
		}

		value = valx.Interface()
		if value == nil {
			res, ok = opts["null"], true
			return res, ok
		}

		val = reflect.ValueOf(value)
	}

	if value == nil {
		res, ok = opts["null"], true
		return res, ok
	}

	res, ok = "", false

	switch valx := value.(type) {
	case QRaw:
		res, ok = val.String(), true

	case QArray:
		resx, err := pq.Array(value).Value()
		if err != nil {
			log.Error(err, "Failed to parsing array")
			return res, ok
		}

		res, ok = ox.SQLValueEscape(fmt.Sprintf("%s", resx)), true

	case QCast:
		if res, ok = ox.ToSQLValueQuery(valx.GetValue()); !ok {
			return res, ok
		}
		res += valx.GetType()

	case QColumn:
		val1 := []string{}
		for i := 0; i < val.Len(); i++ {
			val1 = append(val1, ox.SQLColumnEscape(val.Index(i).String()))
		}
		res, ok = strings.Join(val1, "."), true

	default:
		switch val.Kind() {
		case reflect.Invalid:
			res, ok = opts["null"], true

		case reflect.String:
			res, ok = ox.SQLValueEscape(fmt.Sprintf("%s", value)), true

		case reflect.Bool:
			res, ok = opts["false"], true
			if value.(bool) {
				res = opts["true"]
			}

		case reflect.Map:
			arrs := []string{}
			for _, key := range val.MapKeys() {
				val2, ok2 := ox.ToSQLValueQuery(val.MapIndex(key).Interface())
				if ok2 {
					arrs = append(arrs, val2)
				}
			}
			res, ok = strings.Join(arrs, ", "), true

		case reflect.Slice:
			arrs := []string{}
			for i := 0; i < val.Len(); i++ {
				val2, ok2 := ox.ToSQLValueQuery(val.Index(i).Interface())
				if ok2 {
					arrs = append(arrs, val2)
				}
			}
			res, ok = strings.Join(arrs, ", "), true

		default:
			typ := reflect.TypeOf(value).String()
			ctype := map[string][]string{
				"intx":   []string{"int", "int8", "int16", "int32", "int64"},
				"uintx":  []string{"uint", "uint8", "uint16", "uint32", "uint64"},
				"floatx": []string{"float32", "float64"},
			}

			done := false
			for k2, v2 := range ctype {
				for _, v3 := range v2 {
					if typ == v3 {
						done = true
						typ = k2
						break
					}
				}

				if done {
					break
				}
			}

			switch typ {
			case "intx":
				res, ok = utstring.Int64ToString(val.Int()), true

			case "uintx":
				res, ok = utstring.Uint64ToString(val.Uint()), true

			case "floatx":
				res, ok = utstring.FloatToString(val.Float()), true

			case "time.Time":
				res, ok = ox.SQLValueEscape(uttime.Format(uttime.DefaultDateTimeFormat, value.(time.Time))), true
			}
		}
	}

	return res, ok
}

// SQLValueEscape function
func (ox SQLDriver) SQLValueEscape(value string) (res string) {
	qoute := "'"

	switch ox {
	default:
		res = qoute + strings.ReplaceAll(value, qoute, strings.Repeat(qoute, 2)) + qoute
	}

	return res
}

// SQLColumnEscape function
func (ox SQLDriver) SQLColumnEscape(name string) (res string) {
	qoute := "\""

	switch ox {
	default:
		res = qoute + strings.ReplaceAll(name, qoute, strings.Repeat(qoute, 2)) + qoute
	}

	return res
}

// ToSQLOperatorQuery function
func (ox SQLDriver) ToSQLOperatorQuery(opr Operator) (res string, ok bool) {
	val := string(opr)

	switch ox {
	default:
		ok = true
		switch opr {
		case OperatorAll:
			fallthrough
		case OperatorEmpty:
			return res, false
		}
	}

	res = val
	return res, ok
}

func (ox SQLDriver) ToSQLConditionObject(cond1 interface{}, opr Operator, cond2 interface{}) (res ConditionObj, ok bool) {
	v, ok := "", true
	if v, ok = ox.ToSQLValueQuery(cond1); !ok {
		return res, false
	}
	res.Condition1 = v

	if v, ok = ox.ToSQLOperatorQuery(opr); !ok {
		return res, false
	}
	res.Operator = v

	ok = true
	switch opr {
	case OperatorBetween:
		ok = false
		ovals := reflect.ValueOf(cond2)
		if ovals.Kind().String() == "slice" {
			if ovals.Len() >= 2 {
				val1, val2 := "", ""

				var ok2 bool
				if val1, ok2 = ox.ToSQLValueQuery(ovals.Index(0).Interface()); !ok2 {
					return res, false
				}

				if val2, ok2 = ox.ToSQLValueQuery(ovals.Index(1).Interface()); !ok2 {
					return res, false
				}

				v = fmt.Sprintf("%s AND %s", val1, val2)
				ok = true
			}
		}
		if !ok {
			return ConditionObj{
				Condition1: "TRUE",
				Operator:   "",
				Condition2: "",
			}, true
		}

	case OperatorIn:
		fallthrough
	case OperatorNotIn:
		val1 := ""
		if val1, ok = ox.ToSQLValueQuery(cond2); !ok {
			return res, false
		}

		if ok {
			v = fmt.Sprintf("(%s)", val1)
		}

	case OperatorIsNotNull:
		fallthrough
	case OperatorIsNull:
		v = ""

	default:
		if v, ok = ox.ToSQLValueQuery(cond2); !ok {
			return res, false
		}
	}
	res.Condition2 = v

	return res, true
}

// ToSQLConditionQuery function
func (ox SQLDriver) ToSQLConditionQuery(cond1 interface{}, opr Operator, cond2 interface{}) (res string, ok bool) {
	obj, ok := ox.ToSQLConditionObject(cond1, opr, cond2)
	if ok {
		res = obj.Syntax()
	}

	return res, ok
}

// U2SQLBinding function
func (ox SQLDriver) U2SQLBinding(query string, values map[string]interface{}) (res string) {
	opts := map[string]string{
		"true":       "true",
		"false":      "false",
		"dateFormat": "Y-m-d H:i:s",
	}

	switch ox {
	case DriverPostgreSQL:
		opts["true"] = "'t'"
		opts["false"] = "'f'"
	}

	vals := make(map[string]string)
	for k, dd := range values {
		dv, ok := ox.ToSQLValueQuery(dd)
		if ok {
			vals[k] = dv
		}
	}

	res = u2.Binding(query, vals)
	return
}
