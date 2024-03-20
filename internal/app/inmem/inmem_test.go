package inmem

import (
	"context"
	"testing"
	"time"

	"github.com/ids79/anti-bruteforce/internal/config"
	"github.com/ids79/anti-bruteforce/internal/logger"
	"github.com/stretchr/testify/require"
)

func TestBackets(t *testing.T) {
	ctx := context.Background()
	config := &config.Config{
		Logger: config.LoggerConf{
			Level:       "INFO",
			LogEncoding: "console",
		},
		ExpireLimit:  1,
		TickInterval: 500,
		Limits: config.BruteforceLimits{
			TryForLogin: 10,
			TryForPass:  100,
			TreyForIP:   1000,
		},
	}
	logg := logger.New(config.Logger, "BruteForce:")
	backets := NewBackets(ctx, logg, config)

	t.Run("test access verification", func(t *testing.T) {
		for i := 1; i < config.Limits.TreyForIP; i++ {
			chek, err := backets.AccessVerification("127.0.0.10", IP)
			require.NoError(t, err)
			require.True(t, chek)

			if i < config.Limits.TryForLogin {
				chek, err = backets.AccessVerification("ids", LOGIN)
				require.NoError(t, err)
				require.True(t, chek)
			}
			if i < config.Limits.TryForPass {
				chek, err = backets.AccessVerification("123", PASSWORD)
				require.NoError(t, err)
				require.True(t, chek)
			}
		}
		chek, err := backets.AccessVerification("127.0.0.10", IP)
		require.NoError(t, err)
		require.False(t, chek)
		chek, err = backets.AccessVerification("ids", LOGIN)
		require.NoError(t, err)
		require.False(t, chek)
		chek, err = backets.AccessVerification("123", PASSWORD)
		require.NoError(t, err)
		require.False(t, chek)
		time.Sleep(time.Second * 2)
		chek, err = backets.AccessVerification("127.0.0.10", IP)
		require.NoError(t, err)
		require.True(t, chek)
		chek, err = backets.AccessVerification("ids", LOGIN)
		require.NoError(t, err)
		require.True(t, chek)
		chek, err = backets.AccessVerification("123", PASSWORD)
		require.NoError(t, err)
		require.True(t, chek)
	})

	t.Run("test reset backet", func(t *testing.T) {
		backets.ResetBacket(ctx, "127.0.0.10", IP)
		backets.ResetBacket(ctx, "ids", LOGIN)
		backets.ResetBacket(ctx, "123", PASSWORD)
		for i := 1; i < config.Limits.TreyForIP; i++ {
			chek, err := backets.AccessVerification("127.0.0.10", IP)
			require.Nil(t, err)
			require.True(t, chek)

			if i < config.Limits.TryForLogin {
				chek, err = backets.AccessVerification("ids", LOGIN)
				require.Nil(t, err)
				require.True(t, chek)
			}
			if i < config.Limits.TryForPass {
				chek, err = backets.AccessVerification("123", PASSWORD)
				require.Nil(t, err)
				require.True(t, chek)
			}
		}
		chek, err := backets.AccessVerification("127.0.0.10", IP)
		require.Nil(t, err)
		require.False(t, chek)
		chek, err = backets.AccessVerification("ids", LOGIN)
		require.Nil(t, err)
		require.False(t, chek)
		chek, err = backets.AccessVerification("123", PASSWORD)
		require.Nil(t, err)
		require.False(t, chek)
		backets.ResetBacket(ctx, "127.0.0.10", IP)
		backets.ResetBacket(ctx, "ids", LOGIN)
		backets.ResetBacket(ctx, "123", PASSWORD)
		chek, err = backets.AccessVerification("127.0.0.10", IP)
		require.Nil(t, err)
		require.True(t, chek)
		chek, err = backets.AccessVerification("ids", LOGIN)
		require.Nil(t, err)
		require.True(t, chek)
		chek, err = backets.AccessVerification("123", PASSWORD)
		require.Nil(t, err)
		require.True(t, chek)
	})
}
