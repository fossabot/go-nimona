package fabric

import mock "github.com/stretchr/testify/mock"

// MockUPNP is an autogenerated mock type for the UPNP type
type MockUPNP struct {
	mock.Mock
}

// Clear provides a mock function with given fields: port
func (_m *MockUPNP) Clear(port uint16) error {
	ret := _m.Called(port)

	var r0 error
	if rf, ok := ret.Get(0).(func(uint16) error); ok {
		r0 = rf(port)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ExternalIP provides a mock function with given fields:
func (_m *MockUPNP) ExternalIP() (string, error) {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Forward provides a mock function with given fields: port, desc
func (_m *MockUPNP) Forward(port uint16, desc string) error {
	ret := _m.Called(port, desc)

	var r0 error
	if rf, ok := ret.Get(0).(func(uint16, string) error); ok {
		r0 = rf(port, desc)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Location provides a mock function with given fields:
func (_m *MockUPNP) Location() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}
