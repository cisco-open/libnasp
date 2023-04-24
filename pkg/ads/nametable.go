// Copyright (c) 2022 Cisco and/or its affiliates. All rights reserved.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       https://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package ads

import (
	"fmt"
	"net"

	"emperror.dev/errors"
	"google.golang.org/protobuf/proto"
	istio_xds_v3 "istio.io/istio/pilot/pkg/xds/v3"
	istio_networking_nds_v1 "istio.io/istio/pkg/dns/proto"
)

// resolveHostName resolves the provides host name to IP addresses using
// Istio's NDS
func (c *client) ResolveHost(hostName string) ([]net.IP, error) {
	var hostIPs []net.IP
	hostIP := net.ParseIP(hostName)

	if hostIP != nil {
		// the passed in hostName by caller in fact it's an IP thus no need to resolve it
		hostIPs = append(hostIPs, hostIP)
		return hostIPs, nil
	}

	// resolves host name to IP using Istio name tables
	nameTables := c.resources[istio_xds_v3.NameTableType].GetResources()

	// the passed in host name may not be FQDN, in that case we need to try with existing search domains
	tryHostNames := []string{hostName}
	for _, domain := range c.config.SearchDomains {
		tryHostNames = append(tryHostNames, fmt.Sprintf("%s.%s", hostName, domain))
	}

	var proto proto.Message
	var ok bool
	for _, h := range tryHostNames {
		if proto, ok = nameTables[h]; ok {
			break
		}
	}

	if !ok {
		return nil, &HostNotFoundError{HostName: hostName}
	}

	nameTable, ok := proto.(*istio_networking_nds_v1.NameTable_NameInfo)
	if !ok {
		return nil, errors.Errorf("couldn't convert Name table info %q from protobuf message format to typed object", proto)
	}

	for _, ipStr := range nameTable.GetIps() {
		ip := net.ParseIP(ipStr)
		if ip != nil {
			hostIPs = append(hostIPs, ip)
		}
	}
	return hostIPs, nil
}
