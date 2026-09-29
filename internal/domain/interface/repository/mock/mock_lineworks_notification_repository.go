package mock

import (
	context "context"
	reflect "reflect"
	lineworks "solvi/internal/domain/entity/lineworks"
	time "time"

	gomock "go.uber.org/mock/gomock"
)

type MockLineWorksNotificationRepository struct {
	ctrl     *gomock.Controller
	recorder *MockLineWorksNotificationRepositoryMockRecorder
	isgomock struct{}
}

type MockLineWorksNotificationRepositoryMockRecorder struct {
	mock *MockLineWorksNotificationRepository
}

func NewMockLineWorksNotificationRepository(ctrl *gomock.Controller) *MockLineWorksNotificationRepository {
	mock := &MockLineWorksNotificationRepository{ctrl: ctrl}
	mock.recorder = &MockLineWorksNotificationRepositoryMockRecorder{mock}
	return mock
}

func (m *MockLineWorksNotificationRepository) EXPECT() *MockLineWorksNotificationRepositoryMockRecorder {
	return m.recorder
}

func (m *MockLineWorksNotificationRepository) ClaimDue(ctx context.Context, now time.Time, limit int) ([]lineworks.Notification, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ClaimDue", ctx, now, limit)
	ret0, _ := ret[0].([]lineworks.Notification)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockLineWorksNotificationRepositoryMockRecorder) ClaimDue(ctx, now, limit any) *MockLineWorksNotificationRepositoryClaimDueCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ClaimDue", reflect.TypeOf((*MockLineWorksNotificationRepository)(nil).ClaimDue), ctx, now, limit)
	return &MockLineWorksNotificationRepositoryClaimDueCall{Call: call}
}

type MockLineWorksNotificationRepositoryClaimDueCall struct{ *gomock.Call }

func (c *MockLineWorksNotificationRepositoryClaimDueCall) Return(arg0 []lineworks.Notification, arg1 error) *MockLineWorksNotificationRepositoryClaimDueCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

func (c *MockLineWorksNotificationRepositoryClaimDueCall) DoAndReturn(f func(context.Context, time.Time, int) ([]lineworks.Notification, error)) *MockLineWorksNotificationRepositoryClaimDueCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

func (m *MockLineWorksNotificationRepository) RecoverStale(ctx context.Context, staleBefore time.Time, maxAttempts int) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RecoverStale", ctx, staleBefore, maxAttempts)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockLineWorksNotificationRepositoryMockRecorder) RecoverStale(ctx, staleBefore, maxAttempts any) *MockLineWorksNotificationRepositoryRecoverStaleCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecoverStale", reflect.TypeOf((*MockLineWorksNotificationRepository)(nil).RecoverStale), ctx, staleBefore, maxAttempts)
	return &MockLineWorksNotificationRepositoryRecoverStaleCall{Call: call}
}

type MockLineWorksNotificationRepositoryRecoverStaleCall struct{ *gomock.Call }

func (c *MockLineWorksNotificationRepositoryRecoverStaleCall) Return(arg0 int, arg1 error) *MockLineWorksNotificationRepositoryRecoverStaleCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

func (m *MockLineWorksNotificationRepository) SaveResult(ctx context.Context, notice *lineworks.Notification, attempt *lineworks.NotificationAttempt) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveResult", ctx, notice, attempt)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockLineWorksNotificationRepositoryMockRecorder) SaveResult(ctx, notice, attempt any) *MockLineWorksNotificationRepositorySaveResultCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveResult", reflect.TypeOf((*MockLineWorksNotificationRepository)(nil).SaveResult), ctx, notice, attempt)
	return &MockLineWorksNotificationRepositorySaveResultCall{Call: call}
}

type MockLineWorksNotificationRepositorySaveResultCall struct{ *gomock.Call }

func (c *MockLineWorksNotificationRepositorySaveResultCall) Return(arg0 error) *MockLineWorksNotificationRepositorySaveResultCall {
	c.Call = c.Call.Return(arg0)
	return c
}

func (c *MockLineWorksNotificationRepositorySaveResultCall) DoAndReturn(f func(context.Context, *lineworks.Notification, *lineworks.NotificationAttempt) error) *MockLineWorksNotificationRepositorySaveResultCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

func (m *MockLineWorksNotificationRepository) EnqueueDueDigests(ctx context.Context, notices []lineworks.Notification) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EnqueueDueDigests", ctx, notices)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockLineWorksNotificationRepositoryMockRecorder) EnqueueDueDigests(ctx, notices any) *MockLineWorksNotificationRepositoryEnqueueDueDigestsCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EnqueueDueDigests", reflect.TypeOf((*MockLineWorksNotificationRepository)(nil).EnqueueDueDigests), ctx, notices)
	return &MockLineWorksNotificationRepositoryEnqueueDueDigestsCall{Call: call}
}

type MockLineWorksNotificationRepositoryEnqueueDueDigestsCall struct{ *gomock.Call }

func (c *MockLineWorksNotificationRepositoryEnqueueDueDigestsCall) Return(arg0 error) *MockLineWorksNotificationRepositoryEnqueueDueDigestsCall {
	c.Call = c.Call.Return(arg0)
	return c
}

func (c *MockLineWorksNotificationRepositoryEnqueueDueDigestsCall) DoAndReturn(f func(context.Context, []lineworks.Notification) error) *MockLineWorksNotificationRepositoryEnqueueDueDigestsCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}
