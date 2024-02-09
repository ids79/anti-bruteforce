package app

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIPList(t *testing.T) {
	t.Run("test GetIPRange", func(t *testing.T) {
		ip := "10.10.10.10"
		mask := "25"
		r, err := GetIPRange(ip, mask)
		require.Nil(t, err)
		require.Equal(t, r.IP, 168430090)
		require.Equal(t, r.Mask, 25)
		require.Equal(t, r.IPfrom, 168430080)
		require.Equal(t, r.IPto, 168430207)
		ip = "10.10.10.13"
		mask = "30"
		r, err = GetIPRange(ip, mask)
		require.Nil(t, err)
		require.Equal(t, r.IP, 168430093)
		require.Equal(t, r.Mask, 30)
		require.Equal(t, r.IPfrom, 168430092)
		require.Equal(t, r.IPto, 168430095)
	})
	t.Run("test IPtoStr", func(t *testing.T) {
		ip := 168430091
		str := IPtoStr(ip)
		require.Equal(t, str, "10.10.10.11")
		ip = 1
		str = IPtoStr(ip)
		require.Equal(t, str, "0.0.0.1")
	})
}
