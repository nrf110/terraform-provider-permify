package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	permifytest "github.com/theoriginalstove/testcontainers-permify"
)

func initPermify(t *testing.T) string {
	ctx := context.Background()
	container, err := permifytest.Run(ctx)
	require.NoError(t, err)
	t.Cleanup(func() {
		err = container.Terminate(ctx)
		require.NoError(t, err, "failed to terminate container")
	})

	host, err := container.Host(ctx)
	require.NoError(t, err)
	port, err := container.GRPCPort(ctx)
	require.NoError(t, err)

	return fmt.Sprintf(`
	provider "permify" {
		endpoint = "%s:%d"
	}
	`, host, port)
}
