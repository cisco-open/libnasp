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

package util

import (
	"time"

	"github.com/go-logr/logr"

	"github.com/cisco-open/nasp/pkg/network"
)

func PrintConnectionState(connectionState network.ConnectionState, logger logr.Logger) {
	localAddr := connectionState.LocalAddr().String()
	remoteAddr := connectionState.RemoteAddr().String()
	var localCertID, remoteCertID string

	if cert := connectionState.GetLocalCertificate(); cert != nil {
		localCertID = cert.String()
	}

	if cert := connectionState.GetPeerCertificate(); cert != nil {
		remoteCertID = cert.String()
	}

	logger.Info("connection info", "localAddr", localAddr, "localCertID", localCertID, "remoteAddr", remoteAddr, "remoteCertID", remoteCertID, "ttfb", connectionState.GetTimeToFirstByte().Format(time.RFC3339Nano))
}
