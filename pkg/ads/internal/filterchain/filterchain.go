//  Copyright (c) 2023 Cisco and/or its affiliates. All rights reserved.
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//        https://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.
//

package filterchain

import (
	"fmt"
	"net"
	"strings"

	"emperror.dev/errors"
	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_config_listener_v3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
)

type matchCriteria struct {
	destinationPort      uint32
	destinationIP        net.IP
	serverName           string
	transportProtocol    string
	applicationProtocols []string
}

type MatchOption interface {
	applyTo(*matchCriteria)
	String() string
}

type destinationPortMatchOption struct {
	destinationPort uint32
}

func WithDestinationPort(port uint32) *destinationPortMatchOption {
	return &destinationPortMatchOption{destinationPort: port}
}

func (o *destinationPortMatchOption) applyTo(opts *matchCriteria) {
	opts.destinationPort = o.destinationPort
}

func (o *destinationPortMatchOption) String() string {
	return fmt.Sprintf("destination port=%d", o.destinationPort)
}

type destinationIPMatchOption struct {
	destinationIP net.IP
}

func WithDestinationIP(ip net.IP) *destinationIPMatchOption {
	return &destinationIPMatchOption{destinationIP: ip}
}

func (o *destinationIPMatchOption) applyTo(opts *matchCriteria) {
	opts.destinationIP = o.destinationIP
}

func (o *destinationIPMatchOption) String() string {
	return fmt.Sprintf("destination IP=%s", o.destinationIP)
}

type serverNameMatchOption struct {
	serverName string
}

func (o *serverNameMatchOption) applyTo(opts *matchCriteria) {
	opts.serverName = o.serverName
}

func (o *serverNameMatchOption) String() string {
	return fmt.Sprintf("server name=%s", o.serverName)
}

func WithServerName(serverName string) *serverNameMatchOption {
	return &serverNameMatchOption{serverName: serverName}
}

type transportProtocolMatchOption struct {
	transportProtocol string
}

func (o *transportProtocolMatchOption) applyTo(opts *matchCriteria) {
	opts.transportProtocol = o.transportProtocol
}

func (o *transportProtocolMatchOption) String() string {
	return fmt.Sprintf("transport protocol=%s", o.transportProtocol)
}

func WithTransportProtocol(transportProtocol string) *transportProtocolMatchOption {
	return &transportProtocolMatchOption{transportProtocol: transportProtocol}
}

type applicationProtocolMatchOption struct {
	applicationProtocols []string
}

func (o *applicationProtocolMatchOption) applyTo(opts *matchCriteria) {
	opts.applicationProtocols = o.applicationProtocols
}

func (o *applicationProtocolMatchOption) String() string {
	return fmt.Sprintf("application protocols=%s", o.applicationProtocols)
}

func WithApplicationProtocols(applicationProtocols []string) *applicationProtocolMatchOption {
	return &applicationProtocolMatchOption{applicationProtocols: applicationProtocols}
}

// Filter returns the filter chains from the provided listener that matches the provided
// matching opts according to the rules described at https://github.com/envoyproxy/go-control-plane/blob/v0.9.9/envoy/config/listener/v3/listener_components.pb.go#L211
func Filter(listener *envoy_config_listener_v3.Listener, matchingOpts ...MatchOption) ([]*envoy_config_listener_v3.FilterChain, error) {
	criteria := &matchCriteria{}
	filterChains := listener.GetFilterChains()

	for _, opt := range matchingOpts {
		opt.applyTo(criteria)
	}

	if criteria.destinationPort <= 0 {
		return filterChains, nil
	}

	var err error

	// 1. match by destination port
	filterChains = getFilterChainsMatchingDstPort(filterChains, criteria.destinationPort)

	// 2. match destination IP address
	filterChains, err = getFilterChainsMatchingDstIP(filterChains, criteria.destinationIP)
	if err != nil {
		return nil, errors.WrapIf(err, "could match filter chains by destination address")
	}

	// 3. Server name (e.g. SNI for TLS protocol)
	filterChains = getFilterChainsMatchingServerName(filterChains, criteria.serverName)

	// 4. Transport protocol.
	filterChains = getFilterChainsMatchingTransportProtocol(filterChains, criteria.transportProtocol)

	// 5. Application protocols (e.g. ALPN for TLS protocol).
	filterChains = getFilterChainsMatchingApplicationProtocol(filterChains, criteria.applicationProtocols)

	return filterChains, nil
}

func getFilterChainsMatchingDstPort(filterChains []*envoy_config_listener_v3.FilterChain, dstPort uint32) []*envoy_config_listener_v3.FilterChain {
	var fcsMatchedByDstPort []*envoy_config_listener_v3.FilterChain
	var fcsWithNoDstPort []*envoy_config_listener_v3.FilterChain

	// match by exact destination port first as that is the most specific match
	// if there are no matches by specific destination port then check the next most specific
	// match which is the filter chain matches with no destination port
	for _, fc := range filterChains {
		fcm := fc.GetFilterChainMatch()

		if fcm.GetDestinationPort() != nil {
			if fcm.GetDestinationPort().GetValue() == dstPort {
				fcsMatchedByDstPort = append(fcsMatchedByDstPort, fc)
			}
		} else {
			fcsWithNoDstPort = append(fcsWithNoDstPort, fc)
		}
	}

	if len(fcsMatchedByDstPort) == 0 {
		fcsMatchedByDstPort = fcsWithNoDstPort
	}

	return fcsMatchedByDstPort
}

