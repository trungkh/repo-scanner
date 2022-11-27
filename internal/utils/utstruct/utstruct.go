package utstruct

import (
	"reflect"
	"strings"

	"github.com/gearintellix/structs"
	//"github.com/luthfikw/structs"

	"repo-scanner/internal/utils/serror"
)

// GetMetas from struct
func GetMetas(obj interface{}) (res []*structs.Field) {
	objx := reflect.ValueOf(obj)
	for objx.Kind() == reflect.Ptr {
		objx = objx.Elem()
	}

	if objx.Kind() != reflect.Struct {
		return nil
	}

	ls := structs.Fields(objx.Interface())

	for _, v := range ls {
		if v.IsExported() {
			res = append(res, v)
		}
	}
	return
}

func CastToMap(obj interface{}) (res map[string]interface{}, errx serror.SError) {
	res = make(map[string]interface{})

	metas := GetMetas(obj)
	for _, v := range metas {
		key := v.Name()

		if cur := v.Tag("json"); cur != "" {
			curs := strings.Split(cur, ",")
			for k1, v2 := range curs {
				v2 = strings.TrimSpace(v2)
				switch k1 {
				case 0:
					key = v2

				default:
					switch v2 {
					case "omitempty":
						if v.IsZero() {
							key = "-"
							continue
						}
					}
				}
			}

		} else if cur := v.Tag("key"); cur != "" {
			key = strings.TrimSpace(cur)
		}

		if key != "-" {
			res[key] = v.Value()
		}
	}

	return res, errx
}
