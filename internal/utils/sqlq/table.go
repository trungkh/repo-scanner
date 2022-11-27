package sqlq

import (
	"fmt"
	"strings"

	"github.com/napsy/go-css"

	"repo-scanner/internal/utils/serror"
	"repo-scanner/internal/utils/utstring"
	"repo-scanner/internal/utils/utstruct"
)

type Table struct {
	Schema       string
	Table        string
	Primary      string
	SoftDelete   string
	IsEAV        bool
	ConditionMap OperatorsMap
	Columns      map[string]*Column
}

type Column struct {
	IsPrimary    bool
	IsQuery      bool
	IsSoftDelete bool
	EAVRole      EAVRole
	Name         string
	Alias        string
	Most         bool
	Sortable     bool
	Condition    *Condition
}

type Condition struct {
	AllowOperator []Operator
	Default       Operator
}

func (ox *Column) SetName(nm string) {
	ox.Name = nm
}

func (ox *Column) SetAlias(al string) {
	ox.Alias = al
}

func (ox *Column) SetPrimary() {
	ox.IsPrimary = true
}

func (ox *Column) SetSoftDelete() {
	ox.IsSoftDelete = true
}

func (ox *Column) SetSortable() {
	ox.Sortable = true
}

func (ox *Column) SetQuery(q string) {
	ox.Name = q
	ox.IsQuery = true
}

func (ox *Column) SetMost() {
	ox.Most = true
}

func (ox *Column) SetEAVRole(er string) (errx serror.SError) {
	eavrl := map[string]EAVRole{
		"key":         EAVRoleKey,
		"group":       EAVRoleGroup,
		"description": EAVRoleDescription,
		"value":       EAVRoleValue,
	}

	var ok bool
	ox.EAVRole, ok = eavrl[er]
	if !ok {
		errx = serror.Newc(fmt.Sprintf("Unknown EAV role for '%s'", er), "@")
		return errx
	}

	return errx
}

func (ox *Column) SetCondition(cds string, cmap OperatorsMap) (errx serror.SError) {
	var def Operator
	allow := []Operator{}

	conds := utstring.CleanSpit(cds, ",")

	nconds := []string{}
	for _, v := range conds {
		if strings.HasPrefix(v, "$") {
			if cmap != nil {
				if lst, ok := cmap[utstring.Sub(v, 1, 0)]; ok {
					for _, v2 := range lst {
						nconds = append(nconds, string(v2))
					}
				}
			}
			continue
		}

		nconds = append(nconds, v)
	}

	for k, v := range nconds {
		if k == 0 {
			def = ToOperator(v)
		}
		allow = append(allow, ToOperator(v))
	}

	ox.Condition = &Condition{
		Default:       def,
		AllowOperator: allow,
	}

	return errx
}