func getFilterChainsMatchingDstIP(filterChains []*envoy_config_listener_v3.FilterChain, ip net.IP) ([]*envoy_config_listener_v3.FilterChain, error) {
	// if there are no filter chain matches by CIDR than the next most specific matches are those
	// which don't have a CIDR defined
	var fcsMatchedByDstIP []*envoy_config_listener_v3.FilterChain
	var fcsWithNoDstIP []*envoy_config_listener_v3.FilterChain

	for _, fc := range filterChains {
		cidrs := fc.GetFilterChainMatch().GetPrefixRanges()
		if cidrs != nil {
			// verify destination IP address matches any of the cidrs of the filter chain
			ok, err := matchPrefixRanges(cidrs, ip)
			if err != nil {
				return nil, err
			}
			if ok {
				fcsMatchedByDstIP = append(fcsMatchedByDstIP, fc)
			}
		} else {
			fcsWithNoDstIP = append(fcsWithNoDstIP, fc)
		}
	}

	if len(fcsMatchedByDstIP) == 0 {
		fcsMatchedByDstIP = fcsWithNoDstIP
	}

	return fcsMatchedByDstIP, nil
}

func matchPrefixRanges(prefixRanges []*envoy_config_core_v3.CidrRange, ip net.IP) (bool, error) {
	for _, cidr := range prefixRanges {
		if cidr.GetAddressPrefix() == "" {
			continue
		}

		cidrStr := fmt.Sprintf("%s/%d", cidr.GetAddressPrefix(), cidr.GetPrefixLen().GetValue())
		_, ipnet, err := net.ParseCIDR(cidrStr)
		if err != nil {
			return false, errors.WrapIff(err, "couldn't parse address prefix: %q", cidrStr)
		}

		if ipnet.Contains(ip) {
			return true, nil
		}
	}
	return false, nil
}

func getFilterChainsMatchingServerName(filterChains []*envoy_config_listener_v3.FilterChain, serverName string) []*envoy_config_listener_v3.FilterChain {
	if len(serverName) == 0 {
		return filterChains
	}

	var fcsMatched []*envoy_config_listener_v3.FilterChain
	var fcsWithNoServerNames []*envoy_config_listener_v3.FilterChain

	for _, fc := range filterChains {
		if len(fc.GetFilterChainMatch().GetServerNames()) == 0 {
			fcsWithNoServerNames = append(fcsWithNoServerNames, fc)
			continue
		}

		for matchFound := false; !matchFound && len(serverName) > 0; {
			matchServerNames := []string{
				serverName,
				fmt.Sprintf(".%s", serverName),
				fmt.Sprintf("*.%s", serverName),
			}

			for _, matchServerName := range matchServerNames {
				matchFound = false
				for _, fcmServerName := range fc.GetFilterChainMatch().GetServerNames() {
					if matchServerName == fcmServerName {
						matchFound = true
						break
					}
				}

				if matchFound {
					fcsMatched = append(fcsMatched, fc)
					break
				}
			}

			if !matchFound {
				// e.g. if there is no match for a.b.com, .a.b.com, *.a.b.com
				// than try b.com, .b.com, *.b.com
				idx := strings.Index(serverName, ".")
				if idx == -1 {
					serverName = ""
				} else {
					serverName = serverName[idx+1:]
				}
			}
		}
	}

	if len(fcsMatched) == 0 {
		fcsMatched = fcsWithNoServerNames
	}

	return fcsMatched
}

func getFilterChainsMatchingTransportProtocol(filterChains []*envoy_config_listener_v3.FilterChain, transportProtocol string) []*envoy_config_listener_v3.FilterChain {
	if len(transportProtocol) == 0 {
		return filterChains
	}

	var fcsMatched []*envoy_config_listener_v3.FilterChain
	var fcsWithNoTransportProtocol []*envoy_config_listener_v3.FilterChain

	for _, fc := range filterChains {
		if len(fc.GetFilterChainMatch().GetTransportProtocol()) == 0 {
			fcsWithNoTransportProtocol = append(fcsWithNoTransportProtocol, fc)
			continue
		}

		if transportProtocol == fc.GetFilterChainMatch().GetTransportProtocol() {
			fcsMatched = append(fcsMatched, fc)
		}
	}

	if len(fcsMatched) == 0 {
		fcsMatched = fcsWithNoTransportProtocol
	}

	return fcsMatched
}

func getFilterChainsMatchingApplicationProtocol(filterChains []*envoy_config_listener_v3.FilterChain, applicationProtocols []string) []*envoy_config_listener_v3.FilterChain {
	if len(applicationProtocols) == 0 {
		return filterChains
	}

	var fcsMatched []*envoy_config_listener_v3.FilterChain
	var fcsWithNoApplicationProtocol []*envoy_config_listener_v3.FilterChain

	for _, fc := range filterChains {
		if len(fc.GetFilterChainMatch().GetApplicationProtocols()) == 0 {
			fcsWithNoApplicationProtocol = append(fcsWithNoApplicationProtocol, fc)
			continue
		}

		for _, appProto := range applicationProtocols {
			found := false
			for _, fcmAppProto := range fc.GetFilterChainMatch().GetApplicationProtocols() {
				if fcmAppProto == appProto {
					found = true
					break
				}
			}

			if found {
				fcsMatched = append(fcsMatched, fc)
				break
			}
		}
	}

	if len(fcsMatched) == 0 {
		fcsMatched = fcsWithNoApplicationProtocol
	}

	return fcsMatched
}
