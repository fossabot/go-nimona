// Code generated by mockery v1.0.0
package mesh

import context "context"
import mock "github.com/stretchr/testify/mock"
import net "github.com/nimona/go-nimona/net"

// MockMesh is an autogenerated mock type for the Mesh type
type MockMesh struct {
	mock.Mock
}

// Dial provides a mock function with given fields: ctx, peerID, protocol
func (_m *MockMesh) Dial(ctx context.Context, peerID string, protocol string) (context.Context, net.Conn, error) {
	ret := _m.Called(ctx, peerID, protocol)

	var r0 context.Context
	if rf, ok := ret.Get(0).(func(context.Context, string, string) context.Context); ok {
		r0 = rf(ctx, peerID, protocol)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(context.Context)
		}
	}

	var r1 net.Conn
	if rf, ok := ret.Get(1).(func(context.Context, string, string) net.Conn); ok {
		r1 = rf(ctx, peerID, protocol)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(net.Conn)
		}
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(context.Context, string, string) error); ok {
		r2 = rf(ctx, peerID, protocol)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// GetLocalPeerInfo provides a mock function with given fields:
func (_m *MockMesh) GetLocalPeerInfo() PeerInfo {
	ret := _m.Called()

	var r0 PeerInfo
	if rf, ok := ret.Get(0).(func() PeerInfo); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(PeerInfo)
	}

	return r0
}

// GetPeerInfo provides a mock function with given fields: ctx, peerID
func (_m *MockMesh) GetPeerInfo(ctx context.Context, peerID string) (PeerInfo, error) {
	ret := _m.Called(ctx, peerID)

	var r0 PeerInfo
	if rf, ok := ret.Get(0).(func(context.Context, string) PeerInfo); ok {
		r0 = rf(ctx, peerID)
	} else {
		r0 = ret.Get(0).(PeerInfo)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, peerID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Publish provides a mock function with given fields: msg, topic
func (_m *MockMesh) Publish(msg interface{}, topic string) error {
	ret := _m.Called(msg, topic)

	var r0 error
	if rf, ok := ret.Get(0).(func(interface{}, string) error); ok {
		r0 = rf(msg, topic)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Subscribe provides a mock function with given fields: topic
func (_m *MockMesh) Subscribe(topic string) (chan interface{}, error) {
	ret := _m.Called(topic)

	var r0 chan interface{}
	if rf, ok := ret.Get(0).(func(string) chan interface{}); ok {
		r0 = rf(topic)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(chan interface{})
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(topic)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Unsubscribe provides a mock function with given fields: _a0
func (_m *MockMesh) Unsubscribe(_a0 chan interface{}) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(chan interface{}) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}