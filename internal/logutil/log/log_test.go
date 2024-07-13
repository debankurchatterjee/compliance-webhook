package log

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFrom(t *testing.T) {
	timeformat := "02-01-2006 15:04:05.000 UTC"
	ctxWithLogger := WithLogger(context.Background(), &timeformat)
	ogLogger := From(ctxWithLogger)

	tests := []struct {
		name string
		ctx  func() context.Context
		want bool
	}{
		{
			name: "with logger",
			ctx: func() context.Context {
				return ctxWithLogger
			},
			want: true,
		},
		{
			name: "without logger",
			ctx: func() context.Context {
				return context.Background()
			},
		},
		{
			name: "nil context",
			ctx: func() context.Context {
				return nil
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			logger := From(tc.ctx())
			if tc.want {
				require.Equal(t, ogLogger, logger)
			} else {
				require.NotEqual(t, ogLogger, logger)
			}
		})
	}

}

func TestWithLogger(t *testing.T) {
	timeformat := "02-01-2006 15:04:05.000 UTC"
	ctx := context.Background()
	newCtx := WithLogger(ctx, &timeformat)

	require.NotNil(t, newCtx.Value(logKey))
}
