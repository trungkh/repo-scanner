package sqlq

import (
	"fmt"
	"strings"

	"github.com/gearintellix/serr"
	"github.com/gearintellix/u2"

	"repo-scanner/internal/utils/serror"
	"repo-scanner/internal/utils/utarray"
	"repo-scanner/internal/utils/utint"
	"repo-scanner/internal/utils/utstring"

	log "github.com/sirupsen/logrus"
)

// Building query Select statements
// Available bindings:
//  - [:~] > Table binding path only
//  - [:@] > Table binding path and alias
//  - [:=] > Column binding name only
//  - [::] > Column binding full path
func (ox sqlQuery) Select(q string, args Arguments, allow []string) (qo string, errx serror.SError) {
	var erry serr.SErr

	defer func() {
		if r := recover(); r != nil {
			log.Errorf("Unexpected error has occured on sqlq select, details: %v", r)
		}
	}()

	if len(allow) <= 0 {
		allow = []string{"*"}
	}
	if len(args.Fields) <= 0 {
		args.Fields = []string{"*"}
	}

	var items []u2.TagInfo

	// TODO: here still under development
	// // #region[rgba(255, 82, 116, 0.05)] > reading virtuals
	// type lVirtual struct {
	// 	Index     int
	// 	Name      string
	// 	Element   string
	// 	Operator  Operator
	// 	Statement string
	// 	Body      string
	// }

	// virtuals := []lVirtual{}
	// q, items, erry = u2.ScanTags(q, "virtual")
	// if erry != nil {
	// 	errx = serror.NewFromSErrc(erry, "while scan virtual tags")
	// 	return qo, errx
	// }

	// replacer := map[string]string{}
	// for k, v := range items {
	// 	oprt, stmt := OperatorIsNotNull, ""
	// 	if v.Meta != nil {
	// 		if tmp, ok := v.Meta["operator"]; ok {
	// 			oprt = ToOperator(tmp)
	// 			if oprt == OperatorEmpty {
	// 				errx = serror.Newc(fmt.Sprintf("Cannot use empty operator, on tag virtual:%s", v.Key), "while check operator")
	// 				return qo, errx
	// 			}
	// 		}

	// 		if tmp, ok := v.Meta["statement"]; ok {
	// 			stmt = tmp
	// 		}
	// 	}

	// 	virtuals = append(virtuals, lVirtual{
	// 		Index:    k,
	// 		Name:     v.Key,
	// 		Element:  v.Index,
	// 		Operator: OperatorIsNotNull,
	// 		Body:     v.Value,
	// 	})

	// 	groups = append(groups, lColumn{
	// 		Key:    v.Key,
	// 		Syntax: utstring.Chains(utstring.Trim(v.Value), v.Index),
	// 	})
	// }
	// // #endregion

	type lAlias struct {
		Index int
		Name  string
		Type  string
		IsEAV bool
		Allow []string
		Most  bool
	}

	var (
		alias   = make(map[string]lAlias)
		colPars = make(map[string]string)
	)

	// // #region[rgba(255, 82, 116, 0.05)] > prepare tables
	// for k, v := range ox.tables {
	// 	path := fmt.Sprintf("%s.%s", ox.driver.SQLColumnEscape(v.Schema), ox.driver.SQLColumnEscape(v.Table))

	// 	colPars[fmt.Sprintf(":@%s", k)] = fmt.Sprintf("%s AS %s", path, k)
	// 	colPars[fmt.Sprintf(":~%s", k)] = path
	// }
	// // #endregion

	// #region[rgba(255, 82, 116, 0.05)] > reading foreign objects
	{
		var (
			xBinds = make(map[string]string)
			prfxs  = u2.ScanPrefix(q, []string{
				":@",
				":~",
			})
		)
		for k, v := range prfxs {
			for _, key := range v {
				switch k {
				case ":@", ":~":
					ctb := ox.TableByKey(key)
					if ctb == nil {
						continue
					}

					if _, ok := alias[strings.ToLower(key)]; !ok {
						alias[strings.ToLower(key)] = lAlias{
							Index: -1,
							Name:  key,
							Type:  "foreign",
							Allow: []string{"@"},
							IsEAV: ctb.IsEAV,
							Most:  true,
						}
					}
				}
			}
		}

		if len(xBinds) > 0 {
			q = u2.Binding(q, xBinds)
		}
	}
	// #endregion

	// #region[rgba(255, 82, 116, 0.05)] > reading tables
	q, items, erry = u2.ScanTags(q, "tb")
	if erry != nil {
		errx = serror.NewFromSErrc(erry, "while scan tb tags")
		return qo, errx
	}

	for k, v := range items {
		cur := lAlias{
			Index: k,
			Name:  utstring.Chains(v.Index, v.Key),
			Type:  "table",
			Allow: []string{"@"},
			Most:  true,
		}

		tb, ok := ox.tables[cur.Name]
		if !ok {
			errx = serror.Newc(fmt.Sprintf("Cannot resolve table %s on tag %s", cur.Name, v.Tag+"."+v.Key), "while sqlq_select")
			return qo, errx
		}

		cur.IsEAV = tb.IsEAV

		if allow, ok := v.Meta["allow"]; ok {
			cur.Allow = utstring.CleanSpit(allow, ",")
		}

		alias[strings.ToLower(v.Key)] = cur
	}
	// #endregion

	// #region[rgba(255, 82, 116, 0.05)] > reading joins
	type lJoin struct {
		Key       string
		Type      string
		Name      string
		Field     []string
		Group     []string
		Condition string
		Syntax    string
		Options   map[string]string
	}

	joins := []lJoin{}

	q, items, erry = u2.ScanTags(q, "join")
	if erry != nil {
		errx = serror.NewFromSErrc(erry, "while scan join tags")
		return qo, errx
	}

	for k, v := range items {
		cur := lAlias{
			Index: k,
			Name:  utstring.Chains(v.Index, v.Key),
			Type:  "join",
			Allow: []string{"@"},
		}

		if mst := v.Meta["most"]; mst == "1" {
			cur.Most = true
		}

		tb, ok := ox.tables[cur.Name]
		if !ok {
			errx = serror.Newc(fmt.Sprintf("Cannot resolve table %s on tag %s", cur.Name, v.Tag+"."+v.Key), "while sqlq_select")
			return qo, errx
		}

		cur.IsEAV = tb.IsEAV

		join := lJoin{
			Key:       strings.ToLower(v.Key),
			Name:      cur.Name,
			Type:      "normal",
			Field:     []string{},
			Group:     []string{},
			Condition: utstring.Chains(utstring.Trim(v.Meta["cond"]), "TRUE"),
			Syntax:    utstring.Trim(v.Value),
			Options:   v.Meta,
		}

		if typ, ok := v.Meta["type"]; ok {
			if !utarray.IsExist(typ, []string{"normal", "eav"}) {
				errx = serror.Newc(fmt.Sprintf("Unknown type %s on %s:%s", typ, v.Tag, v.Key), "while reading_joins_on_sqlq_select")
				return qo, errx
			}

			join.Type = typ
		}

		if grp, ok := v.Meta["group"]; ok {
			grps := utstring.CleanSpit(grp, ",")
			for _, v2 := range grps {
				v3 := v2
				if strings.HasPrefix(v2, "!") {
					v3 = utstring.Sub(v3, 1, 0)
					join.Field = append(join.Field, v3)
				}
				join.Group = append(join.Group, v3)
			}
		}

		if join.Type == "eav" && !cur.IsEAV {
			errx = serror.Newc(fmt.Sprintf("Table %s is not support EAV on %s:%s", cur.Name, v.Tag, v.Key), "while reading_joins_on_sqlq_select")
			return qo, errx
		}

		if allow, ok := v.Meta["allow"]; ok {
			cur.Allow = utstring.CleanSpit(allow, ",")
		}

		alias[strings.ToLower(v.Key)] = cur
		joins = append(joins, join)
	}
	// #endregion

	// #region[rgba(255, 82, 116, 0.05)] > reading conditions
	type lCondition struct {
		Key           string
		Name          string
		Syntax        string
		DefaultSyntax string
		IsReferenced  bool
		Most          bool
		Allow         []Operator
		Value         interface{}
		Depends       []string
	}

	conds := []lCondition{}

	q, items, erry = u2.ScanTags(q, "cond")
	if erry != nil {
		errx = serror.NewFromSErrc(erry, "while scan cond tags")
		return qo, errx
	}

	for _, v := range items {
		cur := lCondition{
			Key:           utstring.Trim(v.Key),
			Name:          utstring.Trim(utstring.Chains(v.Index, v.Key)),
			DefaultSyntax: "TRUE",
			IsReferenced:  true,
			Syntax:        "__syntax__",
		}

		if cur.Name == "@" {
			cur.IsReferenced = false
		}

		if v.Value != "" {
			cur.Syntax = utstring.Trim(v.Value)
		}

		if mst := v.Meta["most"]; mst == "1" {
			cur.Most = true
		}

		if alw := v.Meta["allow"]; alw != "" {
			alws := utstring.CleanSpit(alw, ",")
			for _, v2 := range alws {
				opr := ToOperator(v2)

				if opr == OperatorEmpty {
					errx = serror.Newc(fmt.Sprintf("Unknown operator %s on tag %s:%s", v2, v.Tag, v.Key), "while reading_conditions_on_sqlq_select")
					return qo, errx
				}

				cur.Allow = append(cur.Allow, opr)
			}
		}

		if dpn := v.Meta["depends"]; dpn != "" {
			for _, v2 := range utstring.CleanSpit(dpn, ",") {
				v2 = strings.ToLower(v2)
				cur.Depends = append(cur.Depends, v2)

				if _, ok := alias[v2]; !ok {
					errx = serror.Newf("Cannot find object '%s' not exists on tag %s:%s", v2, v.Tag, v.Key)
					return qo, errx
				}
			}
		}

		if def := v.Meta["default"]; def != "" {
			def = utstring.Trim(def)
			cur.DefaultSyntax = ox.driver.SQLValueEscape(def)

			if strings.HasPrefix(def, "!") {
				def = utstring.Sub(def, 1, 0)
				if strings.HasPrefix(def, "!") {
					def = ox.driver.SQLValueEscape(def)
				}
				cur.DefaultSyntax = def
			}
		}

		conds = append(conds, cur)
	}
	// #endregion

	// #region[rgba(255, 82, 116, 0.05)] > reading views
	type lColumn struct {
		Key          string
		Name         string
		Syntax       string
		IsReferenced bool
		NoAlias      bool
		Most         bool
	}

	type lField struct {
		Name         string
		Most         []string
		Allow        []string
		Syntax       map[string]lColumn
		IsReferenced bool
		Raw          string
	}

	type lView struct {
		OriginKey string
		Type      string
		Field     lField
		Column    lColumn
		Params    map[string][]string
		Aliases   map[string]string
	}

	views := []lView{}

	q, items, erry = u2.ScanTags(q, "view")
	if erry != nil {
		errx = serror.NewFromSErrc(erry, "while scan view tags")
		return qo, errx
	}

	for _, v := range items {
		cur := lView{
			OriginKey: v.Key,
		}

		if strings.HasPrefix(v.Key, ":") {
			cur.Type = "field"
			obj := lField{
				Name:         utstring.Trim(v.Key),
				Most:         []string{"@"},
				Allow:        []string{},
				Syntax:       make(map[string]lColumn),
				Raw:          utstring.Trim(v.Value),
				IsReferenced: false,
			}

			if len(obj.Name) > 1 {
				obj.Name = utstring.Sub(obj.Name, 1, 0)
			}

			if obj.Name == "@" {
				obj.IsReferenced = true
			}

			if !utarray.IsExist(obj.Name, []string{"@", ":"}) {
				if _, ok := alias[obj.Name]; !ok {
					errx = serror.Newc(fmt.Sprintf("Cannot find table %s on tag %s:%s", obj.Name, v.Tag, v.Key), "while reading_views_on_sqlq_select")
					return qo, errx
				}
				obj.Name = strings.ToLower(obj.Name)
				obj.IsReferenced = true
			}

			if v.Index != "" {
				indexs := utstring.CleanSpit(v.Index, ",")
				for _, v2 := range indexs {
					if strings.Contains(v2, "!") {
						obj.Most = append(obj.Most, v2)
						continue
					}
					obj.Allow = append(obj.Allow, v2)
				}
			}

			for k2, v2 := range v.Meta {
				k3 := k2
				syntx := lColumn{
					Syntax: v2,
				}

				if arrowIdx := strings.LastIndex(k3, ">"); arrowIdx > 0 {
					syntx.Key = utstring.Trim(utstring.Sub(k3, arrowIdx+1, 0))
					k3 = utstring.Trim(utstring.Sub(k3, 0, arrowIdx))
				}

				if adIdx := strings.LastIndex(k3, "@"); adIdx > 0 {
					syntx.Name = utstring.Trim(utstring.Sub(k3, adIdx+1, 0))
					k3 = utstring.Trim(utstring.Sub(k3, 0, adIdx))
				}

				if strings.HasPrefix(k3, "!") {
					k3 = utstring.Sub(k3, 1, 0)
					obj.Most = append(obj.Most, fmt.Sprintf("!%s", k3))
				}

				if utarray.CheckAllowedLayer([]string{k3}, [][]string{obj.Most, obj.Allow}) {
					if syntx.Syntax == "" {
						syntx.Syntax = ":@"
					}

					if strings.HasPrefix(syntx.Syntax, ":") {
						syntx.Syntax = utstring.Sub(syntx.Syntax, 1, 0)
						syntx.Syntax = strings.ToLower(syntx.Syntax)
						syntx.IsReferenced = true
					}

					if strings.HasSuffix(syntx.Syntax, "!") {
						syntx.Syntax = utstring.Sub(syntx.Syntax, 0, -1)
						syntx.NoAlias = true
					}

					if syntx.Name == "" {
						syntx.Name = k3
					}
					obj.Syntax[k3] = syntx
				}
			}

			cur.Field = obj

		} else {
			cur.Type = "column"
			obj := lColumn{
				Key:    utstring.Trim(v.Key),
				Name:   utstring.Trim(v.Index),
				Syntax: utstring.Trim(utstring.Chains(v.Value, v.Index)),
				Most:   false,
			}

			if strings.HasPrefix(obj.Key, "!") {
				obj.Key = utstring.Sub(obj.Key, 1, 0)
				obj.Most = true

			} else if mst := v.Meta["most"]; mst == "1" {
				obj.Most = true
			}

			if obj.Syntax == "" {
				obj.Syntax = fmt.Sprintf(":%s", obj.Key)
			}

			if strings.HasPrefix(obj.Name, ":") {
				obj.Name = utstring.Sub(obj.Name, 1, 0)
			} else {
				obj.Name = ""
			}

			if strings.HasPrefix(obj.Syntax, ":") {
				obj.Syntax = utstring.Sub(obj.Syntax, 1, 0)
				obj.Syntax = strings.ToLower(obj.Syntax)
				obj.IsReferenced = true
			}

			if strings.HasSuffix(obj.Syntax, "!") {
				obj.Syntax = utstring.Sub(obj.Syntax, 0, -1)
				obj.NoAlias = true
			}

			cur.Column = obj
		}
		views = append(views, cur)
	}
	// #endregion

	// #region[rgba(255, 82, 116, 0.05)] > reading sorts
	type lSort struct {
		OriginKey    string
		Key          string
		Type         string // single, batch
		ReferenceKey string
		Binds        map[string]lColumn
		Allow        []string
		Most         []string
	}

	sorts := []lSort{}

	q, items, erry = u2.ScanTags(q, "sort")
	if erry != nil {
		errx = serror.NewFromSErrc(erry, "while scan sort tags")
		return qo, errx
	}

	for _, v := range items {
		cur := lSort{
			OriginKey: v.Key,
			Most:      []string{"@"},
			Binds:     make(map[string]lColumn),
		}

		if strings.HasPrefix(v.Key, ":") && len(v.Key) > 1 {
			cur.Type = "batch"
			cur.Key = utstring.Sub(v.Key, 1, 0)
			cur.ReferenceKey = utstring.Sub(v.Key, 1, 0)

			if cur.ReferenceKey != "@" {
				if _, ok := alias[cur.ReferenceKey]; !ok {
					errx = serror.Newc(fmt.Sprintf("Cannot find table %s on tag %s:%s", cur.ReferenceKey, v.Tag, v.Key), "while reading_sorts_on_sqlq_select")
					return qo, errx
				}
			}

			if v.Index != "" {
				indexs := utstring.CleanSpit(v.Index, ",")
				for _, v2 := range indexs {
					if strings.Contains(v2, "!") {
						cur.Most = append(cur.Most, v2)
						continue
					}
					cur.Allow = append(cur.Allow, v2)
				}
			}

			for k2, v2 := range v.Meta {
				k3 := k2
				syntx := lColumn{
					Key:    v2,
					Name:   "+",
					Syntax: utstring.Trim(utstring.Chains(v.Value, "__syntax__")),
				}

				if strings.HasPrefix(syntx.Key, "-") {
					syntx.Key = utstring.Sub(syntx.Key, 1, 0)
					syntx.Name = "-"
				}

				if strings.HasPrefix(k3, "!") {
					k3 = utstring.Sub(k3, 1, 0)
					cur.Most = append(cur.Most, fmt.Sprintf("!%s", k3))
				}

				if utarray.CheckAllowedLayer([]string{k3}, [][]string{cur.Most, cur.Allow}) {
					syntx.Key = strings.ToLower(syntx.Key)
					cur.Binds[k3] = syntx
				}
			}

		} else if v.Key != ":" {
			cur.Type = "single"
			cur.Key = v.Key

			if strings.HasPrefix(cur.Key, "!") {
				cur.Key = utstring.Sub(cur.Key, 1, 0)
				cur.Most = []string{"*"}
			}

			if strings.HasPrefix(v.Index, ":") {
				cur.ReferenceKey = utstring.Sub(v.Index, 1, 0)
				obj := lColumn{
					Name:         "+",
					IsReferenced: true,
					Syntax:       utstring.Trim(utstring.Chains(v.Value, "__syntax__")),
				}

				if strings.HasPrefix(cur.ReferenceKey, "-") {
					cur.ReferenceKey = utstring.Sub(cur.ReferenceKey, 1, 0)
					obj.Name = "-"
				}
				obj.Key = strings.ToLower(cur.ReferenceKey)
				cur.Binds[cur.Key] = obj

			} else {
				cur.Binds[cur.Key] = lColumn{
					Syntax: utstring.Trim(utstring.Chains(v.Value, v.Index)),
				}
			}
		}
		cur.ReferenceKey = strings.ToLower(cur.ReferenceKey)
		sorts = append(sorts, cur)
	}
	// #endregion

	// #region[rgba(255, 82, 116, 0.05)] > reading groups
	groups := []lColumn{}

	q, items, erry = u2.ScanTags(q, "group")
	if erry != nil {
		errx = serror.NewFromSErrc(erry, "while scan group tags")
		return qo, errx
	}

	for _, v := range items {
		groups = append(groups, lColumn{
			Key:    v.Key,
			Syntax: utstring.Chains(utstring.Trim(v.Value), v.Index),
		})
	}
	// #endregion

	// === writting parameters ===
	pars := map[string]string{
		"@limit":  "",
		"@offset": "",
		"@orders": "",
	}

	// #region[rgba(246, 255, 74, 0.05)] > processing alias
	type lMColumn struct {
		Column
		TableKey   string
		ColumnKey  string
		ColumnPath string
	}

	type lMTable struct {
		Table
		FullPath   string
		Used       bool
		Name       string
		AllColumns []string
		Columns    map[string]lMColumn
	}

	columns := make(map[string]lMColumn)

	tablex := make(map[string]*lMTable)

	allSortAvailable := []string{}

	for k, v := range alias {
		switch v.Type {
		case "table":
			pars[fmt.Sprintf("#tb:%d", v.Index)] = ""

		case "join":
			pars[fmt.Sprintf("#join:%d", v.Index)] = ""
		}

		var v2 *Table
		var ok bool

		if v2, ok = ox.tables[v.Name]; !ok {
			errx = serror.Newc(fmt.Sprintf("Table %s not exists", v.Name), "while processing_alias_on_sqlq_select")
			return qo, errx
		}

		path := fmt.Sprintf("%s.%s", ox.driver.SQLColumnEscape(v2.Schema), ox.driver.SQLColumnEscape(v2.Table))
		nalias := ox.driver.SQLColumnEscape(k)

		ntable := lMTable{
			Table:      *v2,
			Used:       v.Most,
			Name:       path,
			AllColumns: []string{},
			Columns:    make(map[string]lMColumn),
			FullPath:   fmt.Sprintf("%s AS %s", path, nalias),
		}

		pars[fmt.Sprintf("tb:%s", k)] = ntable.FullPath

		if v.Type == "table" {
			pars[fmt.Sprintf("#tb:%d", v.Index)] = ntable.FullPath
			ntable.Used = true
		}

		colPars[fmt.Sprintf(":@%s", v.Name)] = ntable.FullPath
		colPars[fmt.Sprintf(":~%s", v.Name)] = path

		for k3, v3 := range v2.Columns {
			colKey := fmt.Sprintf("%s.%s", k, k3)

			if v3.Sortable {
				allSortAvailable = append(allSortAvailable, colKey)
			}

			colPath := fmt.Sprintf("(%s)", v3.Name)
			if !v3.IsQuery {
				colPath = fmt.Sprintf("%s.%s", ox.driver.SQLColumnEscape(k), ox.driver.SQLColumnEscape(v3.Name))
			}

			colPars[fmt.Sprintf("::%s", colKey)] = colPath
			if _, ok := colPars[fmt.Sprintf("::%s", k3)]; !ok {
				colPars[fmt.Sprintf("::%s", k3)] = colPath
			}

			colPars[fmt.Sprintf("::%s.%s", k, v3.Alias)] = colPath
			if _, ok := colPars[fmt.Sprintf("::%s", v3.Alias)]; !ok {
				colPars[fmt.Sprintf("::%s", v3.Alias)] = colPath
			}

			if nm := ox.driver.SQLColumnEscape(v3.Name); nm != "" {
				colPars[fmt.Sprintf(":=%s", colKey)] = nm
				if _, ok := colPars[fmt.Sprintf(":=%s", k3)]; !ok {
					colPars[fmt.Sprintf(":=%s", k3)] = nm
				}

				colPars[fmt.Sprintf(":=%s.%s", k, v3.Alias)] = nm
				if _, ok := colPars[fmt.Sprintf(":=%s", v3.Alias)]; !ok {
					colPars[fmt.Sprintf(":=%s", v3.Alias)] = nm
				}
			}

			ncol := lMColumn{
				Column:     *v3,
				TableKey:   k,
				ColumnKey:  k3,
				ColumnPath: colPath,
			}

			columns[colKey] = ncol
			ntable.Columns[k3] = ncol

			if v3.Most || utarray.CheckAllowedLayer([]string{colKey, k3}, [][]string{v.Allow, allow}) {
				ntable.AllColumns = append(ntable.AllColumns, k3)
			}

			if _, ok := columns[k3]; !ok {
				columns[k3] = ncol
			}
		}

		tablex[k] = &ntable
	}
	// #endregion

	// #region[rgba(246, 255, 74, 0.05)] > processing field arguments
	argAllowFields := []string{}
	argFields := make(map[string]lColumn)

	for _, v := range args.Fields {
		if v == "*" {
			argAllowFields = append(argAllowFields, "@")
			continue
		}

		vs := strings.Split(v, ":")
		if len(vs) > 1 {
			argAllowFields = append(argAllowFields, vs[0])
			argFields[vs[0]] = lColumn{
				Key:  vs[0],
				Name: vs[1],
			}
			continue
		}

		argAllowFields = append(argAllowFields, v)
		argFields[v] = lColumn{
			Key: v,
		}
	}
	// #endregion

	// #region[rgba(246, 255, 74, 0.05)] > processing all columns
	type lAColumn struct {
		Relation     int
		Index        string
		Key          []string
		Name         string
		IsReferenced bool
	}

	type lAQuery struct {
		ArgUsed      []string
		Syntax       string
		Alias        string
		ReferenceKey string
	}

	type lAItem struct {
		Index   int
		Columns []lAColumn
		Queries []lAQuery
	}

	allArgColumns := []string{}
	allUsedColumns := []string{}

	virtualColumns := map[string]lColumn{}

	applyItems := make(map[int]lAItem)
	for i := 0; i < 2; i++ {
		for k, v := range views {
			isAll := (v.Type == "field" && v.Field.Name == "@")
			if (i == 0 && isAll) || (i == 1 && !isAll) {
				continue
			}

			pars[fmt.Sprintf("#view:%d", k)] = ""

			ncols := []lAColumn{}
			applies := []lAQuery{}

			if v.Aliases == nil {
				v.Aliases = make(map[string]string)
			}
			if v.Params == nil {
				v.Params = make(map[string][]string)
			}

			switch v.Type {
			case "field":
				iteration := -1
				for cali := range alias {
					iteration++

					aliasName := v.Field.Name

					if aliasName == ":" {
						if iteration > 0 {
							break
						}

					} else if !isAll && cali != aliasName {
						continue
					}

					if isAll {
						aliasName = cali
					}

					ali := alias[aliasName]
					tbx := tablex[aliasName]

					if tbx != nil && ali.Type == "join" {
						jon := joins[ali.Index]
						if jon.Type == "eav" {
							v.Field.Allow = append(v.Field.Allow, jon.Field...)

							prfx := fmt.Sprintf("%s.", aliasName)
							for _, v2 := range argFields {
								if !strings.HasPrefix(v2.Key, prfx) {
									continue
								}

								tbx.Used = true

								coln := utstring.Sub(v2.Key, len(prfx), 0)

								if utarray.CheckAllowedLayer([]string{coln, fmt.Sprintf("%s.%s", aliasName, coln)}, [][]string{v.Field.Allow, ali.Allow}) {
									if _, ok := virtualColumns[v2.Key]; !ok {
										virtualColumns[v2.Key] = lColumn{
											Key:    fmt.Sprintf("%d:%s", k, v2.Name),
											Name:   aliasName,
											Syntax: coln,
										}
									}
								}
							}

							v.Field.Most = []string{"@"}
							v.Field.Allow = []string{"-"}
						}
					}

					if v.Field.IsReferenced {
						for _, v2 := range tbx.AllColumns {
							pattern := []string{
								fmt.Sprintf("%s.%s", aliasName, v2),
								v2,
							}

							ccol := tbx.Columns[v2]
							callows := [][]string{
								v.Field.Most,
								v.Field.Allow,
							}
							if isAll {
								callows = append(callows, [][]string{ali.Allow, allow}...)
							}

							if !ccol.Most && !utarray.CheckAllowedLayer(pattern, callows) {
								continue
							}

							found := false
							for _, v3 := range v.Field.Syntax {
								if found {
									break
								}

								for _, v4 := range pattern {
									if v3.Name == v4 {
										found = true
										break
									}
								}
							}
							if found {
								continue
							}

							v.Field.Syntax[v2] = lColumn{
								Name:         v2,
								Syntax:       "@",
								IsReferenced: true,
							}
						}
					}

					for k2, v2 := range v.Field.Syntax {
						cur := lAColumn{
							Relation:     len(applies),
							Index:        k2,
							Key:          []string{utstring.Chains(v2.Key, v2.Name, k2)},
							Name:         v2.Syntax,
							IsReferenced: v2.IsReferenced,
						}

						columnUsed := []string{}
						allowOriginal := []string{}

						ccolStx, ccolStxOk := lMColumn{}, false
						if v.Field.IsReferenced {
							ccolStx, ccolStxOk = tbx.Columns[v2.Name]

						} else if strings.Contains(v2.Name, ".") {
							ccolStx, ccolStxOk = columns[v2.Name]
						}

						if isAll && !ccolStxOk {
							continue
						}

						if ccolStxOk {
							columnUsed = append(columnUsed, fmt.Sprintf("%s.%s", ccolStx.TableKey, ccolStx.ColumnKey))
							allowOriginal = append(allowOriginal, []string{
								ccolStx.ColumnKey,
								fmt.Sprintf("%s.%s", ccolStx.TableKey, ccolStx.ColumnKey),
							}...)
						}

						ccol := lMColumn{}
						if cur.IsReferenced {
							if v.Field.IsReferenced {
								if cur.Name == "@" {
									cur.Name = fmt.Sprintf("%s.%s", ccolStx.TableKey, ccolStx.ColumnKey)

								} else {
									for k3, v3 := range tbx.Columns {
										if cur.Name == k3 {
											cur.Name = fmt.Sprintf("%s.%s", v3.TableKey, v3.ColumnKey)
											break
										}
									}
								}
							}

							ok := false
							if ccol, ok = columns[cur.Name]; !ok {
								errx = serror.Newc(fmt.Sprintf("Cannot find column %s on tag view:%s", cur.Name, v.OriginKey), "while processing_all_columns_on_sqlq_select")
								return qo, errx
							}

							columnUsed = append(columnUsed, fmt.Sprintf("%s.%s", ccol.TableKey, ccol.ColumnKey))

							nm := utstring.Chains(v2.Key, ccol.Alias, k2)
							cur.Key = []string{
								nm,
								fmt.Sprintf("%s_%s", ccol.TableKey, nm),
								fmt.Sprintf("%s_%s", ccol.TableKey, ccol.ColumnKey),
							}

							for i := 1; i <= 5; i++ {
								cur.Key = append(cur.Key, []string{
									cur.Key[0] + utstring.IntToString(i),
									cur.Key[1] + utstring.IntToString(i),
									cur.Key[2] + utstring.IntToString(i),
								}...)
							}
						}

						pattern := []string{
							cur.Index,
						}

						if v.Field.IsReferenced {
							cpattern := cur.Index
							if !strings.Contains(cur.Index, ".") {
								cpattern = fmt.Sprintf("%s.%s", aliasName, cur.Index)
							}

							if isAll && utarray.IsExist(cpattern, allUsedColumns) {
								continue
							}

							pattern = append(pattern, cpattern)
							pattern = append(pattern, fmt.Sprintf("%s.%s", aliasName, ccol.Alias))
							if cur.Index != cur.Name {
								pattern = append(pattern, []string{
									cur.Name,
									fmt.Sprintf("%s.%s", aliasName, cur.Name),
								}...)
							}
						}
						allUsedColumns = append(allUsedColumns, columnUsed...)

						paramUsed := []string{}
						for _, v := range pattern {
							if !utarray.IsExist(v, allArgColumns) {
								allArgColumns = append(allArgColumns, v)
								paramUsed = append(paramUsed, v)
							}
						}

						if len(paramUsed) <= 0 {
							errx = serror.Newc(fmt.Sprintf("Duplicated key %s on tag view:%s", cur.Index, v.OriginKey), "while processing_all_columns_on_sqlq_select")
							return qo, errx
						}

						v.Params[cur.Index] = paramUsed

						if !v2.NoAlias {
							ncols = append(ncols, cur)
						}

						arg := lColumn{}
						for _, v3 := range paramUsed {
							if a, ok := argFields[v3]; ok {
								arg = a
								break
							}
						}

						qval := lAQuery{
							Alias: arg.Name,
						}

						if cur.IsReferenced {
							qval.ReferenceKey = fmt.Sprintf("%s.%s", ccol.TableKey, ccol.ColumnKey)
						}

						allowPattern := [][]string{}
						mostPattern := [][]string{v.Field.Most}

						if ccolStxOk && ccolStx.Most {
							mostPattern = append(mostPattern, []string{"@", fmt.Sprintf("!%s", paramUsed[0])})
						}

						if cur.IsReferenced {
							aliAllowed := alias[ccol.TableKey].Allow
							if curc, ok := columns[cur.Name]; ok {
								if utarray.CheckAllowedLayer([]string{curc.ColumnKey}, [][]string{aliAllowed}) {
									aliAllowed = append(aliAllowed, cur.Index)
								}
							}
							allowPattern = append(allowPattern, aliAllowed)
						}
						if v.Field.IsReferenced {
							allowPattern = append(allowPattern, ali.Allow)
						} else {
							allowPattern = append(allowPattern, allow)
						}

						if len(allowOriginal) > 0 && utarray.CheckAllowedLayer(paramUsed, [][]string{argAllowFields}) {
							if utarray.CheckAllowedLayer(allowOriginal, allowPattern) {
								mostPattern = append(mostPattern, []string{"@", fmt.Sprintf("!%s", paramUsed[0])})
							}
						}

						allowPattern = append([][]string{argAllowFields}, allowPattern...)
						if utarray.CheckAllowedLayer(paramUsed, append(mostPattern, allowPattern...)) {
							qval.ArgUsed = paramUsed
							qval.Syntax = v2.Syntax

							if cur.IsReferenced {
								qval.Syntax = ccol.ColumnPath

								tablex[ccol.TableKey].Used = true

							} else if ccolStxOk {
								qval.ReferenceKey = fmt.Sprintf("%s.%s", ccolStx.TableKey, ccolStx.ColumnKey)

								tablex[ccolStx.TableKey].Used = true
								qval.Syntax = u2.Binding(qval.Syntax, map[string]string{
									"table":  ox.driver.SQLColumnEscape(ccolStx.TableKey),
									"column": ox.driver.SQLColumnEscape(ccolStx.ColumnKey),
									"syntax": ccolStx.ColumnPath,
									"@":      ccolStx.ColumnPath,
								})
							}

							if v.Field.IsReferenced {
								tbx.Used = true
							}
						}
						applies = append(applies, qval)
					}
				}

			case "column":
				cur := lAColumn{
					Relation:     len(applies),
					Index:        v.Column.Key,
					Key:          []string{v.Column.Key},
					Name:         v.Column.Syntax,
					IsReferenced: v.Column.IsReferenced,
				}

				ccolStx, ccolStxOk := lMColumn{}, false
				if v.Column.Name != "" && strings.Contains(v.Column.Name, ".") {
					ccolStx, ccolStxOk = columns[v.Column.Name]
				}

				if ccolStxOk {
					allUsedColumns = append(allUsedColumns, fmt.Sprintf("%s.%s", ccolStx.TableKey, ccolStx.ColumnKey))
				}

				ccol := lMColumn{}
				if cur.IsReferenced {
					ok := false
					if ccol, ok = columns[cur.Name]; !ok {
						errx = serror.Newc(fmt.Sprintf("Cannot find column %s on tag view:%s", cur.Name, v.OriginKey), "while processing_all_columns_on_sqlq_select")
						return qo, errx
					}

					cur.Key = []string{
						ccol.Alias,
						fmt.Sprintf("%s_%s", ccol.TableKey, ccol.Alias),
					}

					for i := 1; i <= 5; i++ {
						cur.Key = append(cur.Key, []string{
							cur.Key[0] + utstring.IntToString(i),
							cur.Key[1] + utstring.IntToString(i),
						}...)
					}
				}

				if utarray.IsExist(cur.Index, allArgColumns) {
					errx = serror.Newc(fmt.Sprintf("Duplicated key %s on tag view:%s", cur.Index, v.OriginKey), "while processing_all_columns_on_sqlq_select")
					return qo, errx
				}

				allArgColumns = append(allArgColumns, cur.Index)
				v.Params[cur.Index] = []string{cur.Index}

				if !v.Column.NoAlias {
					ncols = append(ncols, cur)
				}

				arg := argFields[cur.Index]
				qval := lAQuery{
					Alias: arg.Name,
				}

				if cur.IsReferenced {
					qval.ReferenceKey = fmt.Sprintf("%s.%s", ccol.TableKey, ccol.ColumnKey)
				}

				if v.Column.Most || utarray.CheckAllowedLayer([]string{cur.Index}, [][]string{argAllowFields, allow}) {
					qval.ArgUsed = []string{cur.Index}
					qval.Syntax = v.Column.Syntax

					if cur.IsReferenced {
						qval.Syntax = ccol.ColumnPath

						tablex[ccol.TableKey].Used = true

					} else if ccolStxOk {
						qval.ReferenceKey = fmt.Sprintf("%s.%s", ccolStx.TableKey, ccolStx.ColumnKey)

						tablex[ccolStx.TableKey].Used = true
						qval.Syntax = u2.Binding(qval.Syntax, map[string]string{
							"table":  ox.driver.SQLColumnEscape(ccolStx.TableKey),
							"column": ox.driver.SQLColumnEscape(ccolStx.ColumnKey),
							"syntax": ccolStx.ColumnPath,
							"@":      ccolStx.ColumnPath,
						})
					}
				}
				applies = append(applies, qval)
			}

			applyItems[k] = lAItem{
				Index:   k,
				Columns: ncols,
				Queries: applies,
			}

			views[k] = v
		}
	}
	// #endregion

	// #region[rgba(74, 143, 255, 0.05)] > applying fields
	allColumnDictionary := make(map[string]lColumn)
	allColumnTables := make(map[string]lColumn)
	allColumns := make(map[string]*lColumn)

	usedVirtualColumns := make(map[string][]string)

	for k, v := range virtualColumns {
		if utarray.IsExist(k, allArgColumns) {
			continue
		}

		kt := v.Key
		ka := ""
		if i := strings.Index(kt, ":"); i > 0 {
			ka = utstring.Sub(kt, i+1, 0)
			kt = utstring.Sub(kt, 0, i)
		}

		k2 := int(utint.StringToInt(kt, -1))
		kn := fmt.Sprintf("%s.%s", v.Name, v.Syntax)
		vn := strings.ReplaceAll(v.Syntax, ".", "_")

		cur := applyItems[k2]

		obj := lAColumn{
			Relation: len(cur.Queries),
			Index:    kn,
			Name:     fmt.Sprintf("@%s", kn),
			Key: []string{
				vn,
				fmt.Sprintf("%s_%s", v.Name, vn),
			},
			IsReferenced: true,
		}

		for i := 1; i <= 5; i++ {
			obj.Key = append(obj.Key, []string{
				obj.Key[0] + utstring.IntToString(i),
				obj.Key[1] + utstring.IntToString(i),
			}...)
		}

		cur.Columns = append(cur.Columns, obj)

		obj2 := lAQuery{
			ArgUsed:      []string{kn},
			Syntax:       fmt.Sprintf("%s.%s", ox.driver.SQLColumnEscape(v.Name), ox.driver.SQLColumnEscape(fmt.Sprintf("_eav_%s", vn))),
			Alias:        ka,
			ReferenceKey: fmt.Sprintf("@%s", kn),
		}
		cur.Queries = append(cur.Queries, obj2)

		if _, ok := usedVirtualColumns[v.Name]; !ok {
			usedVirtualColumns[v.Name] = []string{}
		}
		usedVirtualColumns[v.Name] = append(usedVirtualColumns[v.Name], k)

		applyItems[k2] = cur
	}

	isContainViews := false
	for k, view := range views {
		v := applyItems[k]

		for _, v2 := range v.Columns {
			k2 := ""
			for _, v3 := range v2.Key {
				if utstring.Trim(v3) == "" {
					continue
				}

				if _, ok := allColumns[v3]; !ok {
					k2 = v3
					break
				}
			}

			k3 := k2
			apd := v.Queries[v2.Relation]
			if apd.Syntax != "" {
				if apd.Alias == "@" && v2.IsReferenced {
					ccol := columns[v2.Name]
					apd.Alias = ccol.Name
					k3 = apd.Alias
				}
			}

			if k3 == "" {
				errx = serror.Newc(fmt.Sprintf("Duplicated column name %s on tag view:%s", v2.Index, view.OriginKey), "while applying_fields_on_sqlq_select")
				return qo, errx
			}

			allColumns[k2] = &lColumn{
				Key:          utstring.IntToString(v.Index),
				Name:         v2.Index,
				IsReferenced: v2.IsReferenced,
				Syntax:       k3,
			}

			view.Aliases[v2.Index] = k2

			if apd.Syntax != "" {
				aliasUsed := utstring.Chains(apd.Alias, k2)
				view.Aliases[v2.Index] = aliasUsed

				for _, v3 := range apd.ArgUsed {
					allColumnDictionary[v3] = lColumn{
						Key:     apd.Syntax,
						Name:    aliasUsed,
						Syntax:  apd.ReferenceKey,
						NoAlias: false,
					}
				}

				if apd.ReferenceKey != "" {
					allColumnTables[apd.ReferenceKey] = lColumn{
						Key:    k2,
						Name:   aliasUsed,
						Syntax: apd.ReferenceKey,
					}
				}

				apd.Syntax = fmt.Sprintf("%s AS %s", apd.Syntax, ox.driver.SQLColumnEscape(aliasUsed))
				v.Queries[v2.Relation] = apd
			}
		}

		qapplies := []string{}
		for _, v2 := range v.Queries {
			if v2.Syntax != "" {
				if v2.ReferenceKey == "" {
					binders := u2.ScanPrefix(v2.Syntax, []string{":~", ":@", "::", ":="})
					for _, v3 := range binders[":~"] {
						tbx, ok := tablex[strings.ToLower(v3)]
						if !ok {
							errx = serror.Newc(fmt.Sprintf("Cannot find table %s on tag view:%s", v3, view.OriginKey), "while applying_fields_on_sqlq_select")
							return qo, errx
						}

						v2.Syntax = u2.Binding(v2.Syntax, map[string]string{
							fmt.Sprintf(":~%s", v3): tbx.Name,
						})

						tbx.Used = true
					}

					for _, v3 := range binders[":@"] {
						tbx, ok := tablex[strings.ToLower(v3)]
						if !ok {
							errx = serror.Newc(fmt.Sprintf("Cannot find table %s on tag view:%s", v3, view.OriginKey), "while applying_fields_on_sqlq_select")
							return qo, errx
						}

						v2.Syntax = u2.Binding(v2.Syntax, map[string]string{
							fmt.Sprintf(":~%s", v3): tbx.FullPath,
						})

						tbx.Used = true
					}

					for _, v3 := range binders["::"] {
						ccol, ok := columns[strings.ToLower(v3)]
						if !ok {
							errx = serror.Newc(fmt.Sprintf("Cannot find column %s on tag view:%s", v3, view.OriginKey), "while applying_fields_on_sqlq_select")
							return qo, errx
						}

						v2.Syntax = u2.Binding(v2.Syntax, map[string]string{
							fmt.Sprintf("::%s", v3): ccol.ColumnPath,
						})

						tablex[ccol.TableKey].Used = true
					}

					for _, v3 := range binders[":="] {
						ccol, ok := columns[strings.ToLower(v3)]
						if !ok {
							errx = serror.Newc(fmt.Sprintf("Cannot find column %s on tag view:%s", v3, view.OriginKey), "while applying_fields_on_sqlq_select")
							return qo, errx
						}

						v2.Syntax = u2.Binding(v2.Syntax, map[string]string{
							fmt.Sprintf(":=%s", v3): ox.driver.SQLColumnEscape(ccol.Name),
						})

						tablex[ccol.TableKey].Used = true
					}
				}

				for _, v3 := range v2.ArgUsed {
					if _, ok := allColumnDictionary[v3]; !ok {
						allColumnDictionary[v3] = lColumn{
							Key:     v2.Syntax,
							Name:    "",
							Syntax:  v2.ReferenceKey,
							NoAlias: true,
						}
					}
				}

				qapplies = append(qapplies, v2.Syntax)
			}
		}

		if len(qapplies) > 0 {
			k2 := fmt.Sprintf("#view:%d", v.Index)
			if isContainViews {
				pars[k2] = ", "
			}
			pars[k2] += strings.Join(qapplies, ", ")

			isContainViews = true
		}

		views[v.Index] = view
	}
	// #endregion

	// #region[rgba(246, 255, 74, 0.05)] > processing condition arguments
	argConds := make(map[string]lCondition)

	for k, v := range args.Conditions {
		ks := utstring.CleanSpit(k, ":")
		if len(ks) > 1 {
			argConds[ks[0]] = lCondition{
				Name:   ks[0],
				Syntax: ks[1],
				Value:  v,
			}
			continue
		}

		argConds[k] = lCondition{
			Name:   utstring.Trim(k),
			Syntax: "",
			Value:  v,
		}
	}
	// #endregion

	// #region[rgba(74, 143, 255, 0.05)] > applying conditions
	for k, v := range conds {
		k2 := fmt.Sprintf("#cond:%d", k)
		pars[k2] = ""

		if colc, ok := argConds[v.Key]; ok {
			for _, v := range v.Depends {
				tablex[v].Used = true
			}

			if v.IsReferenced {
				col, colOk := columns[strings.ToLower(v.Name)]
				if !colOk {
					errx = serror.Newf("Cannot find object %s on tag cond:%s", v.Name, v.Key)
					return qo, errx
				}

				tablex[col.TableKey].Used = true

				if col.Condition == nil {
					errx = serror.Newf("Column %s has no condition available, on tag cond:%s", colc.Key, v.Key)
					return qo, errx
				}

				var (
					opr     = utstring.Chains(colc.Syntax, string(col.Condition.Default))
					allows  = [][]Operator{v.Allow, col.Condition.AllowOperator}
					allowed = false
				)
				for _, v2 := range allows {
					if OperatorExists(ToOperator(opr), v2) {
						allowed = true
						break
					}
				}

				if !allowed {
					errx = serror.Newf("Operator %s is not allowed on %s", opr, colc.Key)
					return qo, errx
				}

				condObj, ok := ox.driver.ToSQLConditionObject(
					QColumn{col.TableKey, col.Name},
					ToOperator(opr),
					colc.Value,
				)
				if !ok {
					errx = serror.Newf("Cannot resolve condition on tag cond:%s", v.Key)
					return qo, errx
				}

				condOpts := map[string]string{
					"name":     condObj.Condition1,
					"operator": condObj.Operator,
					"value":    condObj.Condition2,
					"syntax":   condObj.Syntax(),
				}

				if ok {
					v.Syntax = u2.Binding(v.Syntax, colPars)
					pars[k2] = u2.Binding(v.Syntax, condOpts)
					continue
				}

				continue
			}

			// not referenced
			{
				var (
					opr     = utstring.Chains(colc.Syntax, "=")
					allows  = [][]Operator{v.Allow}
					allowed = false
				)
				for _, v2 := range allows {
					if OperatorExists(ToOperator(opr), v2) {
						allowed = true
						break
					}
				}

				if !allowed {
					errx = serror.Newf("Operator %s is not allowed on %s", opr, colc.Key)
					return qo, errx
				}

				condOpts := map[string]string{
					"operator": string(ToOperator(opr)),
				}

				colVal, ok := ox.driver.ToSQLValueQuery(colc.Value)
				if !ok {
					errx = serror.Newf("Failed to parser sql value query on %s", colc.Key)
					return qo, errx
				}

				condOpts["value"] = colVal

				v.Syntax = u2.Binding(v.Syntax, colPars)
				pars[k2] = u2.Binding(v.Syntax, condOpts)
			}

			continue
		}

		if v.Most {
			errx = serror.Newf("Condition %s must filled", v.Key)
			return qo, errx
		}

		pars[k2] = v.DefaultSyntax
	}
	// #endregion

	// #region[rgba(246, 255, 74, 0.05)] > processing all sorting
	type lASort struct {
		Value        lColumn
		ReferenceKey string
		Syntax       string
	}

	allSortDictionary := make(map[string]lASort)
	allSortOrder := make(map[int][]string)
	allSortColumnUsed := []string{}
	allSortUsed := []string{}
	allSortIndex := -1
	isAllSort := false

	for k, v := range sorts {
		pars[fmt.Sprintf("#sort:%d", k)] = ""

		if _, ok := allSortOrder[k]; !ok {
			allSortOrder[k] = []string{}
		}

		switch v.Type {
		case "batch":
			if v.ReferenceKey == "@" {
				allSortIndex = k
				isAllSort = true
				continue
			}

			tbx := tablex[v.ReferenceKey]

			for k2, v2 := range tbx.Columns {
				ckey := fmt.Sprintf("%s.%s", v.ReferenceKey, v2.Alias)
				pattern := []string{
					ckey,
					v2.Alias,
				}

				allSortColumnUsed = append(allSortColumnUsed, ckey)

				argPattern := []string{
					pattern[0],
					pattern[1],
				}

				defSort := "+"
				bindSyntax := "__syntax__"
				if bind, ok := v.Binds[k2]; ok {
					if utarray.IsExist(bind.Key, allSortUsed) {
						errx = serror.Newc(fmt.Sprintf("Duplicated key %s on tag sort:%s", bind.Key, v.OriginKey), "while processing_all_sorting_on_sqlq_select")
						return qo, errx
					}

					argPattern = []string{bind.Key}
					bindSyntax = bind.Syntax
					defSort = bind.Name
				}

				paramUsed := []string{}
				for _, v3 := range argPattern {
					if !utarray.IsExist(v3, allSortUsed) {
						allSortUsed = append(allSortUsed, v3)
						paramUsed = append(paramUsed, v3)
					}
				}

				if len(paramUsed) <= 0 {
					errx = serror.Newc(fmt.Sprintf("Duplicated key %s on tag sort:%s", ckey, v.OriginKey), "while processing_all_sorting_on_sqlq_select")
					return qo, errx
				}

				if utarray.CheckAllowedLayer(pattern, [][]string{v.Most, v.Allow, allSortAvailable}) {
					isMost := utarray.CheckAllowedLayer(pattern, [][]string{v.Most, []string{"-"}})

					item := lASort{}
					if col, ok := allColumnTables[ckey]; ok {
						if !col.NoAlias {
							item = lASort{
								Value: lColumn{
									Key:  ox.driver.SQLColumnEscape(col.Name),
									Name: defSort,
									Most: isMost,
								},
								Syntax:       bindSyntax,
								ReferenceKey: ckey,
							}
						}
					}

					if item == (lASort{}) {
						if col, ok := columns[ckey]; ok {
							item = lASort{
								Value: lColumn{
									Key:  col.ColumnPath,
									Name: defSort,
									Most: isMost,
								},
								Syntax:       bindSyntax,
								ReferenceKey: fmt.Sprintf("%s.%s", col.TableKey, col.ColumnKey),
							}
						}
					}

					if item == (lASort{}) {
						errx = serror.Newc(fmt.Sprintf("Unknown column %s on tag sort:%s", ckey, v.OriginKey), "while processing_all_sorting_on_sqlq_select")
						return qo, errx
					}

					allSortOrder[k] = append(allSortOrder[k], paramUsed[0])
					for _, v3 := range paramUsed {
						allSortDictionary[v3] = item
					}
				}
			}

		case "single":
			for k2, v2 := range v.Binds {
				if utarray.IsExist(k2, allSortUsed) {
					errx = serror.Newc(fmt.Sprintf("Duplicated key %s on tag sort:%s", k2, v.OriginKey), "while processing_all_sorting_on_sqlq_select")
					return qo, errx
				}

				if v2.IsReferenced {
					if col, ok := allColumnDictionary[v2.Key]; ok {
						if col.Syntax != "" {
							allSortColumnUsed = append(allSortColumnUsed, fmt.Sprintf("%s#%s", col.Syntax, k2))
						}

					} else if col, ok := columns[v2.Key]; ok {
						allSortColumnUsed = append(allSortColumnUsed, fmt.Sprintf("%s.%s#%s", col.TableKey, col.ColumnKey, k2))
					}
				}

				if utarray.CheckAllowedLayer([]string{k2}, [][]string{v.Most}) {
					isMost := utarray.CheckAllowedLayer([]string{k2}, [][]string{v.Most, []string{"-"}})

					allSortOrder[k] = append(allSortOrder[k], k2)
					if v2.IsReferenced {
						if col, ok := allColumnDictionary[v2.Key]; ok {
							if !col.NoAlias {
								allSortDictionary[k2] = lASort{
									Value: lColumn{
										Key:  ox.driver.SQLColumnEscape(col.Name),
										Name: utstring.Chains(v2.Name, "+"),
										Most: isMost,
									},
									Syntax:       v2.Syntax,
									ReferenceKey: col.Syntax,
								}
								continue
							}
						}

						if col, ok := columns[v2.Key]; ok {
							allSortDictionary[k2] = lASort{
								Value: lColumn{
									Key:  col.ColumnPath,
									Name: utstring.Chains(v2.Name, "+"),
									Most: isMost,
								},
								Syntax:       v2.Syntax,
								ReferenceKey: fmt.Sprintf("%s.%s", col.TableKey, col.ColumnKey),
							}
							continue
						}

						errx = serror.Newc(fmt.Sprintf("Unknown column %s on tag sort:%s", k2, v.OriginKey), "while processing_all_sorting_on_sqlq_select")
						return qo, errx
					}

					allSortDictionary[k2] = lASort{
						Value: lColumn{
							Syntax: v2.Syntax,
							Most:   isMost,
						},
						Syntax: "__syntax__",
					}
				}
			}
		}
	}

	if isAllSort {
		for _, v := range allSortAvailable {
			if !utarray.IsExist(v, allSortColumnUsed) {
				col := columns[v]
				ckey := fmt.Sprintf("%s.%s", col.TableKey, col.ColumnKey)
				cali := fmt.Sprintf("%s.%s", col.TableKey, col.Alias)

				pattern := []string{
					col.Alias,
					cali,
				}

				item := lASort{
					Value: lColumn{
						Key:  col.ColumnPath,
						Name: "+",
					},
					Syntax:       "__syntax__",
					ReferenceKey: ckey,
				}

				isFirst := true
				for _, v2 := range pattern {
					if utarray.IsExist(v2, allSortUsed) {
						continue
					}

					if isFirst {
						allSortOrder[allSortIndex] = append(allSortOrder[allSortIndex], v2)
					}
					isFirst = false

					allSortUsed = append(allSortUsed, v2)
					allSortDictionary[v2] = item
				}
			}
		}
	}
	// #endregion

	// #region[rgba(74, 143, 255, 0.05)] > applying sortings
	allSortArgUsed := []string{}
	allSortSyntax := []string{}

	for _, v := range args.Sorting {
		v2 := v
		mode := "ASC"

		if strings.HasPrefix(v, "-") {
			v2 = utstring.Sub(v2, 1, 0)
			mode = "DESC"
		}

		if srt, ok := allSortDictionary[v2]; ok {
			allSortArgUsed = append(allSortArgUsed, v2)

			if srt.ReferenceKey != "" {
				ccol := columns[srt.ReferenceKey]
				tablex[ccol.TableKey].Used = true
			}

			sbind := map[string]string{
				"name":   srt.Value.Syntax,
				"mode":   mode,
				"syntax": fmt.Sprintf("%s %s", srt.Value.Syntax, mode),
			}
			if srt.Value.Key != "" {
				sbind["name"] = srt.Value.Key
				sbind["syntax"] = fmt.Sprintf("%s %s", srt.Value.Key, mode)
			}

			syntx := u2.Binding(srt.Syntax, sbind)
			syntx = u2.Binding(syntx, colPars)
			allSortSyntax = append(allSortSyntax, syntx)
		}
	}

	if len(allSortArgUsed) <= 0 {
		for k := range sorts {
			for _, v := range allSortOrder[k] {
				v2 := allSortDictionary[v]
				if v2.Value.Most {
					if v2.ReferenceKey != "" {
						ccol := columns[v2.ReferenceKey]
						tablex[ccol.TableKey].Used = true
					}

					mode := ""
					if v2.Value.Key != "" {
						mode = "ASC"
						if v2.Value.Name == "-" {
							mode = "DESC"
						}
					}

					syntx := u2.Binding(v2.Value.Syntax, map[string]string{
						"name":   v2.Value.Key,
						"mode":   mode,
						"syntax": fmt.Sprintf("%s %s", v2.Value.Key, mode),
					})
					syntx = u2.Binding(syntx, colPars)
					allSortSyntax = append(allSortSyntax, syntx)
				}
			}
		}
	}

	if len(allSortSyntax) > 0 {
		pars["#sort:0"] = fmt.Sprintf("ORDER BY %s", strings.Join(allSortSyntax, ", "))
	}
	// #endregion

	// #region[rgba(246, 255, 74, 0.05)] > processing all join
	type lAJoin struct {
		Table  string
		Syntax string
	}

	applyJoins := make(map[int]lAJoin)
	for k, v := range joins {
		obj := lAJoin{
			Syntax: v.Syntax,
		}

		tbx := tablex[v.Key]
		obj.Table = tbx.FullPath

		privColPars := make(map[string]string)
		for k2, v2 := range tbx.Columns {
			privColPars[fmt.Sprintf("::%s", k2)] = v2.ColumnPath
			privColPars[fmt.Sprintf("::%s", v2.Alias)] = v2.ColumnPath
			privColPars[fmt.Sprintf("::@.%s", k2)] = v2.ColumnPath
			privColPars[fmt.Sprintf("::@.%s", v2.Alias)] = v2.ColumnPath

			if nm := ox.driver.SQLColumnEscape(v2.Name); nm != "" {
				privColPars[fmt.Sprintf(":=%s", k2)] = nm
				privColPars[fmt.Sprintf(":=%s", v2.Alias)] = nm
				privColPars[fmt.Sprintf(":=@.%s", k2)] = nm
				privColPars[fmt.Sprintf(":=@.%s", v2.Alias)] = nm
			}
		}

		prfx := u2.ScanPrefix(fmt.Sprintf("%s\n%s", v.Condition, v.Syntax), []string{"::", ":=", ":#"})
		for _, v2 := range prfx["::"] {
			v3 := strings.Split(v2, ".")

			if len(v3) >= 2 && v3[0] != "@" {
				if ccol, ok := columns[strings.ToLower(v2)]; ok {
					if tbx.Used {
						tablex[ccol.TableKey].Used = true
					}
					continue
				}

				errx = serror.Newc(fmt.Sprintf("Cannot find column %s on tag join:%s", v2, v.Name), "while processing_all_join_on_sqlq_select")
				return qo, errx
			}
		}

		for _, v2 := range prfx[":="] {
			v3 := strings.Split(v2, ".")

			if len(v3) >= 2 && v3[0] != "@" {
				if ccol, ok := columns[strings.ToLower(v2)]; ok {
					if tbx.Used {
						tablex[ccol.TableKey].Used = true
					}
					continue
				}

				errx = serror.Newc(fmt.Sprintf("Cannot find column %s on tag join:%s", v2, v.Name), "while processing_all_join_on_sqlq_select")
				return qo, errx
			}
		}

		for _, v2 := range prfx[":#"] {
			v2sp := utstring.CleanSpit(v2, ".")
			if len(v2sp) <= 1 {
				continue
			}

			switch v2sp[0] {
			case "conds":
				v3 := "NULL"
				if v4, ok := args.Conditions[strings.Join(v2sp[1:], ".")]; ok {
					if v5, ok := ox.driver.ToSQLValueQuery(v4); ok {
						v3 = v5
					}
				}
				privColPars[fmt.Sprintf(":#%s", v2)] = v3
			}
		}

		cond := v.Condition
		cond = u2.Binding(cond, privColPars)
		cond = u2.Binding(cond, colPars)

		obj.Syntax = u2.Binding(obj.Syntax, privColPars)
		obj.Syntax = u2.Binding(obj.Syntax, colPars)

		privPars := make(map[string]string)

		switch v.Type {
		case "eav":
			eavRoles := make(map[EAVRole]string)
			for _, v2 := range tbx.Columns {
				switch v2.EAVRole {
				case EAVRoleKey, EAVRoleGroup, EAVRoleValue:
					eavRoles[v2.EAVRole] = v2.ColumnPath
				}
			}

			if _, ok := eavRoles[EAVRoleKey]; !ok {
				errx = serror.Newc(fmt.Sprintf("No EAV role key on table %s on tag join:%s", v.Key, v.Key), "while processing_all_join_on_sqlq_select")
				return qo, errx
			}

			if _, ok := eavRoles[EAVRoleValue]; !ok {
				errx = serror.Newc(fmt.Sprintf("No EAV role value on table %s on tag join:%s", v.Key, v.Key), "while processing_all_join_on_sqlq_select")
				return qo, errx
			}

			colValNm := eavRoles[EAVRoleValue]
			if cst, ok := v.Options["vcast"]; ok {
				if len(cst) > 0 {
					switch string(cst[0]) {
					case "$":
						colValNm = fmt.Sprintf(utstring.Sub(cst, 1, 0), colValNm)

					case "!":
						colValNm = fmt.Sprintf("%s%s", colValNm, utstring.Sub(cst, 1, 0))

					default:
						colValNm = fmt.Sprintf("%s::%s", colValNm, cst)
					}
				}
			}

			tempx := "MAX(CASE WHEN %s = __value__ __group__ THEN %s ELSE NULL END) AS __alias__"
			tempx = fmt.Sprintf(tempx, eavRoles[EAVRoleKey], colValNm)
			colGrpNm := ""

			if grp, ok := eavRoles[EAVRoleGroup]; ok {
				colGrpNm = grp
			}

			if vcs, ok := usedVirtualColumns[v.Key]; ok {
				var (
					vccols      = []string{}
					condKeyOnly = []string{}
					condCode    = []string{}
				)

				for _, v2 := range vcs {
					cvc := virtualColumns[v2]
					cvcNm := strings.ReplaceAll(cvc.Syntax, ".", "_")

					colNm, colGrp := cvc.Syntax, ""
					if _, ok := eavRoles[EAVRoleGroup]; ok {
						if ei := strings.Index(colNm, "."); ei > 0 {
							colGrp = utstring.Sub(colNm, 0, ei)
							colNm = utstring.Sub(colNm, ei+1, 0)
						}
					}

					colNmVal, _ := ox.driver.ToSQLValueQuery(colNm)
					privJonPars := map[string]string{
						"value": colNmVal,
						"group": "",
						"alias": ox.driver.SQLColumnEscape(fmt.Sprintf("_eav_%s", cvcNm)),
					}
					if colGrp != "" {
						colGrpVal, _ := ox.driver.ToSQLValueQuery(colGrp)
						privJonPars["group"] = fmt.Sprintf(" AND %s = %s", colGrpNm, colGrpVal)
						condCode = append(condCode, fmt.Sprintf("%s.%s", colGrp, colNm))

					} else {
						condKeyOnly = append(condKeyOnly, colNm)
					}

					vccols = append(vccols, u2.Binding(tempx, privJonPars))
				}

				for _, v2 := range v.Field {
					if ccol, ok := tbx.Columns[v2]; ok {
						vccols = append(vccols, ccol.ColumnPath)
					}
				}

				if len(condCode) > 0 || len(condKeyOnly) > 0 {
					var tmpStx []string

					if len(condCode) > 0 {
						stxCode, ok := ox.driver.ToSQLConditionQuery(
							QRaw(fmt.Sprintf("CONCAT_WS('.', %s, %s)", eavRoles[EAVRoleGroup], eavRoles[EAVRoleKey])),
							OperatorIn,
							condCode,
						)
						if !ok {
							errx = serror.Newc(fmt.Sprintf("Failed to create condition code query on tag join:%s", v.Key), "@")
							return "", errx
						}
						tmpStx = append(tmpStx, stxCode)
					}

					if len(condKeyOnly) > 0 {
						stxKeyOnly, ok := ox.driver.ToSQLConditionQuery(QRaw(eavRoles[EAVRoleKey]), OperatorIn, condKeyOnly)
						if !ok {
							errx = serror.Newc(fmt.Sprintf("Failed to create condition keyOnly query on tag join:%s", v.Key), "@")
							return "", errx
						}
						tmpStx = append(tmpStx, stxKeyOnly)
					}

					cond = fmt.Sprintf("%s AND (%s)", cond, strings.Join(tmpStx, " OR "))
				}

				grpVal := ""
				if len(v.Group) > 0 {
					grps := []string{}
					for _, v2 := range v.Group {
						if ccol, ok := tbx.Columns[v2]; ok {
							grps = append(grps, ccol.ColumnPath)
							continue
						}

						errx = serror.Newc(fmt.Sprintf("Cannot find column %s on tag join:%s", v2, v.Key), "while processing_all_join_on_sqlq_select")
						return "", errx
					}

					grpVal = fmt.Sprintf("GROUP BY %s", strings.Join(grps, ", "))
				}

				privPars["alias"] = ox.driver.SQLColumnEscape(v.Key)
				privPars["syntax"] = u2.Binding(`
					SELECT
						__fields__
					FROM __table__
					WHERE
						__conditions__
					__group__
				`, map[string]string{
					"fields":     strings.Join(vccols, ", "),
					"table":      obj.Table,
					"conditions": cond,
					"group":      grpVal,
				})

				obj.Syntax = utstring.Chains(obj.Syntax, "JOIN LATERAL (__syntax__) AS __alias__ ON TRUE")
			}

		default:
			privPars["alias"] = ox.Driver().SQLColumnEscape(v.Key)
			privPars["name"] = tbx.Name
			privPars["table"] = obj.Table
			privPars["conditions"] = cond
			privPars["syntax"] = fmt.Sprintf("%s ON %s", privPars["table"], privPars["conditions"])

			obj.Syntax = utstring.Chains(obj.Syntax, "JOIN __syntax__")
		}

		applyJoins[k] = lAJoin{
			Table:  v.Key,
			Syntax: u2.Binding(obj.Syntax, privPars),
		}
	}
	// #endregion

	// #region[rgba(74, 143, 255, 0.05)] > applying joins
	for k, v := range applyJoins {
		tbx := tablex[v.Table]

		if tbx.Used {
			pars[fmt.Sprintf("#join:%d", k)] = v.Syntax
		}
	}
	// #endregion

	// #region[rgba(74, 143, 255, 0.05)] > applying groups
	allGroupSyntax := []string{}

	for k, v := range groups {
		k2 := fmt.Sprintf("#group:%d", k)
		pars[k2] = ""

		if col, ok := allColumnDictionary[v.Key]; ok {
			syntx := utstring.Chains(v.Syntax, col.Key)
			syntx = u2.Binding(syntx, map[string]string{
				"@": col.Key,
			})
			syntx = u2.Binding(syntx, colPars)

			allGroupSyntax = append(allGroupSyntax, syntx)
		}
	}

	if len(allGroupSyntax) > 0 {
		pars["#group:0"] = fmt.Sprintf("GROUP BY %s", strings.Join(allGroupSyntax, ", "))
	}
	// #endregion

	// #region[rgba(74, 143, 255, 0.05)] > applying limit
	if args.Limit >= 0 {
		if args.Limit == 0 {
			args.Limit = 100
		}

		pars["@limit"] = fmt.Sprintf("LIMIT %d", args.Limit)
	}
	// #endregion

	// #region[rgba(74, 143, 255, 0.05)] > applying offset
	if args.Offset > 0 {
		pars["@offset"] = fmt.Sprintf("OFFSET %d", args.Offset)
	}
	// #endregion

	q = u2.Binding(q, colPars)
	return u2.Binding(q, pars), nil
}
