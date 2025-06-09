package collectorprovider

import (
	"context"

	"go.opentelemetry.io/collector/confmap"
	"go.uber.org/zap"
)

type fakeConfmapProvider struct {
	scheme string
	ret    func(ctx context.Context, uri string, watcher confmap.WatcherFunc) (*confmap.Retrieved, error)
	logger *zap.Logger
}

func (f *fakeConfmapProvider) Retrieve(ctx context.Context, uri string, watcher confmap.WatcherFunc) (*confmap.Retrieved, error) {
	return f.ret(ctx, uri, watcher)
}

func (f *fakeConfmapProvider) Scheme() string {
	return f.scheme
}

func (f *fakeConfmapProvider) Shutdown(context.Context) error {
	return nil
}

func newFakeConfmapProvider(scheme string, ret func(ctx context.Context, uri string, watcher confmap.WatcherFunc) (*confmap.Retrieved, error)) confmap.ProviderFactory {
	return confmap.NewProviderFactory(func(ps confmap.ProviderSettings) confmap.Provider {
		return &fakeConfmapProvider{
			scheme: scheme,
			ret:    ret,
			logger: ps.Logger,
		}
	})
}
