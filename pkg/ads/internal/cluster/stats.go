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

package cluster

import (
	"wwwin-github.cisco.com/eti/nasp/pkg/ads/internal/util"
)

// Stats stores various items related to an upstream cluster
type Stats struct {
	// WeightedClusterSelectionCount is the number of times the upstream cluster was selected in case of weighted routing
	// to clusters (traffic shifting scenario)
	WeightedClusterSelectionCount uint32
}

// ClustersStats stores Stats for a collection of clusters
type ClustersStats struct {
	*util.KeyValueCollection[string, Stats]
}

func (cs *ClustersStats) IncWeightedClusterSelectionCounter(clusterName string) {
	cs.Lock()
	defer cs.Unlock()

	if stats, ok := cs.Items()[clusterName]; ok {
		stats.WeightedClusterSelectionCount++
		cs.Items()[clusterName] = stats
	}
}

func (cs *ClustersStats) ResetWeightedClusterSelectionCounter(clusterName string) {
	cs.Lock()
	defer cs.Unlock()

	if stats, ok := cs.Items()[clusterName]; ok {
		stats.WeightedClusterSelectionCount = 0
		cs.Items()[clusterName] = stats
	}
}

func NewClustersStats() *ClustersStats {
	return &ClustersStats{
		KeyValueCollection: util.NewKeyValueCollection[string, Stats](),
	}
}
