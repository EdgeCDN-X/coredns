package kubernetes

import (
	"context"
	"net"
	"strings"
	"testing"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/kubernetes/object"
	"github.com/coredns/coredns/request"

	"github.com/miekg/dns"
	api "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	mcs "sigs.k8s.io/mcs-api/pkg/apis/v1alpha1"
)

func TestEndpointHostname(t *testing.T) {
	tests := []struct {
		ip               string
		hostname         string
		expected         string
		podName          string
		endpointNameMode bool
	}{
		{"10.11.12.13", "", "10-11-12-13", "", false},
		{"10.11.12.13", "epname", "epname", "", false},
		{"10.11.12.13", "", "10-11-12-13", "hello-abcde", false},
		{"10.11.12.13", "epname", "epname", "hello-abcde", false},
		{"10.11.12.13", "epname", "epname", "hello-abcde", true},
		{"10.11.12.13", "", "hello-abcde", "hello-abcde", true},
	}
	for _, test := range tests {
		result := endpointHostname(object.EndpointAddress{IP: test.ip, Hostname: test.hostname, TargetRefName: test.podName}, test.endpointNameMode)
		if result != test.expected {
			t.Errorf("Expected endpoint name for (ip:%v hostname:%v) to be '%v', but got '%v'", test.ip, test.hostname, test.expected, result)
		}
	}
}

type APIConnServiceTest struct{}

func (APIConnServiceTest) HasSynced() bool                             { return true }
func (APIConnServiceTest) Run()                                        {}
func (APIConnServiceTest) Stop() error                                 { return nil }
func (APIConnServiceTest) PodIndex(string) []*object.Pod               { return nil }
func (APIConnServiceTest) SvcIndexReverse(string) []*object.Service    { return nil }
func (APIConnServiceTest) SvcExtIndexReverse(string) []*object.Service { return nil }
func (APIConnServiceTest) EpIndexReverse(string) []*object.Endpoints   { return nil }
func (APIConnServiceTest) Modified(ModifiedMode) int64                 { return 0 }

func (APIConnServiceTest) SvcIndex(string) []*object.Service {
	svcs := []*object.Service{
		{
			Name:       "svc1",
			Namespace:  "testns",
			ClusterIPs: []string{"10.0.0.1"},
			Ports: []api.ServicePort{
				{Name: "http", Protocol: "tcp", Port: 80},
			},
		},
		{
			Name:       "svc-dual-stack",
			Namespace:  "testns",
			ClusterIPs: []string{"10.0.0.2", "10::2"},
			Ports: []api.ServicePort{
				{Name: "http", Protocol: "tcp", Port: 80},
			},
		},
		{
			Name:       "hdls1",
			Namespace:  "testns",
			ClusterIPs: []string{api.ClusterIPNone},
		},
		{
			Name:         "external",
			Namespace:    "testns",
			ExternalName: "coredns.io",
			Type:         api.ServiceTypeExternalName,
			Ports: []api.ServicePort{
				{Name: "http", Protocol: "tcp", Port: 80},
			},
		},
	}
	return svcs
}

func (APIConnServiceTest) ServiceList() []*object.Service {
	svcs := []*object.Service{
		{
			Name:       "svc1",
			Namespace:  "testns",
			ClusterIPs: []string{"10.0.0.1"},
			Ports: []api.ServicePort{
				{Name: "http", Protocol: "tcp", Port: 80},
			},
		},
		{
			Name:       "svc-dual-stack",
			Namespace:  "testns",
			ClusterIPs: []string{"10.0.0.2", "10::2"},
			Ports: []api.ServicePort{
				{Name: "http", Protocol: "tcp", Port: 80},
			},
		},
		{
			Name:       "hdls1",
			Namespace:  "testns",
			ClusterIPs: []string{api.ClusterIPNone},
		},
		{
			Name:         "external",
			Namespace:    "testns",
			ExternalName: "coredns.io",
			Type:         api.ServiceTypeExternalName,
			Ports: []api.ServicePort{
				{Name: "http", Protocol: "tcp", Port: 80},
			},
		},
	}
	return svcs
}

