// Code generated by MockGen. DO NOT EDIT.
// Source: account/api (interfaces: AccountService_ListStreamServer,TransactionService_ListStreamServer)

// Package server is a generated GoMock package.
package server

import (
	account "account/api"
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	metadata "google.golang.org/grpc/metadata"
)

// MockAccountService_ListStreamServer is a mock of AccountService_ListStreamServer interface.
type MockAccountService_ListStreamServer struct {
	ctrl     *gomock.Controller
	recorder *MockAccountService_ListStreamServerMockRecorder
}

// MockAccountService_ListStreamServerMockRecorder is the mock recorder for MockAccountService_ListStreamServer.
type MockAccountService_ListStreamServerMockRecorder struct {
	mock *MockAccountService_ListStreamServer
}

// NewMockAccountService_ListStreamServer creates a new mock instance.
func NewMockAccountService_ListStreamServer(ctrl *gomock.Controller) *MockAccountService_ListStreamServer {
	mock := &MockAccountService_ListStreamServer{ctrl: ctrl}
	mock.recorder = &MockAccountService_ListStreamServerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockAccountService_ListStreamServer) EXPECT() *MockAccountService_ListStreamServerMockRecorder {
	return m.recorder
}

// Context mocks base method.
func (m *MockAccountService_ListStreamServer) Context() context.Context {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Context")
	ret0, _ := ret[0].(context.Context)
	return ret0
}

// Context indicates an expected call of Context.
func (mr *MockAccountService_ListStreamServerMockRecorder) Context() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Context", reflect.TypeOf((*MockAccountService_ListStreamServer)(nil).Context))
}

// RecvMsg mocks base method.
func (m *MockAccountService_ListStreamServer) RecvMsg(arg0 interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RecvMsg", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// RecvMsg indicates an expected call of RecvMsg.
func (mr *MockAccountService_ListStreamServerMockRecorder) RecvMsg(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecvMsg", reflect.TypeOf((*MockAccountService_ListStreamServer)(nil).RecvMsg), arg0)
}

// Send mocks base method.
func (m *MockAccountService_ListStreamServer) Send(arg0 *account.Account) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Send", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Send indicates an expected call of Send.
func (mr *MockAccountService_ListStreamServerMockRecorder) Send(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Send", reflect.TypeOf((*MockAccountService_ListStreamServer)(nil).Send), arg0)
}

// SendHeader mocks base method.
func (m *MockAccountService_ListStreamServer) SendHeader(arg0 metadata.MD) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendHeader", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendHeader indicates an expected call of SendHeader.
func (mr *MockAccountService_ListStreamServerMockRecorder) SendHeader(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendHeader", reflect.TypeOf((*MockAccountService_ListStreamServer)(nil).SendHeader), arg0)
}

// SendMsg mocks base method.
func (m *MockAccountService_ListStreamServer) SendMsg(arg0 interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendMsg", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendMsg indicates an expected call of SendMsg.
func (mr *MockAccountService_ListStreamServerMockRecorder) SendMsg(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendMsg", reflect.TypeOf((*MockAccountService_ListStreamServer)(nil).SendMsg), arg0)
}

// SetHeader mocks base method.
func (m *MockAccountService_ListStreamServer) SetHeader(arg0 metadata.MD) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetHeader", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetHeader indicates an expected call of SetHeader.
func (mr *MockAccountService_ListStreamServerMockRecorder) SetHeader(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetHeader", reflect.TypeOf((*MockAccountService_ListStreamServer)(nil).SetHeader), arg0)
}

// SetTrailer mocks base method.
func (m *MockAccountService_ListStreamServer) SetTrailer(arg0 metadata.MD) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetTrailer", arg0)
}

// SetTrailer indicates an expected call of SetTrailer.
func (mr *MockAccountService_ListStreamServerMockRecorder) SetTrailer(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetTrailer", reflect.TypeOf((*MockAccountService_ListStreamServer)(nil).SetTrailer), arg0)
}

// MockTransactionService_ListStreamServer is a mock of TransactionService_ListStreamServer interface.
type MockTransactionService_ListStreamServer struct {
	ctrl     *gomock.Controller
	recorder *MockTransactionService_ListStreamServerMockRecorder
}

// MockTransactionService_ListStreamServerMockRecorder is the mock recorder for MockTransactionService_ListStreamServer.
type MockTransactionService_ListStreamServerMockRecorder struct {
	mock *MockTransactionService_ListStreamServer
}

// NewMockTransactionService_ListStreamServer creates a new mock instance.
func NewMockTransactionService_ListStreamServer(ctrl *gomock.Controller) *MockTransactionService_ListStreamServer {
	mock := &MockTransactionService_ListStreamServer{ctrl: ctrl}
	mock.recorder = &MockTransactionService_ListStreamServerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockTransactionService_ListStreamServer) EXPECT() *MockTransactionService_ListStreamServerMockRecorder {
	return m.recorder
}

// Context mocks base method.
func (m *MockTransactionService_ListStreamServer) Context() context.Context {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Context")
	ret0, _ := ret[0].(context.Context)
	return ret0
}

// Context indicates an expected call of Context.
func (mr *MockTransactionService_ListStreamServerMockRecorder) Context() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Context", reflect.TypeOf((*MockTransactionService_ListStreamServer)(nil).Context))
}

// RecvMsg mocks base method.
func (m *MockTransactionService_ListStreamServer) RecvMsg(arg0 interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RecvMsg", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// RecvMsg indicates an expected call of RecvMsg.
func (mr *MockTransactionService_ListStreamServerMockRecorder) RecvMsg(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecvMsg", reflect.TypeOf((*MockTransactionService_ListStreamServer)(nil).RecvMsg), arg0)
}

// Send mocks base method.
func (m *MockTransactionService_ListStreamServer) Send(arg0 *account.Transaction) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Send", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Send indicates an expected call of Send.
func (mr *MockTransactionService_ListStreamServerMockRecorder) Send(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Send", reflect.TypeOf((*MockTransactionService_ListStreamServer)(nil).Send), arg0)
}

// SendHeader mocks base method.
func (m *MockTransactionService_ListStreamServer) SendHeader(arg0 metadata.MD) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendHeader", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendHeader indicates an expected call of SendHeader.
func (mr *MockTransactionService_ListStreamServerMockRecorder) SendHeader(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendHeader", reflect.TypeOf((*MockTransactionService_ListStreamServer)(nil).SendHeader), arg0)
}

// SendMsg mocks base method.
func (m *MockTransactionService_ListStreamServer) SendMsg(arg0 interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendMsg", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendMsg indicates an expected call of SendMsg.
func (mr *MockTransactionService_ListStreamServerMockRecorder) SendMsg(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendMsg", reflect.TypeOf((*MockTransactionService_ListStreamServer)(nil).SendMsg), arg0)
}

// SetHeader mocks base method.
func (m *MockTransactionService_ListStreamServer) SetHeader(arg0 metadata.MD) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetHeader", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetHeader indicates an expected call of SetHeader.
func (mr *MockTransactionService_ListStreamServerMockRecorder) SetHeader(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetHeader", reflect.TypeOf((*MockTransactionService_ListStreamServer)(nil).SetHeader), arg0)
}

// SetTrailer mocks base method.
func (m *MockTransactionService_ListStreamServer) SetTrailer(arg0 metadata.MD) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetTrailer", arg0)
}

// SetTrailer indicates an expected call of SetTrailer.
func (mr *MockTransactionService_ListStreamServerMockRecorder) SetTrailer(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetTrailer", reflect.TypeOf((*MockTransactionService_ListStreamServer)(nil).SetTrailer), arg0)
}
