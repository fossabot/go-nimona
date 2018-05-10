// Code generated by mockery v1.0.0. DO NOT EDIT.
package mesh

import mock "github.com/stretchr/testify/mock"
import net "net"

// MockHandler is an autogenerated mock type for the Handler type
type MockHandler struct {
	mock.Mock
}

// Handle provides a mock function with given fields: _a0
func (_m *MockHandler) Handle(_a0 net.Conn) (net.Conn, error) {
	ret := _m.Called(_a0)

	var r0 net.Conn
	if rf, ok := ret.Get(0).(func(net.Conn) net.Conn); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(net.Conn)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(net.Conn) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Initiate provides a mock function with given fields: _a0
func (_m *MockHandler) Initiate(_a0 net.Conn) (net.Conn, error) {
	ret := _m.Called(_a0)

	var r0 net.Conn
	if rf, ok := ret.Get(0).(func(net.Conn) net.Conn); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(net.Conn)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(net.Conn) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}