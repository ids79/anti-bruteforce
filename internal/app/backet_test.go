package app

import (
	"context"
	"testing"
	"time"

	"github.com/ids79/anti-bruteforcer/internal/config"
	"github.com/ids79/anti-bruteforcer/internal/logger"
	"github.com/stretchr/testify/require"
)

func TestBackets(t *testing.T) {
	ctx := context.Background()
	config := config.NewConfig("../../configs/config.toml")
	config.ExpireLimit = 1
	logg := logger.New(config.Logger, "BruteForce:")
	backets := NewBackets(ctx, logg, &config)

	t.Run("test access verification", func(t *testing.T) {
		for i := 1; i < config.Limits.K; i++ {
			chek, _ := backets.AccessVerification("127.0.0.10", IP)
			require.True(t, chek)

			if i < config.Limits.N {
				chek, _ = backets.AccessVerification("ids", LOGIN)
				require.True(t, chek)
			}
			if i < config.Limits.M {
				chek, _ = backets.AccessVerification("123", PASSWORD)
				require.True(t, chek)
			}
		}
		chek, _ := backets.AccessVerification("127.0.0.10", IP)
		require.False(t, chek)
		chek, _ = backets.AccessVerification("ids", LOGIN)
		require.False(t, chek)
		chek, _ = backets.AccessVerification("123", PASSWORD)
		require.False(t, chek)
		time.Sleep(time.Second * 2)
		chek, _ = backets.AccessVerification("127.0.0.10", IP)
		require.True(t, chek)
		chek, _ = backets.AccessVerification("ids", LOGIN)
		require.True(t, chek)
		chek, _ = backets.AccessVerification("123", PASSWORD)
		require.True(t, chek)
	})

	t.Run("test reset backet", func(t *testing.T) {
		backets.ResetBacket("127.0.0.10", IP)
		backets.ResetBacket("ids", LOGIN)
		backets.ResetBacket("123", PASSWORD)
		for i := 1; i < config.Limits.K; i++ {
			chek, _ := backets.AccessVerification("127.0.0.10", IP)
			require.True(t, chek)

			if i < config.Limits.N {
				chek, _ = backets.AccessVerification("ids", LOGIN)
				require.True(t, chek)
			}
			if i < config.Limits.M {
				chek, _ = backets.AccessVerification("123", PASSWORD)
				require.True(t, chek)
			}
		}
		chek, _ := backets.AccessVerification("127.0.0.10", IP)
		require.False(t, chek)
		chek, _ = backets.AccessVerification("ids", LOGIN)
		require.False(t, chek)
		chek, _ = backets.AccessVerification("123", PASSWORD)
		require.False(t, chek)
		backets.ResetBacket("127.0.0.10", IP)
		backets.ResetBacket("ids", LOGIN)
		backets.ResetBacket("123", PASSWORD)
		chek, _ = backets.AccessVerification("127.0.0.10", IP)
		require.True(t, chek)
		chek, _ = backets.AccessVerification("ids", LOGIN)
		require.True(t, chek)
		chek, _ = backets.AccessVerification("123", PASSWORD)
		require.True(t, chek)
	})
}
