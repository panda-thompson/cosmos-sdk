package serverv2_app

import (
	"context"
	"testing"

	counterv1 "cosmossdk.io/api/cosmos/counter/v1"
	"github.com/stretchr/testify/require"
)

func TestBuild(t *testing.T) {
	stf, err := NewSTF()
	require.NoError(t, err)
	ctx := stf.MakeContext(context.Background(), []byte("sender"), nil, 50_000_00)
	resp, err := stf.Handle(ctx, []byte("sender"), &counterv1.MsgEcho{
		Echo: "echo",
	})
	require.NoError(t, err)
	t.Logf("%s", resp)
}
