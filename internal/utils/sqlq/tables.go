package sqlq

import (
	"repo-scanner/internal/utils/serror"
)

type Tables map[string]*Table

type NewTableOption struct {
	Schema       string
	Table        string
	IsEAV        bool
	ConditionMap OperatorsMap
}

func (ox *Tables) Add(key string, dat Table) {
	if *ox == nil {
		*ox = make(Tables)
	}

	(*ox)[key] = &dat
}

// TODO: still draft
// func (ox *Tables) LoadFromDB(db *sqlx.DB) (errx serror.SError) {
// 	for _, v := range *ox {
// 		errx = v.LoadFromDB(db)
// 		if errx != nil {
// 			return errx
// 		}
// 	}

// 	return nil
// }

func (ox *Tables) AddFromStruct(key string, opt NewTableOption, obj interface{}) (errx serror.SError) {
	if *ox == nil {
		*ox = make(Tables)
	}

	out := Table{
		Schema:       opt.Schema,
		Table:        opt.Table,
		IsEAV:        opt.IsEAV,
		ConditionMap: opt.ConditionMap,
	}
	errx = out.LoadFromStruct(obj)
	if errx == nil {
		ox.Add(key, out)
	}

	return errx
}
