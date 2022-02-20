//go:build windows
// +build windows

/*
Copyright 2018-2022 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package kernelspace

import (
	"fmt"
	"net"

	"github.com/Microsoft/hcsshim"

	v1 "k8s.io/api/core/v1"
)

type externalIPInfo struct {
	ip    string
	hnsID string
}

type loadBalancerIngressInfo struct {
	ip    string
	hnsID string
}

func Enum(p v1.Protocol) uint16 {
	if p == v1.ProtocolTCP {
		return 6
	}
	if p == v1.ProtocolUDP {
		return 17
	}
	if p == v1.ProtocolSCTP {
		return 132
	}
	return 0
}

type closeable interface {
	Close() error
}

//Uses mac prefix and IPv4 address to return a mac address
//This ensures mac addresses are unique for proper load balancing
//There is a possibility of MAC collisions but this Mac address is used for remote windowsEndpoint only
//and not sent on the wire.
func conjureMac(macPrefix string, ip net.IP) string {
	if ip4 := ip.To4(); ip4 != nil {
		a, b, c, d := ip4[0], ip4[1], ip4[2], ip4[3]
		return fmt.Sprintf("%v-%02x-%02x-%02x-%02x", macPrefix, a, b, c, d)
	} else if ip6 := ip.To16(); ip6 != nil {
		a, b, c, d := ip6[15], ip6[14], ip6[13], ip6[12]
		return fmt.Sprintf("%v-%02x-%02x-%02x-%02x", macPrefix, a, b, c, d)
	}
	return "02-11-22-33-44-55"
}

// KernelCompatTester tests whether the required kernel capabilities are
// present to run the windows kernel proxier.
type KernelCompatTester interface {
	IsCompatible() error
}

// CanUseWinKernelProxier returns true if we should use the Kernel Proxier
// instead of the "classic" userspace Proxier.  This is determined by checking
// the windows kernel version and for the existence of kernel features.
func CanUseWinKernelProxier(kcompat KernelCompatTester) (bool, error) {
	// Check that the kernel supports what we need.
	if err := kcompat.IsCompatible(); err != nil {
		return false, err
	}
	return true, nil
}

type WindowsKernelCompatTester struct{}

// IsCompatible returns true if winkernel can support this mode of proxy
func (lkct WindowsKernelCompatTester) IsCompatible() error {
	_, err := hcsshim.HNSListPolicyListRequest()
	if err != nil {
		return fmt.Errorf("Windows kernel is not compatible for Kernel mode")
	}
	return nil
}