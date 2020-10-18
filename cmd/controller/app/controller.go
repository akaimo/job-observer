package app

import "context"

func Run(stopCh <-chan struct{}) {
	rootCtx := ContextWithStopCh(context.Background(), stopCh)
}

func ContextWithStopCh(ctx context.Context, stopCh <-chan struct{}) context.Context {
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		defer cancel()
		select {
		case <-ctx.Done():
		case <-stopCh:
		}
	}()
	return ctx
}
