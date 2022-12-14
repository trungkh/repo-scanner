// Code generated by mockery v2.14.0. DO NOT EDIT.

package mocks

import (
	model "repo-scanner/internal/model"

	mock "github.com/stretchr/testify/mock"

	serror "repo-scanner/internal/utils/serror"

	types "github.com/jmoiron/sqlx/types"
)

// IScanningRepository is an autogenerated mock type for the IScanningRepository type
type IScanningRepository struct {
	mock.Mock
}

// AddNewScanning provides a mock function with given fields: _a0, _a1
func (_m *IScanningRepository) AddNewScanning(_a0 *model.Trx, _a1 int64) (model.ScanningResponse, serror.SError) {
	ret := _m.Called(_a0, _a1)

	var r0 model.ScanningResponse
	if rf, ok := ret.Get(0).(func(*model.Trx, int64) model.ScanningResponse); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Get(0).(model.ScanningResponse)
	}

	var r1 serror.SError
	if rf, ok := ret.Get(1).(func(*model.Trx, int64) serror.SError); ok {
		r1 = rf(_a0, _a1)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(serror.SError)
		}
	}

	return r0, r1
}

// EditScanningStatusById provides a mock function with given fields: _a0, _a1, _a2, _a3
func (_m *IScanningRepository) EditScanningStatusById(_a0 *model.Trx, _a1 int64, _a2 string, _a3 types.JSONText) (model.ScanningResponse, serror.SError) {
	ret := _m.Called(_a0, _a1, _a2, _a3)

	var r0 model.ScanningResponse
	if rf, ok := ret.Get(0).(func(*model.Trx, int64, string, types.JSONText) model.ScanningResponse); ok {
		r0 = rf(_a0, _a1, _a2, _a3)
	} else {
		r0 = ret.Get(0).(model.ScanningResponse)
	}

	var r1 serror.SError
	if rf, ok := ret.Get(1).(func(*model.Trx, int64, string, types.JSONText) serror.SError); ok {
		r1 = rf(_a0, _a1, _a2, _a3)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(serror.SError)
		}
	}

	return r0, r1
}

// GetScanningList provides a mock function with given fields: _a0
func (_m *IScanningRepository) GetScanningList(_a0 model.ScanningListRequest) ([]model.ScanningListResponse, serror.SError) {
	ret := _m.Called(_a0)

	var r0 []model.ScanningListResponse
	if rf, ok := ret.Get(0).(func(model.ScanningListRequest) []model.ScanningListResponse); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]model.ScanningListResponse)
		}
	}

	var r1 serror.SError
	if rf, ok := ret.Get(1).(func(model.ScanningListRequest) serror.SError); ok {
		r1 = rf(_a0)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(serror.SError)
		}
	}

	return r0, r1
}

type mockConstructorTestingTNewIScanningRepository interface {
	mock.TestingT
	Cleanup(func())
}

// NewIScanningRepository creates a new instance of IScanningRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewIScanningRepository(t mockConstructorTestingTNewIScanningRepository) *IScanningRepository {
	mock := &IScanningRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
