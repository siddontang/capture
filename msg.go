package main

import (
	"github.com/pingcap/kvproto/pkg/coprocessor"
	"github.com/pingcap/kvproto/pkg/kvrpcpb"
	"github.com/pingcap/kvproto/pkg/pdpb"
	"github.com/pingcap/kvproto/pkg/raft_serverpb"
	"github.com/pingcap/kvproto/pkg/tikvpb"

	"go.etcd.io/etcd/etcdserver/etcdserverpb"

	"github.com/golang/protobuf/proto"
)

var pathMsgs map[string][]proto.Message

func init() {
	pathMsgs = map[string][]proto.Message{
		"/pdpb.PD/GetMembers":        {new(pdpb.GetMembersRequest), new(pdpb.GetMembersResponse)},
		"/pdpb.PD/Tso":               {new(pdpb.TsoRequest), new(pdpb.TsoResponse)},
		"/pdpb.PD/Bootstrap":         {new(pdpb.BootstrapRequest), new(pdpb.BootstrapResponse)},
		"/pdpb.PD/IsBootstrapped":    {new(pdpb.IsBootstrappedRequest), new(pdpb.IsBootstrappedResponse)},
		"/pdpb.PD/AllocID":           {new(pdpb.AllocIDRequest), new(pdpb.AllocIDResponse)},
		"/pdpb.PD/GetStore":          {new(pdpb.GetStoreRequest), new(pdpb.GetStoreResponse)},
		"/pdpb.PD/PutStore":          {new(pdpb.PutStoreRequest), new(pdpb.PutStoreResponse)},
		"/pdpb.PD/GetAllStores":      {new(pdpb.GetAllStoresRequest), new(pdpb.GetAllStoresResponse)},
		"/pdpb.PD/StoreHeartbeat":    {new(pdpb.StoreHeartbeatRequest), new(pdpb.StoreHeartbeatResponse)},
		"/pdpb.PD/RegionHeartbeat":   {new(pdpb.RegionHeartbeatRequest), new(pdpb.RegionHeartbeatResponse)},
		"/pdpb.PD/GetRegion":         {new(pdpb.GetRegionRequest), new(pdpb.GetRegionResponse)},
		"/pdpb.PD/GetPrevRegion":     {new(pdpb.GetRegionRequest), new(pdpb.GetRegionResponse)},
		"/pdpb.PD/GetRegionByID":     {new(pdpb.GetRegionByIDRequest), new(pdpb.GetRegionResponse)},
		"/pdpb.PD/AskSplit":          {new(pdpb.AskSplitRequest), new(pdpb.AskSplitResponse)},
		"/pdpb.PD/ReportSplit":       {new(pdpb.ReportSplitRequest), new(pdpb.ReportSplitResponse)},
		"/pdpb.PD/AskBatchSplit":     {new(pdpb.AskBatchSplitRequest), new(pdpb.AskBatchSplitResponse)},
		"/pdpb.PD/ReportBatchSplit":  {new(pdpb.ReportBatchSplitRequest), new(pdpb.ReportBatchSplitResponse)},
		"/pdpb.PD/GetClusterConfig":  {new(pdpb.GetClusterConfigRequest), new(pdpb.GetClusterConfigResponse)},
		"/pdpb.PD/PutClusterConfig":  {new(pdpb.PutClusterConfigRequest), new(pdpb.PutClusterConfigResponse)},
		"/pdpb.PD/ScatterRegion":     {new(pdpb.ScatterRegionRequest), new(pdpb.ScatterRegionResponse)},
		"/pdpb.PD/GetGCSafePoint":    {new(pdpb.GetGCSafePointRequest), new(pdpb.GetGCSafePointResponse)},
		"/pdpb.PD/UpdateGCSafePoint": {new(pdpb.UpdateGCSafePointRequest), new(pdpb.UpdateGCSafePointResponse)},
		"/pdpb.PD/SyncRegions":       {new(pdpb.SyncRegionRequest), new(pdpb.SyncRegionResponse)},

		"/tikvpb.Tikv/KvGet":              {new(kvrpcpb.GetRequest), new(kvrpcpb.GetResponse)},
		"/tikvpb.Tikv/KvScan":             {new(kvrpcpb.ScanRequest), new(kvrpcpb.ScanResponse)},
		"/tikvpb.Tikv/KvPrewrite":         {new(kvrpcpb.PrewriteRequest), new(kvrpcpb.PrewriteResponse)},
		"/tikvpb.Tikv/KvCommit":           {new(kvrpcpb.CommitRequest), new(kvrpcpb.CommitResponse)},
		"/tikvpb.Tikv/KvImport":           {new(kvrpcpb.ImportRequest), new(kvrpcpb.ImportResponse)},
		"/tikvpb.Tikv/KvCleanup":          {new(kvrpcpb.CleanupRequest), new(kvrpcpb.CleanupResponse)},
		"/tikvpb.Tikv/KvBatchGet":         {new(kvrpcpb.BatchGetRequest), new(kvrpcpb.BatchGetResponse)},
		"/tikvpb.Tikv/KvBatchRollback":    {new(kvrpcpb.BatchRollbackRequest), new(kvrpcpb.BatchRollbackResponse)},
		"/tikvpb.Tikv/KvScanLock":         {new(kvrpcpb.ScanLockRequest), new(kvrpcpb.ScanLockResponse)},
		"/tikvpb.Tikv/KvResolveLock":      {new(kvrpcpb.ResolveLockRequest), new(kvrpcpb.ResolveLockResponse)},
		"/tikvpb.Tikv/KvGC":               {new(kvrpcpb.GCRequest), new(kvrpcpb.GCResponse)},
		"/tikvpb.Tikv/KvDeleteRange":      {new(kvrpcpb.DeleteRangeRequest), new(kvrpcpb.DeleteRangeResponse)},
		"/tikvpb.Tikv/RawGet":             {new(kvrpcpb.RawGetRequest), new(kvrpcpb.RawGetResponse)},
		"/tikvpb.Tikv/RawBatchGet":        {new(kvrpcpb.RawBatchGetRequest), new(kvrpcpb.RawBatchGetResponse)},
		"/tikvpb.Tikv/RawPut":             {new(kvrpcpb.RawPutRequest), new(kvrpcpb.RawPutResponse)},
		"/tikvpb.Tikv/RawBatchPut":        {new(kvrpcpb.RawBatchPutRequest), new(kvrpcpb.RawBatchPutResponse)},
		"/tikvpb.Tikv/RawDelete":          {new(kvrpcpb.RawDeleteRequest), new(kvrpcpb.RawDeleteResponse)},
		"/tikvpb.Tikv/RawBatchDelete":     {new(kvrpcpb.RawBatchDeleteRequest), new(kvrpcpb.RawBatchDeleteResponse)},
		"/tikvpb.Tikv/RawScan":            {new(kvrpcpb.RawScanRequest), new(kvrpcpb.RawScanResponse)},
		"/tikvpb.Tikv/RawDeleteRange":     {new(kvrpcpb.RawDeleteRangeRequest), new(kvrpcpb.RawDeleteRangeResponse)},
		"/tikvpb.Tikv/RawBatchScan":       {new(kvrpcpb.RawBatchScanRequest), new(kvrpcpb.RawBatchScanResponse)},
		"/tikvpb.Tikv/UnsafeDestroyRange": {new(kvrpcpb.UnsafeDestroyRangeRequest), new(kvrpcpb.UnsafeDestroyRangeResponse)},
		"/tikvpb.Tikv/Coprocessor":        {new(coprocessor.Request), new(coprocessor.Response)},
		"/tikvpb.Tikv/CoprocessorStream":  {new(coprocessor.Request), new(coprocessor.Response)},
		"/tikvpb.Tikv/Raft":               {new(raft_serverpb.RaftMessage), new(raft_serverpb.Done)},
		"/tikvpb.Tikv/Snapshot":           {new(raft_serverpb.SnapshotChunk), new(raft_serverpb.Done)},
		"/tikvpb.Tikv/SplitRegion ":       {new(kvrpcpb.SplitRegionRequest), new(kvrpcpb.SplitRegionResponse)},
		"/tikvpb.Tikv/MvccGetByKey":       {new(kvrpcpb.MvccGetByKeyRequest), new(kvrpcpb.MvccGetByKeyResponse)},
		"/tikvpb.Tikv/MvccGetByStartTs":   {new(kvrpcpb.MvccGetByStartTsRequest), new(kvrpcpb.MvccGetByStartTsResponse)},
		"/tikvpb.Tikv/BatchCommands":      {new(tikvpb.BatchCommandsRequest), new(tikvpb.BatchCommandsResponse)},

		"/etcdserverpb.KV/Range":       {new(etcdserverpb.RangeRequest), new(etcdserverpb.RangeResponse)},
		"/etcdserverpb.KV/Put":         {new(etcdserverpb.PutRequest), new(etcdserverpb.PutResponse)},
		"/etcdserverpb.KV/DeleteRange": {new(etcdserverpb.DeleteRangeRequest), new(etcdserverpb.DeleteRangeResponse)},
		"/etcdserverpb.KV/Txn":         {new(etcdserverpb.TxnRequest), new(etcdserverpb.TxnResponse)},
		"/etcdserverpb.KV/Compact":     {new(etcdserverpb.CompactionRequest), new(etcdserverpb.CompactionResponse)},

		"/etcdserverpb.Watch/Watch": {new(etcdserverpb.WatchRequest), new(etcdserverpb.WatchResponse)},

		"/etcdserverpb.Lease/LeaseGrant":      {new(etcdserverpb.LeaseGrantRequest), new(etcdserverpb.LeaseGrantResponse)},
		"/etcdserverpb.Lease/LeaseRevoke":     {new(etcdserverpb.LeaseRevokeRequest), new(etcdserverpb.LeaseRevokeResponse)},
		"/etcdserverpb.Lease/LeaseKeepAlive":  {new(etcdserverpb.LeaseKeepAliveRequest), new(etcdserverpb.LeaseKeepAliveResponse)},
		"/etcdserverpb.Lease/LeaseTimeToLive": {new(etcdserverpb.LeaseTimeToLiveRequest), new(etcdserverpb.LeaseTimeToLiveResponse)},
		"/etcdserverpb.Lease/LeaseLeases":     {new(etcdserverpb.LeaseLeasesRequest), new(etcdserverpb.LeaseLeasesResponse)},

		"/etcdserverpb.Cluster/MemberAdd":    {new(etcdserverpb.MemberAddRequest), new(etcdserverpb.MemberAddResponse)},
		"/etcdserverpb.Cluster/MemberRemove": {new(etcdserverpb.MemberRemoveRequest), new(etcdserverpb.MemberRemoveResponse)},
		"/etcdserverpb.Cluster/MemberUpdate": {new(etcdserverpb.MemberUpdateRequest), new(etcdserverpb.MemberUpdateResponse)},
		"/etcdserverpb.Cluster/MemberList":   {new(etcdserverpb.MemberListRequest), new(etcdserverpb.MemberListResponse)},

		"/etcdserverpb.Maintenance/Alarm":      {new(etcdserverpb.AlarmRequest), new(etcdserverpb.AlarmResponse)},
		"/etcdserverpb.Maintenance/Status":     {new(etcdserverpb.StatusRequest), new(etcdserverpb.StatusResponse)},
		"/etcdserverpb.Maintenance/Defragment": {new(etcdserverpb.DefragmentRequest), new(etcdserverpb.DefragmentResponse)},
		"/etcdserverpb.Maintenance/Hash":       {new(etcdserverpb.HashRequest), new(etcdserverpb.HashResponse)},
		"/etcdserverpb.Maintenance/HashKV":     {new(etcdserverpb.HashKVRequest), new(etcdserverpb.HashKVResponse)},
		"/etcdserverpb.Maintenance/Snapshot":   {new(etcdserverpb.SnapshotRequest), new(etcdserverpb.SnapshotResponse)},
		"/etcdserverpb.Maintenance/MoveLeader": {new(etcdserverpb.MoveLeaderRequest), new(etcdserverpb.MoveLeaderResponse)},

		"/etcdserverpb.Auth/AuthEnable":           {new(etcdserverpb.AuthEnableRequest), new(etcdserverpb.AuthEnableResponse)},
		"/etcdserverpb.Auth/AuthDisable":          {new(etcdserverpb.AuthDisableRequest), new(etcdserverpb.AuthDisableResponse)},
		"/etcdserverpb.Auth/Authenticate":         {new(etcdserverpb.AuthenticateRequest), new(etcdserverpb.AuthenticateResponse)},
		"/etcdserverpb.Auth/UserAdd":              {new(etcdserverpb.AuthUserAddRequest), new(etcdserverpb.AuthUserAddResponse)},
		"/etcdserverpb.Auth/UserGet":              {new(etcdserverpb.AuthUserGetRequest), new(etcdserverpb.AuthUserGetResponse)},
		"/etcdserverpb.Auth/UserList":             {new(etcdserverpb.AuthUserListRequest), new(etcdserverpb.AuthUserListResponse)},
		"/etcdserverpb.Auth/UserDelete":           {new(etcdserverpb.AuthUserDeleteRequest), new(etcdserverpb.AuthUserDeleteResponse)},
		"/etcdserverpb.Auth/UserChangePassword":   {new(etcdserverpb.AuthUserChangePasswordRequest), new(etcdserverpb.AuthUserChangePasswordResponse)},
		"/etcdserverpb.Auth/UserGrantRole":        {new(etcdserverpb.AuthUserGrantRoleRequest), new(etcdserverpb.AuthUserGrantRoleResponse)},
		"/etcdserverpb.Auth/UserRevokeRole":       {new(etcdserverpb.AuthUserRevokeRoleRequest), new(etcdserverpb.AuthUserRevokeRoleResponse)},
		"/etcdserverpb.Auth/RoleAdd":              {new(etcdserverpb.AuthRoleAddRequest), new(etcdserverpb.AuthRoleAddResponse)},
		"/etcdserverpb.Auth/RoleGet":              {new(etcdserverpb.AuthRoleGetRequest), new(etcdserverpb.AuthRoleGetResponse)},
		"/etcdserverpb.Auth/RoleList":             {new(etcdserverpb.AuthRoleListRequest), new(etcdserverpb.AuthRoleListResponse)},
		"/etcdserverpb.Auth/RoleDelete":           {new(etcdserverpb.AuthRoleDeleteRequest), new(etcdserverpb.AuthRoleDeleteResponse)},
		"/etcdserverpb.Auth/RoleGrantPermission":  {new(etcdserverpb.AuthRoleGrantPermissionRequest), new(etcdserverpb.AuthRoleGrantPermissionResponse)},
		"/etcdserverpb.Auth/RoleRevokePermission": {new(etcdserverpb.AuthRoleRevokePermissionRequest), new(etcdserverpb.AuthRoleRevokePermissionResponse)},
	}
}
