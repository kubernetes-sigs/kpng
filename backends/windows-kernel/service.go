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

package winkernel

import (
	"k8s.io/kubernetes/pkg/proxy"
	"k8s.io/klog/v2"
)

// internal struct for string service information
type serviceInfo struct {
        *proxy.BaseServiceInfo
        targetPort             int
        externalIPs            []*externalIPInfo
        loadBalancerIngressIPs []*loadBalancerIngressInfo
        hnsID                  string
        nodePorthnsID          string
        policyApplied          bool
        remoteEndpoint         *endpoints
        hns                    HostNetworkService
        preserveDIP            bool
        localTrafficDSR        bool
}


func (svcInfo *serviceInfo) deleteAllHnsLoadBalancerPolicy() {
        // Remove the Hns Policy corresponding to this service
        hns := svcInfo.hns
        hns.deleteLoadBalancer(svcInfo.hnsID)
        svcInfo.hnsID = ""

        hns.deleteLoadBalancer(svcInfo.nodePorthnsID)
        svcInfo.nodePorthnsID = ""

        for _, externalIP := range svcInfo.externalIPs {
                hns.deleteLoadBalancer(externalIP.hnsID)
                externalIP.hnsID = ""
        }
        for _, lbIngressIP := range svcInfo.loadBalancerIngressIPs {
                hns.deleteLoadBalancer(lbIngressIP.hnsID)
                lbIngressIP.hnsID = ""
        }
}

func (svcInfo *serviceInfo) cleanupAllPolicies(proxyEndpoints []proxy.Endpoint) {
        klog.V(3).InfoS("Service cleanup", "serviceInfo", svcInfo)
        // Skip the svcInfo.policyApplied check to remove all the policies
        svcInfo.deleteAllHnsLoadBalancerPolicy()
        // Cleanup Endpoints references
        for _, ep := range proxyEndpoints {
                epInfo, ok := ep.(*endpoints)
                if ok {
                        epInfo.Cleanup()
                }
        }
        if svcInfo.remoteEndpoint != nil {
                svcInfo.remoteEndpoint.Cleanup()
        }

        svcInfo.policyApplied = false
}
