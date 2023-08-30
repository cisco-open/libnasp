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

package request

// headerVersion derives the header version from the request api key and request api version
//
//nolint:funlen,gocognit,gocyclo,cyclop,maintidx
func headerVersion(apiKey, apiVersion int16) int16 {
	switch apiKey {
	case 0: // Produce
		if apiVersion >= 9 {
			return 2
		}
		return 1
	case 1: // Fetch
		if apiVersion >= 12 {
			return 2
		}
		return 1
	case 2: // ListOffsets
		if apiVersion >= 6 {
			return 2
		}
		return 1
	case 3: // Metadata
		if apiVersion >= 9 {
			return 2
		}
		return 1
	case 4: // LeaderAndIsr
		if apiVersion >= 4 {
			return 2
		}
		return 1
	case 5: // StopReplica
		if apiVersion >= 2 {
			return 2
		}
		return 1
	case 6: // UpdateMetadata
		if apiVersion >= 6 {
			return 2
		}
		return 1
	case 7: // ControlledShutdown
		// Version 0 of ControlledShutdownRequest has a non-standard request header
		// which does not include clientId.  Version 1 of ControlledShutdownRequest
		// and later use the standard request header.
		if apiVersion == 0 {
			return 0
		}
		if apiVersion >= 3 {
			return 2
		}
		return 1
	case 8: // OffsetCommit
		if apiVersion >= 8 {
			return 2
		}
		return 1
	case 9: // OffsetFetch
		if apiVersion >= 6 {
			return 2
		}
		return 1
	case 10: // FindCoordinator
		if apiVersion >= 3 {
			return 2
		}
		return 1
	case 11: // JoinGroup
		if apiVersion >= 6 {
			return 2
		}
		return 1
	case 12: // Heartbeat
		if apiVersion >= 4 {
			return 2
		}
		return 1
	case 13: // LeaveGroup
		if apiVersion >= 4 {
			return 2
		}
		return 1
	case 14: // SyncGroup
		if apiVersion >= 4 {
			return 2
		}
		return 1
	case 15: // DescribeGroups
		if apiVersion >= 5 {
			return 2
		}
		return 1
	case 16: // ListGroups
		if apiVersion >= 3 {
			return 2
		}
		return 1
	case 17: // SaslHandshake
		return 1
	case 18: // ApiVersions
		if apiVersion >= 3 {
			return 2
		}
		return 1
	case 19: // CreateTopics
		if apiVersion >= 5 {
			return 2
		}
		return 1
	case 20: // DeleteTopics
		if apiVersion >= 4 {
			return 2
		}
		return 1
	case 21: // DeleteRecords
		if apiVersion >= 2 {
			return 2
		}
		return 1
	case 22: // InitProducerId
		if apiVersion >= 2 {
			return 2
		}
		return 1
	case 23: // OffsetForLeaderEpoch
		if apiVersion >= 4 {
			return 2
		}
		return 1
	case 24: // AddPartitionsToTxn
		if apiVersion >= 3 {
			return 2
		}
		return 1
	case 25: // AddOffsetsToTxn
		if apiVersion >= 3 {
			return 2
		}
		return 1
	case 26: // EndTxn
		if apiVersion >= 3 {
			return 2
		}
		return 1
	case 27: // WriteTxnMarkers
		if apiVersion >= 1 {
			return 2
		}
		return 1
	case 28: // TxnOffsetCommit
		if apiVersion >= 3 {
			return 2
		}
		return 1
	case 29: // DescribeAcls
		if apiVersion >= 2 {
			return 2
		}
		return 1
	case 30: // CreateAcls
		if apiVersion >= 2 {
			return 2
		}
		return 1
	case 31: // DeleteAcls
		if apiVersion >= 2 {
			return 2
		}
		return 1
	case 32: // DescribeConfigs
		if apiVersion >= 4 {
			return 2
		}
		return 1
	case 33: // AlterConfigs
		if apiVersion >= 2 {
			return 2
		}
		return 1
	case 34: // AlterReplicaLogDirs
		if apiVersion >= 2 {
			return 2
		}
		return 1
	case 35: // DescribeLogDirs
		if apiVersion >= 2 {
			return 2
		}
		return 1
	case 36: // SaslAuthenticate
		if apiVersion >= 2 {
			return 2
		}
		return 1
	case 37: // CreatePartitions
		if apiVersion >= 2 {
			return 2
		}
		return 1
	case 38: // CreateDelegationToken
		if apiVersion >= 2 {
			return 2
		}
		return 1
	case 39: // RenewDelegationToken
		if apiVersion >= 2 {
			return 2
		}
		return 1
	case 40: // ExpireDelegationToken
		if apiVersion >= 2 {
			return 2
		}
		return 1
	case 41: // DescribeDelegationToken
		if apiVersion >= 2 {
			return 2
		}
		return 1
	case 42: // DeleteGroups
		if apiVersion >= 2 {
			return 2
		}
		return 1
	case 43: // ElectLeaders
		if apiVersion >= 2 {
			return 2
		}
		return 1
	case 44: // IncrementalAlterConfigs
		if apiVersion >= 1 {
			return 2
		}
		return 1
	case 45: // AlterPartitionReassignments
		return 2
	case 46: // ListPartitionReassignments
		return 2
	case 47: // OffsetDelete
		return 1
	case 48: // DescribeClientQuotas
		if apiVersion >= 1 {
			return 2
		}
		return 1
	case 49: // AlterClientQuotas
		if apiVersion >= 1 {
			return 2
		}
		return 1
	case 50: // DescribeUserScramCredentials
		return 2
	case 51: // AlterUserScramCredentials
		return 2
	case 52: // Vote
		return 2
	case 53: // BeginQuorumEpoch
		return 1
	case 54: // EndQuorumEpoch
		return 1
	case 55: // DescribeQuorum
		return 2
	case 56: // AlterIsr
		return 2
	case 57: // UpdateFeatures
		return 2
	case 58: // Envelope
		return 2
	case 59: // FetchSnapshot
		return 2
	case 60: // DescribeCluster
		return 2
	case 61: // DescribeProducers
		return 2
	case 62: // BrokerRegistration
		return 2
	case 63: // BrokerHeartbeat
		return 2
	case 64: // UnregisterBroker
		return 2
	case 65: // DescribeTransactions
		return 2
	case 66: // ListTransactions
		return 2
	case 67: // AllocateProducerIds
		return 2
	default:
		return -1
	}
}
