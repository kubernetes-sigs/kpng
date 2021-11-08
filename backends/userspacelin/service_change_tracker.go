package userspacelin

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/events"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/proxy/metrics"
	"sigs.k8s.io/kpng/pkg/api/localnetv1"
)

type rajasServiceChange struct {
	current  *localnetv1.Service
	previous *localnetv1.Service
}

type RajasServiceChangeTracker struct {
	// items maps a service to its serviceChangee (which is just a map[servicePortName]servicePort)
	items map[types.NamespacedName]*rajasServiceChange

	// processServiceMapChange processServiceMapChangeFunc
	ipFamily v1.IPFamily

	recorder events.EventRecorder
}

// Update updates given service's change map based on the <previous, current> service pair.  It returns true if items changed,
// otherwise return false.  Update can be used to add/update/delete items of ServiceChangeMap.  For example,
// Add item
//   - pass <nil, service> as the <previous, current> pair.
// Update item
//   - pass <oldService, service> as the <previous, current> pair.
// Delete item
//   - pass <service, nil> as the <previous, current> pair.
func (sct *RajasServiceChangeTracker) Update(current *localnetv1.Service) bool {
	svc := current
	if svc == nil {
		return false
	}
	metrics.ServiceChangesTotal.Inc()
	namespacedName := types.NamespacedName{Namespace: svc.Namespace, Name: svc.Name}
	var change *rajasServiceChange
	var ok bool
	if change, ok = sct.items[namespacedName]; !ok {
		change = &rajasServiceChange{}
		sct.items[namespacedName] = change
	}

	rcc := &rajasServiceChange{
		previous: sct.items[namespacedName].current,
		current:  current,
	}
	// TODO make sure i did the pointers right here?
	*change = *rcc

	// *change = sct.serviceToServiceMap(current)
	// klog.V(2).Infof("Service %s updated: %d ports", namespacedName, len(*change))
	metrics.ServiceChangesPending.Set(float64(len(sct.items)))
	return len(sct.items) > 0
}

func (sct *RajasServiceChangeTracker) Delete(namespace, name string) bool {
	metrics.ServiceChangesTotal.Inc()
	namespacedName := types.NamespacedName{Namespace: namespace, Name: name}
	sct.items[namespacedName] = nil
	klog.V(2).Infof("Service %s updated for delete", namespacedName)
	metrics.ServiceChangesPending.Set(float64(len(sct.items)))
	return len(sct.items) > 0
}

// ServicePortName carries a namespace + name + portname.  This is the unique
// identifier for a load-balanced service.
type ServicePortName struct {
	types.NamespacedName
	Port     string
	Protocol localnetv1.Protocol
	PortName string // FYI Jay added this, because we needed it for the BuildPortsToEndpointsMap function by KPNG
}

/**
// serviceToServiceMap translates a single Service object to a ServiceMap.
//
// NOTE: service object should NOT be modified.
func (sct *RajasServiceChangeTracker) serviceToServiceMap(service *localnetv1.Service) rajasServiceChange {
	if service == nil {
		return nil
	}
	clusterIP := GetClusterIPByFamily(sct.ipFamily, service)
	if clusterIP == "" {
		return nil
	}
	serviceMap := make(rajasServiceChange)
	svcName := types.NamespacedName{Namespace: service.Namespace, Name: service.Name}
	for i := range service.Ports {
		servicePort := service.Ports[i]
		svcPortName := ServicePortName{NamespacedName: svcName, Port: servicePort.Name, Protocol: servicePort.Protocol}
		baseSvcInfo := sct.newBaseServiceInfo(servicePort, service)
		if sct.makeServiceInfo != nil {
			serviceMap[svcPortName] = sct.makeServiceInfo(servicePort, service, baseSvcInfo)
		} else {
			serviceMap[svcPortName] = baseSvcInfo
		}
	}
	return serviceMap
}
**/
