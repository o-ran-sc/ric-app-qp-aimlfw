package mocks

import "gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"

type FakeRMRClient struct{}

func NewFakeRMRClient() FakeRMRClient {
	return FakeRMRClient{}
}

func (frw FakeRMRClient) SendRts(msg *xapp.RMRParams) bool {
	return true
}

func (rw FakeRMRClient) GetRicMessageName(id int) string {
	return "ricMessageName"
}
