package ha

import (
	"context"

	"github.com/cortexproject/cortex/pkg/ring"
)

// lazyTransferer lazily implements FlushTransferer, allowing to defer the creation of an actual
// FlushTransferer to after a ring.NewLifecycler is called.
type lazyTransferer struct {
	inner ring.FlushTransferer
}

func (t *lazyTransferer) Flush() {
	if t.inner != nil {
		t.Flush()
	}
}

func (t *lazyTransferer) TransferOut(ctx context.Context) error {
	if t.inner != nil {
		return t.inner.TransferOut(ctx)
	}

	return nil
}
