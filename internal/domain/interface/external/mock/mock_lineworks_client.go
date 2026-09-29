package mock

import (
	context "context"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

type MockLineWorksClient struct {
	ctrl     *gomock.Controller
	recorder *MockLineWorksClientMockRecorder
	isgomock struct{}
}

type MockLineWorksClientMockRecorder struct {
	mock *MockLineWorksClient
}

func NewMockLineWorksClient(ctrl *gomock.Controller) *MockLineWorksClient {
	mock := &MockLineWorksClient{ctrl: ctrl}
	mock.recorder = &MockLineWorksClientMockRecorder{mock}
	return mock
}

func (m *MockLineWorksClient) EXPECT() *MockLineWorksClientMockRecorder {
	return m.recorder
}

func (m *MockLineWorksClient) SendChannelMessage(ctx context.Context, channelID, text string) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendChannelMessage", ctx, channelID, text)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockLineWorksClientMockRecorder) SendChannelMessage(ctx, channelID, text any) *MockLineWorksClientSendChannelMessageCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendChannelMessage", reflect.TypeOf((*MockLineWorksClient)(nil).SendChannelMessage), ctx, channelID, text)
	return &MockLineWorksClientSendChannelMessageCall{Call: call}
}

type MockLineWorksClientSendChannelMessageCall struct{ *gomock.Call }

func (c *MockLineWorksClientSendChannelMessageCall) Return(arg0 int, arg1 error) *MockLineWorksClientSendChannelMessageCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

func (m *MockLineWorksClient) SendUserMessage(ctx context.Context, userID, text string) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendUserMessage", ctx, userID, text)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockLineWorksClientMockRecorder) SendUserMessage(ctx, userID, text any) *MockLineWorksClientSendUserMessageCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendUserMessage", reflect.TypeOf((*MockLineWorksClient)(nil).SendUserMessage), ctx, userID, text)
	return &MockLineWorksClientSendUserMessageCall{Call: call}
}

type MockLineWorksClientSendUserMessageCall struct{ *gomock.Call }

func (c *MockLineWorksClientSendUserMessageCall) Return(arg0 int, arg1 error) *MockLineWorksClientSendUserMessageCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}