func (APIConnServiceTest) EpIndex(string) []*object.Endpoints {
	eps := []*object.Endpoints{
		{
			Subsets: []object.EndpointSubset{
				{
					Addresses: []object.EndpointAddress{
						{IP: "172.0.0.1", Hostname: "ep1a"},
					},
					Ports: []object.EndpointPort{
						{Port: 80, Protocol: "tcp", Name: "http"},
					},
				},
			},
			Name:      "svc1-slice1",
			Namespace: "testns",
			Index:     object.EndpointsKey("svc1", "testns"),
		},
		{
			Subsets: []object.EndpointSubset{
				{
					Addresses: []object.EndpointAddress{
						{IP: "172.0.0.2"},
					},
					Ports: []object.EndpointPort{
						{Port: 80, Protocol: "tcp", Name: "http"},
					},
				},
			},
			Name:      "hdls1-slice1",
			Namespace: "testns",
			Index:     object.EndpointsKey("hdls1", "testns"),
		},
		{
			Subsets: []object.EndpointSubset{
				{
					Addresses: []object.EndpointAddress{
						{IP: "10.9.8.7", NodeName: "test.node.foo.bar"},
					},
				},
			},
		},
	}
	return eps
}

func (APIConnServiceTest) EndpointsList() []*object.Endpoints {
	eps := []*object.Endpoints{
		{
			Subsets: []object.EndpointSubset{
				{
					Addresses: []object.EndpointAddress{
						{IP: "172.0.0.1", Hostname: "ep1a"},
					},
					Ports: []object.EndpointPort{
						{Port: 80, Protocol: "tcp", Name: "http"},
					},
				},
			},
			Name:      "svc1-slice1",
			Namespace: "testns",
			Index:     object.EndpointsKey("svc1", "testns"),
		},
		{
			Subsets: []object.EndpointSubset{
				{
					Addresses: []object.EndpointAddress{
						{IP: "172.0.0.2"},
					},
					Ports: []object.EndpointPort{
						{Port: 80, Protocol: "tcp", Name: "http"},
					},
				},
			},
			Name:      "hdls1-slice1",
			Namespace: "testns",
			Index:     object.EndpointsKey("hdls1", "testns"),
		},
		{
			Subsets: []object.EndpointSubset{
				{
					Addresses: []object.EndpointAddress{
						{IP: "172.0.0.2"},
					},
					Ports: []object.EndpointPort{
						{Port: 80, Protocol: "tcp", Name: "http"},
					},
				},
			},
			Name:      "hdls1-slice2",
			Namespace: "testns",
			Index:     object.EndpointsKey("hdls1", "testns"),
		},
		{
			Subsets: []object.EndpointSubset{
				{
					Addresses: []object.EndpointAddress{
						{IP: "10.9.8.7", NodeName: "test.node.foo.bar"},
					},
				},
			},
		},
	}
	return eps
}

func (APIConnServiceTest) SvcImportIndex(string) []*object.ServiceImport {
	svcs := []*object.ServiceImport{
		{
			Name:       "svc1",
			Namespace:  "testns",
			Type:       mcs.ClusterSetIP,
			ClusterIPs: []string{"10.0.0.1"},
			Ports: []mcs.ServicePort{
				{Name: "http", Protocol: "tcp", Port: 80},
			},
		},
		{
			Name:       "svc-dual-stack",
			Namespace:  "testns",
			Type:       mcs.ClusterSetIP,
			ClusterIPs: []string{"10.0.0.2", "10::2"},
			Ports: []mcs.ServicePort{
				{Name: "http", Protocol: "tcp", Port: 80},
			},
		},
		{
			Name:       "hdls1",
			Namespace:  "testns",
			Type:       mcs.Headless,
			ClusterIPs: []string{},
		},
	}
	return svcs
}

func (APIConnServiceTest) ServiceImportList() []*object.ServiceImport {
	svcs := []*object.ServiceImport{
		{
			Name:       "",
			Namespace:  "testns",
			Type:       mcs.ClusterSetIP,
			ClusterIPs: []string{"10.0.0.1"},
			Ports: []mcs.ServicePort{
				{Name: "http", Protocol: "tcp", Port: 80},
			},
		},
		{
			Name:       "svc-dual-stack",
			Namespace:  "testns",
			Type:       mcs.ClusterSetIP,
			ClusterIPs: []string{"10.0.0.2", "10::2"},
			Ports: []mcs.ServicePort{
				{Name: "http", Protocol: "tcp", Port: 80},
			},
		},
		{
			Name:      "hdls1",
			Namespace: "testns",
			Type:      mcs.Headless,
		},
	}
	return svcs
}

