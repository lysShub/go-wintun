package wintun_test

import (
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/lysShub/wintun-go"
	"github.com/stretchr/testify/require"
)

func Test_Adapter_InterfaceIndex(t *testing.T) {
	tun, err := wintun.LoadWintun(wintun.DLL)
	require.NoError(t, err)
	defer tun.Close()

	name := "testadapterinterfaceindex"

	a, err := tun.CreateAdapter(name, "", nil)
	require.NoError(t, err)
	defer a.Close()

	nicid, err := a.InterfaceIndex()
	require.NoError(t, err)

	b, err := exec.Command("netsh", "int", "ipv4", "show", "interfaces").CombinedOutput()
	require.NoError(t, err)

	for _, line := range strings.Split(string(b), "\n") {
		if strings.Contains(line, name) {
			require.True(t, strings.Contains(line, strconv.Itoa(nicid)))
			return
		}
	}
	t.Errorf("can't found nic: \n %s", string(b))
}

func Test_Aapter_Address(t *testing.T) {

	tun, err := wintun.LoadWintun(wintun.DLL)
	require.NoError(t, err)
	defer tun.Close()

	name := "testadapteraddress"

	a, err := tun.CreateAdapter(name, "", nil)
	require.NoError(t, err)
	defer a.Close()

	se, err := a.StartSession(wintun.MinRingCapacity)
	require.NoError(t, err)
	defer se.Close()
	time.Sleep(time.Second * 10) // wait DHCP

	_, err = a.Addresses()
	require.NoError(t, err)
}

func TestWintun(t *testing.T) {

	tun, err := wintun.LoadWintun(wintun.DLL)
	require.NoError(t, err)
	defer tun.Close()

	_, err = tun.DriverVersion()
	require.Equal(t, syscall.ERROR_FILE_NOT_FOUND, err)

	// guid := windows.GUID{
	// 	Data1: 0xdeadbabe,
	// 	Data2: 0xcafe,
	// 	Data3: 0xbeef,
	// 	Data4: [8]byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef},
	// }

	a, err := tun.CreateAdapter("test", "example", nil)
	require.NoError(t, err)
	defer a.Close()

	v, err := tun.DriverVersion()
	require.NoError(t, err)
	require.NotZero(t, v)

	{
		ifIdx, err := a.InterfaceIndex()
		require.NoError(t, err)
		i, err := net.InterfaceByIndex(ifIdx)
		require.NoError(t, err)
		addrs, err := i.Addrs()
		require.NoError(t, err)
		require.Equal(t, 1, len(addrs))
		require.Equal(t, "ip+net", addrs[0].Network())
		fmt.Println(addrs[0].String())
	}

	s, err := a.StartSession(wintun.MinRingCapacity)
	require.NoError(t, err)
	defer s.Close()

	{
		time.Sleep(time.Second * 20) // wait DHCP allocates IP addresses

		ifIdx, err := a.InterfaceIndex()
		require.NoError(t, err)
		i, err := net.InterfaceByIndex(ifIdx)
		require.NoError(t, err)
		addrs, err := i.Addrs()
		require.NoError(t, err)
		require.Equal(t, 2, len(addrs))
		// require.Equal(t, "ip", addrs[0].Network())
		fmt.Println(addrs[0].String())
		fmt.Println(addrs[1].String())
	}

	for {
		p, err := s.ReceivePacket()
		require.NoError(t, err)
		s.ReleasePacket(p)
	}

}