func (ox *Table) LoadFromStruct(obj interface{}) (errx serror.SError) {
	metas := utstruct.GetMetas(obj)
	if len(metas) > 0 {
		for _, v := range metas {
			if !v.IsExported() || v.Tag("sqlq") == "-" {
				continue
			}

			col := Column{
				IsPrimary: false,
				IsQuery:   false,
				EAVRole:   EAVRoleNone,
				Most:      false,
				Sortable:  false,
				Condition: nil,
			}

			if syntx := utstring.Trim(v.Tag("sqlq")); syntx != "" {
				dats, err := css.Unmarshal([]byte(syntx))
				if err != nil {
					errx = serror.NewFromErrorc(err, "Failed to read struct meta for 'sqlq' field")
					return errx
				}

				if len(dats) <= 0 || len(dats) > 1 {
					errx = serror.Newf("Invalid meta 'sqlq' on field %s", v.Name())
					return errx
				}

				for kx, vx := range dats {
					name := utstring.Trim(string(kx))
					if name == "@" {
						name = utstring.Chains(v.Tag("key"), v.Tag("json"))
					}

					if name == "" {
						errx = serror.Newc("Cannot empty field name", "@")
						return errx
					}

					col.SetAlias(name)
					col.SetName(utstring.Chains(vx["key"], vx["db"], v.Tag("db"), strings.ToLower(name)))

					if vx["prime"] == "true" || vx["primary"] == "true" {
						ox.Primary = strings.ToLower(name)
						col.SetPrimary()
					}

					if vx["sdel"] == "true" || vx["soft-del"] == "true" {
						ox.SoftDelete = strings.ToLower(name)
						col.SetSoftDelete()
					}

					if qx := utstring.Trim(vx["query"]); qx != "" {
						col.SetQuery(qx)
					}

					if vx["most"] == "true" {
						col.SetMost()
					}

					if vx["sort"] == "true" || vx["sortable"] == "true" {
						col.SetSortable()
					}

					if eavr := utstring.Trim(vx["eav"]); eavr != "" {
						errx = col.SetEAVRole(eavr)
						if errx != nil {
							errx.AddCommentf("while set EAV role, field '%s'", name)
							return errx
						}
					}

					if conds := utstring.Trim(vx["conds"]); conds != "" {
						errx = col.SetCondition(conds, ox.ConditionMap)
						if errx != nil {
							errx.AddCommentf("while set condition, field '%s'", name)
							return errx
						}
					}

					ox.AddColumn(strings.ToLower(name), col)
					break
				}
				continue
			}

			name := utstring.Chains(v.Tag("json"), v.Name())
			key := utstring.Chains(v.Tag("tkey"), v.Tag("db"), strings.ToLower(name))

			col.SetAlias(name)
			col.SetName(key)

			if v.Tag("tprime") == "1" || (strings.ToLower(key) == "id" && ox.Primary == "") {
				ox.Primary = strings.ToLower(name)
				col.SetPrimary()
			}

			if v.Tag("tdel") == "1" {
				ox.SoftDelete = strings.ToLower(name)
				col.SetSoftDelete()
			}

			if q := utstring.Trim(v.Tag("tquery")); q != "" {
				col.SetQuery(q)
			}

			if v.Tag("tmost") == "1" {
				col.SetMost()
			}

			if v.Tag("tsort") == "1" {
				col.SetSortable()
			}

			if eavr := utstring.Trim(v.Tag("teav")); eavr != "" {
				errx = col.SetEAVRole(eavr)
				if errx != nil {
					errx.AddCommentf("while set EAV role, field '%s'", name)
					return errx
				}
			}

			if conds := utstring.Trim(v.Tag("tconds")); conds != "" {
				errx = col.SetCondition(strings.Join(utstring.CleanSpit(conds, ":"), ","), ox.ConditionMap)
				if errx != nil {
					errx.AddCommentf("while set condition, field '%s'", name)
					return errx
				}
			}

			ox.AddColumn(strings.ToLower(name), col)
		}
	}
	return errx
}

// TODO: still draft
// func (ox *Table) LoadFromDB(db *sqlx.DB) (errx serror.SError) {
// 	return serror.New("Function not yet supported")
// }

func (ox *Table) AddColumn(key string, dat Column) {
	if ox.Columns == nil {
		ox.Columns = make(map[string]*Column)
	}
	ox.Columns[key] = &dat
}

func (ox Table) GetColumnByEAVRole(role EAVRole) string {
	if ox.IsEAV {
		for k, v := range ox.Columns {
			if v.EAVRole == role {
				return k
			}
		}
	}

	return ""
}

func (ox Table) GetAvailableEAVColumns() (out map[EAVRole]*Column) {
	out = make(map[EAVRole]*Column)
	for _, v := range ox.Columns {
		switch v.EAVRole {
		case EAVRoleKey:
			fallthrough
		case EAVRoleGroup:
			fallthrough
		case EAVRoleDescription:
			fallthrough
		case EAVRoleValue:
			if _, ok := out[v.EAVRole]; !ok {
				out[v.EAVRole] = v
			}
		}
	}
	return out
}