func (APIConnServiceTest) McEpIndex(string) []*object.MultiClusterEndpoints {
	eps := []*object.MultiClusterEndpoints{
		{
			Endpoints: object.Endpoints{
				Subsets: []object.EndpointSubset{
					{
						Addresses: []object.EndpointAddress{
							{IP: "172.0.0.1", Hostname: "ep1a"},
						},
						Ports: []object.EndpointPort{
							{Port: 80, Protocol: "tcp", Name: "http"},
						},
					},
				},
				Name:      "svc1-slice1",
				Namespace: "testns",
				Index:     object.EndpointsKey("svc1", "testns"),
			},
			ClusterId: "cluster1",
		},
		{
			Endpoints: object.Endpoints{
				Subsets: []object.EndpointSubset{
					{
						Addresses: []object.EndpointAddress{
							{IP: "172.0.0.2"},
						},
						Ports: []object.EndpointPort{
							{Port: 80, Protocol: "tcp", Name: "http"},
						},
					},
				},
				Name:      "hdls1-slice1",
				Namespace: "testns",
				Index:     object.EndpointsKey("hdls1", "testns"),
			},
			ClusterId: "cluster1",
		},
	}
	return eps
}

func (APIConnServiceTest) GetNodeByName(ctx context.Context, name string) (*api.Node, error) {
	return &api.Node{
		ObjectMeta: meta.ObjectMeta{
			Name: "test.node.foo.bar",
		},
	}, nil
}

func (APIConnServiceTest) GetNamespaceByName(name string) (*object.Namespace, error) {
	return &object.Namespace{
		Name: name,
	}, nil
}

func TestServices(t *testing.T) {
	k := New([]string{"interwebs.test.", "clusterset.test."})
	k.opts.multiclusterZones = []string{"clusterset.test."}
	k.APIConn = &APIConnServiceTest{}

	type svcAns struct {
		host string
		key  string
	}
	type svcTest struct {
		qname  string
		qtype  uint16
		answer []svcAns
	}
	tests := []svcTest{
		// Cluster IP Services
		{qname: "svc1.testns.svc.interwebs.test.", qtype: dns.TypeA, answer: []svcAns{{host: "10.0.0.1", key: "/" + coredns + "/test/interwebs/svc/testns/svc1"}}},
		{qname: "_http._tcp.svc1.testns.svc.interwebs.test.", qtype: dns.TypeSRV, answer: []svcAns{{host: "10.0.0.1", key: "/" + coredns + "/test/interwebs/svc/testns/svc1"}}},
		{qname: "ep1a.svc1.testns.svc.interwebs.test.", qtype: dns.TypeA, answer: []svcAns{{host: "172.0.0.1", key: "/" + coredns + "/test/interwebs/svc/testns/svc1/ep1a"}}},

		// Dual-Stack Cluster IP Service
		{
			qname: "_http._tcp.svc-dual-stack.testns.svc.interwebs.test.",
			qtype: dns.TypeSRV,
			answer: []svcAns{
				{host: "10.0.0.2", key: "/" + coredns + "/test/interwebs/svc/testns/svc-dual-stack"},
				{host: "10::2", key: "/" + coredns + "/test/interwebs/svc/testns/svc-dual-stack"},
			},
		},

		// External Services
		{qname: "external.testns.svc.interwebs.test.", qtype: dns.TypeCNAME, answer: []svcAns{{host: "coredns.io", key: "/" + coredns + "/test/interwebs/svc/testns/external"}}},

		// Headless Services
		{qname: "hdls1.testns.svc.interwebs.test.", qtype: dns.TypeA, answer: []svcAns{{host: "172.0.0.2", key: "/" + coredns + "/test/interwebs/svc/testns/hdls1/172-0-0-2"}}},

		// ClusterSet MultiCluster IP Services
		{qname: "svc1.testns.svc.clusterset.test.", qtype: dns.TypeA, answer: []svcAns{{host: "10.0.0.1", key: "/" + coredns + "/test/clusterset/svc/testns/svc1"}}},
		{qname: "_http._tcp.svc1.testns.svc.clusterset.test.", qtype: dns.TypeSRV, answer: []svcAns{{host: "10.0.0.1", key: "/" + coredns + "/test/clusterset/svc/testns/svc1"}}},
		{qname: "ep1a.cluster1.svc1.testns.svc.clusterset.test.", qtype: dns.TypeA, answer: []svcAns{{host: "172.0.0.1", key: "/" + coredns + "/test/clusterset/svc/testns/svc1/cluster1/ep1a"}}},

		// Dual-Stack ClusterSet MultiCluster IP Service
		{
			qname: "_http._tcp.svc-dual-stack.testns.svc.clusterset.test.",
			qtype: dns.TypeSRV,
			answer: []svcAns{
				{host: "10.0.0.2", key: "/" + coredns + "/test/clusterset/svc/testns/svc-dual-stack"},
				{host: "10::2", key: "/" + coredns + "/test/clusterset/svc/testns/svc-dual-stack"},
			},
		},

		// Headless MultiCluster Services
		{qname: "hdls1.testns.svc.clusterset.test.", qtype: dns.TypeA, answer: []svcAns{{host: "172.0.0.2", key: "/" + coredns + "/test/clusterset/svc/testns/hdls1/cluster1/172-0-0-2"}}},
	}

	for i, test := range tests {
		zone := "interwebs.test."
		if strings.Contains(test.qname, "clusterset.test") {
			zone = "clusterset.test."
		}
		state := request.Request{
			Req:  &dns.Msg{Question: []dns.Question{{Name: test.qname, Qtype: test.qtype}}},
			Zone: zone,
		}
		svcs, e := k.Services(context.TODO(), state, false, plugin.Options{})
		if e != nil {
			t.Errorf("Test %d: got error '%v'", i, e)
			continue
		}
		if len(svcs) != len(test.answer) {
			t.Errorf("Test %d, expected %v answer, got %v", i, len(test.answer), len(svcs))
			continue
		}

		for j := range svcs {
			if test.answer[j].host != svcs[j].Host {
				t.Errorf("Test %d, expected host '%v', got '%v'", i, test.answer[j].host, svcs[j].Host)
			}
			if test.answer[j].key != svcs[j].Key {
				t.Errorf("Test %d, expected key '%v', got '%v'", i, test.answer[j].key, svcs[j].Key)
			}
		}
	}
}

