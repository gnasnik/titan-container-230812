package manager

import (
	"context"
	"github.com/filecoin-project/go-jsonrpc"
	"github.com/filecoin-project/go-jsonrpc/auth"
	"github.com/gnasnik/titan-container/api"
	"github.com/gnasnik/titan-container/api/client"
	"golang.org/x/xerrors"
	"net/http"
)

type remoteProvider struct {
	api.Provider
	closer jsonrpc.ClientCloser
}

func connectRemoteProvider(ctx context.Context, fa api.Common, url string) (*remoteProvider, error) {
	token, err := fa.AuthNew(ctx, []auth.Permission{"admin"})
	if err != nil {
		return nil, xerrors.Errorf("creating auth token for remote connection: %w", err)
	}
	headers := http.Header{}
	headers.Add("Authorization", "Bearer "+string(token))

	papi, closer, err := client.NewProvider(context.TODO(), url, headers)
	if err != nil {
		return nil, xerrors.Errorf("creating jsonrpc client: %w", err)
	}

	ver, err := papi.Version(ctx)
	if err != nil {
		closer()
		return nil, err
	}

	if !ver.EqMajorMinor(api.ProviderAPIVersion0) {
		return nil, xerrors.Errorf("unsupported provider api version: %s (expected %s)", ver, api.ProviderAPIVersion0)
	}

	return &remoteProvider{papi, closer}, nil
}

func (r *remoteProvider) Close() error {
	r.closer()
	return nil
}

var _ api.Provider = &remoteProvider{}
