// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/influxdata/influxdb-client-go/v2/api (interfaces: QueryAPI)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	api "github.com/influxdata/influxdb-client-go/v2/api"
	domain "github.com/influxdata/influxdb-client-go/v2/domain"
)

// MockQueryAPI is a mock of QueryAPI interface.
type MockQueryAPI struct {
	ctrl     *gomock.Controller
	recorder *MockQueryAPIMockRecorder
}

// MockQueryAPIMockRecorder is the mock recorder for MockQueryAPI.
type MockQueryAPIMockRecorder struct {
	mock *MockQueryAPI
}

// NewMockQueryAPI creates a new mock instance.
func NewMockQueryAPI(ctrl *gomock.Controller) *MockQueryAPI {
	mock := &MockQueryAPI{ctrl: ctrl}
	mock.recorder = &MockQueryAPIMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockQueryAPI) EXPECT() *MockQueryAPIMockRecorder {
	return m.recorder
}

// Query mocks base method.
func (m *MockQueryAPI) Query(arg0 context.Context, arg1 string) (*api.QueryTableResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Query", arg0, arg1)
	ret0, _ := ret[0].(*api.QueryTableResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Query indicates an expected call of Query.
func (mr *MockQueryAPIMockRecorder) Query(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Query", reflect.TypeOf((*MockQueryAPI)(nil).Query), arg0, arg1)
}

// QueryRaw mocks base method.
func (m *MockQueryAPI) QueryRaw(arg0 context.Context, arg1 string, arg2 *domain.Dialect) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QueryRaw", arg0, arg1, arg2)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// QueryRaw indicates an expected call of QueryRaw.
func (mr *MockQueryAPIMockRecorder) QueryRaw(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryRaw", reflect.TypeOf((*MockQueryAPI)(nil).QueryRaw), arg0, arg1, arg2)
}

// QueryRawWithParams mocks base method.
func (m *MockQueryAPI) QueryRawWithParams(arg0 context.Context, arg1 string, arg2 *domain.Dialect, arg3 interface{}) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QueryRawWithParams", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// QueryRawWithParams indicates an expected call of QueryRawWithParams.
func (mr *MockQueryAPIMockRecorder) QueryRawWithParams(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryRawWithParams", reflect.TypeOf((*MockQueryAPI)(nil).QueryRawWithParams), arg0, arg1, arg2, arg3)
}

// QueryWithParams mocks base method.
func (m *MockQueryAPI) QueryWithParams(arg0 context.Context, arg1 string, arg2 interface{}) (*api.QueryTableResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QueryWithParams", arg0, arg1, arg2)
	ret0, _ := ret[0].(*api.QueryTableResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// QueryWithParams indicates an expected call of QueryWithParams.
func (mr *MockQueryAPIMockRecorder) QueryWithParams(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryWithParams", reflect.TypeOf((*MockQueryAPI)(nil).QueryWithParams), arg0, arg1, arg2)
}