func TestServicesAuthority(t *testing.T) {
	k := New([]string{"interwebs.test."})
	k.APIConn = &APIConnServiceTest{}

	type svcAns struct {
		host string
		key  string
	}
	type svcTest struct {
		localIPs []net.IP
		qname    string
		qtype    uint16
		answer   []svcAns
	}
	tests := []svcTest{
		{localIPs: []net.IP{net.ParseIP("1.2.3.4")}, qname: "ns.dns.interwebs.test.", qtype: dns.TypeA, answer: []svcAns{{host: "1.2.3.4", key: "/" + coredns + "/test/interwebs/dns/ns"}}},
		{localIPs: []net.IP{net.ParseIP("1.2.3.4")}, qname: "ns.dns.interwebs.test.", qtype: dns.TypeAAAA},
		{localIPs: []net.IP{net.ParseIP("1:2::3:4")}, qname: "ns.dns.interwebs.test.", qtype: dns.TypeA},
		{localIPs: []net.IP{net.ParseIP("1:2::3:4")}, qname: "ns.dns.interwebs.test.", qtype: dns.TypeAAAA, answer: []svcAns{{host: "1:2::3:4", key: "/" + coredns + "/test/interwebs/dns/ns"}}},
		{
			localIPs: []net.IP{net.ParseIP("1.2.3.4"), net.ParseIP("1:2::3:4")},
			qname:    "ns.dns.interwebs.test.",
			qtype:    dns.TypeNS, answer: []svcAns{
				{host: "1.2.3.4", key: "/" + coredns + "/test/interwebs/dns/ns"},
				{host: "1:2::3:4", key: "/" + coredns + "/test/interwebs/dns/ns"},
			},
		},
	}

	for i, test := range tests {
		k.localIPs = test.localIPs

		state := request.Request{
			Req:  &dns.Msg{Question: []dns.Question{{Name: test.qname, Qtype: test.qtype}}},
			Zone: k.Zones[0],
		}
		svcs, e := k.Services(context.TODO(), state, false, plugin.Options{})
		if e != nil {
			t.Errorf("Test %d: got error '%v'", i, e)
			continue
		}
		if test.answer != nil && len(svcs) != len(test.answer) {
			t.Errorf("Test %d, expected 1 answer, got %v", i, len(svcs))
			continue
		}
		if test.answer == nil && len(svcs) != 0 {
			t.Errorf("Test %d, expected no answer, got %v", i, len(svcs))
			continue
		}

		if test.answer == nil && len(svcs) == 0 {
			continue
		}

		for i, answer := range test.answer {
			if answer.host != svcs[i].Host {
				t.Errorf("Test %d, expected host '%v', got '%v'", i, answer.host, svcs[i].Host)
			}
			if answer.key != svcs[i].Key {
				t.Errorf("Test %d, expected key '%v', got '%v'", i, answer.key, svcs[i].Key)
			}
		}
	}
}
