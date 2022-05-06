// Code generated by mockery v2.12.0. DO NOT EDIT.

package acl

import (
	testing "testing"

	mock "github.com/stretchr/testify/mock"
)

// MockTokenWriter is an autogenerated mock type for the TokenWriter type
type MockTokenWriter struct {
	mock.Mock
}

// Delete provides a mock function with given fields: secretID, fromLogout
func (_m *MockTokenWriter) Delete(secretID string, fromLogout bool) error {
	ret := _m.Called(secretID, fromLogout)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, bool) error); ok {
		r0 = rf(secretID, fromLogout)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewMockTokenWriter creates a new instance of MockTokenWriter. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockTokenWriter(t testing.TB) *MockTokenWriter {
	mock := &MockTokenWriter{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}