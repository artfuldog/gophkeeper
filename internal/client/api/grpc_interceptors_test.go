package api

import (
	"context"
	"testing"
)

func TestAuthInterceptor(t *testing.T) {
	ts := NewTestSuiteGRPClient(t)
	ts.Client.config.SetServer("127.0.0.1:3200")
	ts.Client.config.SetUser("username")
	ts.Client.Token = "token1234%^"

	ctx, cancel := context.WithCancel(context.Background())
	ctrlCh := make(chan struct{})
	defer cancel()

	ts.Client.Connect(ctx, ctrlCh)
	ts.Client.UserLogin(ctx, "", "", "")
	ts.Client.GetItem(ctx, "", "")
}
