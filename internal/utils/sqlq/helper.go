package sqlq

import (
	"fmt"
	"reflect"

	"repo-scanner/internal/utils/serror"
	"repo-scanner/internal/utils/utinterface"
)

func (ox *EAVData) ApplyFromNestedJSON(data map[string]interface{}) (errx serror.SError) {
	ox.Meta = make(map[string]interface{})
	ox.Data = make(map[string]EAVItem)

	for k, v := range data {
		v2 := reflect.ValueOf(v)

		switch v2.Kind() {
		case reflect.Map:
			for _, k3 := range v2.MapKeys() {
				v3 := v2.MapIndex(k3)

				if v3.IsNil() {
					continue
				}

				cur := EAVItem{}

				switch v3.Type().Kind() {
				case reflect.Map:
					v4 := v2.MapIndex(reflect.ValueOf("id"))
					if !v4.IsNil() {
						cur.ID = utinterface.ToInt(v4.Interface(), 0)
					}

					v4 = v2.MapIndex(reflect.ValueOf("value"))
					if !v4.IsNil() {
						cur.Value = utinterface.ToString(v4.Interface())
					}

					// TODO: add metas from other attributes

				default:
					cur.Value = utinterface.ToString(v3.Interface())
				}
				ox.Data[fmt.Sprintf("%s.%s", k, k3.String())] = cur
			}

		default:
			ox.Meta[k] = v
		}
	}

	return nil
}
