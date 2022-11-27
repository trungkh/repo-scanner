package model

import (
	"repo-scanner/internal/utils/serror"
	"repo-scanner/internal/utils/sqlq"
)

type (
	Metas map[string]interface{}

	WriteItem struct {
		Key   int
		ID    int64
		RawID interface{}
		Error serror.SError
		Metas Metas
	}

	WriteResult struct {
		Items []WriteItem
	}

	HeaderDetailTable struct {
		Header      sqlq.TableData
		Detail      sqlq.EAVData
		DeleteMetas map[string]interface{}
	}
)

func (ox WriteResult) IsSuccess() bool {
	for _, v := range ox.Items {
		if v.Error != nil {
			return false
		}
	}
	return true
}

func (ox WriteResult) AffectedCount() int64 {
	result := int64(0)
	for _, v := range ox.Items {
		if v.Error == nil {
			result++
		}
	}
	return result
}
