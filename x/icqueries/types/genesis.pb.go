// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: stride/interchainquery/v1/genesis.proto

package types

import (
	fmt "fmt"
	io "io"
	math "math"
	math_bits "math/bits"
	time "time"

	_ "github.com/cosmos/cosmos-proto"
	github_com_cosmos_cosmos_sdk_types "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/gogoproto/gogoproto"
	proto "github.com/cosmos/gogoproto/proto"
	_ "github.com/cosmos/gogoproto/types"
	github_com_cosmos_gogoproto_types "github.com/cosmos/gogoproto/types"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf
var _ = time.Kitchen

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

type TimeoutPolicy int32

const (
	TimeoutPolicy_REJECT_QUERY_RESPONSE  TimeoutPolicy = 0
	TimeoutPolicy_RETRY_QUERY_REQUEST    TimeoutPolicy = 1
	TimeoutPolicy_EXECUTE_QUERY_CALLBACK TimeoutPolicy = 2
)

var TimeoutPolicy_name = map[int32]string{
	0: "REJECT_QUERY_RESPONSE",
	1: "RETRY_QUERY_REQUEST",
	2: "EXECUTE_QUERY_CALLBACK",
}

var TimeoutPolicy_value = map[string]int32{
	"REJECT_QUERY_RESPONSE":  0,
	"RETRY_QUERY_REQUEST":    1,
	"EXECUTE_QUERY_CALLBACK": 2,
}

func (x TimeoutPolicy) String() string {
	return proto.EnumName(TimeoutPolicy_name, int32(x))
}

func (TimeoutPolicy) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_74cd646eb05658fd, []int{0}
}

type Query struct {
	Id                 string        `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	ConnectionId       string        `protobuf:"bytes,2,opt,name=connection_id,json=connectionId,proto3" json:"connection_id,omitempty"`
	ChainId            string        `protobuf:"bytes,3,opt,name=chain_id,json=chainId,proto3" json:"chain_id,omitempty"`
	QueryType          string        `protobuf:"bytes,4,opt,name=query_type,json=queryType,proto3" json:"query_type,omitempty"`
	Request            []byte        `protobuf:"bytes,5,opt,name=request,proto3" json:"request,omitempty"`
	CallbackModuleName string        `protobuf:"bytes,7,opt,name=callback_module_name,json=callbackModuleName,proto3" json:"callback_module_name,omitempty"`
	CallbackId         string        `protobuf:"bytes,8,opt,name=callback_id,json=callbackId,proto3" json:"callback_id,omitempty"`
	TimeoutTimestamp   uint64        `protobuf:"varint,9,opt,name=timeout_timestamp,json=timeoutTimestamp,proto3" json:"timeout_timestamp,omitempty"`
	RequestSent        bool          `protobuf:"varint,11,opt,name=request_sent,json=requestSent,proto3" json:"request_sent,omitempty"`
	ExtraId            string        `protobuf:"bytes,12,opt,name=extra_id,json=extraId,proto3" json:"extra_id,omitempty"`
	TimeoutPolicy      TimeoutPolicy `protobuf:"varint,13,opt,name=timeout_policy,json=timeoutPolicy,proto3,enum=stride.interchainquery.v1.TimeoutPolicy" json:"timeout_policy,omitempty"`
	TimeoutDuration    time.Duration `protobuf:"bytes,14,opt,name=timeout_duration,json=timeoutDuration,proto3,stdduration" json:"timeout_duration"`
	SubmissionHeight   uint64        `protobuf:"varint,15,opt,name=submission_height,json=submissionHeight,proto3" json:"submission_height,omitempty"`
}

func (m *Query) Reset()         { *m = Query{} }
func (m *Query) String() string { return proto.CompactTextString(m) }
func (*Query) ProtoMessage()    {}
func (*Query) Descriptor() ([]byte, []int) {
	return fileDescriptor_74cd646eb05658fd, []int{0}
}
func (m *Query) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Query) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Query.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Query) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Query.Merge(m, src)
}
func (m *Query) XXX_Size() int {
	return m.Size()
}
func (m *Query) XXX_DiscardUnknown() {
	xxx_messageInfo_Query.DiscardUnknown(m)
}

var xxx_messageInfo_Query proto.InternalMessageInfo

func (m *Query) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *Query) GetConnectionId() string {
	if m != nil {
		return m.ConnectionId
	}
	return ""
}

func (m *Query) GetChainId() string {
	if m != nil {
		return m.ChainId
	}
	return ""
}

func (m *Query) GetQueryType() string {
	if m != nil {
		return m.QueryType
	}
	return ""
}

func (m *Query) GetRequest() []byte {
	if m != nil {
		return m.Request
	}
	return nil
}

func (m *Query) GetCallbackModuleName() string {
	if m != nil {
		return m.CallbackModuleName
	}
	return ""
}

func (m *Query) GetCallbackId() string {
	if m != nil {
		return m.CallbackId
	}
	return ""
}

func (m *Query) GetTimeoutTimestamp() uint64 {
	if m != nil {
		return m.TimeoutTimestamp
	}
	return 0
}

func (m *Query) GetRequestSent() bool {
	if m != nil {
		return m.RequestSent
	}
	return false
}

func (m *Query) GetExtraId() string {
	if m != nil {
		return m.ExtraId
	}
	return ""
}

func (m *Query) GetTimeoutPolicy() TimeoutPolicy {
	if m != nil {
		return m.TimeoutPolicy
	}
	return TimeoutPolicy_REJECT_QUERY_RESPONSE
}

func (m *Query) GetTimeoutDuration() time.Duration {
	if m != nil {
		return m.TimeoutDuration
	}
	return 0
}

func (m *Query) GetSubmissionHeight() uint64 {
	if m != nil {
		return m.SubmissionHeight
	}
	return 0
}

type DataPoint struct {
	Id           string                                 `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	RemoteHeight github_com_cosmos_cosmos_sdk_types.Int `protobuf:"bytes,2,opt,name=remote_height,json=remoteHeight,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Int" json:"remote_height"`
	LocalHeight  github_com_cosmos_cosmos_sdk_types.Int `protobuf:"bytes,3,opt,name=local_height,json=localHeight,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Int" json:"local_height"`
	Value        []byte                                 `protobuf:"bytes,4,opt,name=value,proto3" json:"result,omitempty"`
}

func (m *DataPoint) Reset()         { *m = DataPoint{} }
func (m *DataPoint) String() string { return proto.CompactTextString(m) }
func (*DataPoint) ProtoMessage()    {}
func (*DataPoint) Descriptor() ([]byte, []int) {
	return fileDescriptor_74cd646eb05658fd, []int{1}
}
func (m *DataPoint) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *DataPoint) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_DataPoint.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *DataPoint) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DataPoint.Merge(m, src)
}
func (m *DataPoint) XXX_Size() int {
	return m.Size()
}
func (m *DataPoint) XXX_DiscardUnknown() {
	xxx_messageInfo_DataPoint.DiscardUnknown(m)
}

var xxx_messageInfo_DataPoint proto.InternalMessageInfo

func (m *DataPoint) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *DataPoint) GetValue() []byte {
	if m != nil {
		return m.Value
	}
	return nil
}

// GenesisState defines the epochs module's genesis state.
type GenesisState struct {
	Queries []Query `protobuf:"bytes,1,rep,name=queries,proto3" json:"queries"`
}

func (m *GenesisState) Reset()         { *m = GenesisState{} }
func (m *GenesisState) String() string { return proto.CompactTextString(m) }
func (*GenesisState) ProtoMessage()    {}
func (*GenesisState) Descriptor() ([]byte, []int) {
	return fileDescriptor_74cd646eb05658fd, []int{2}
}
func (m *GenesisState) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *GenesisState) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_GenesisState.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *GenesisState) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GenesisState.Merge(m, src)
}
func (m *GenesisState) XXX_Size() int {
	return m.Size()
}
func (m *GenesisState) XXX_DiscardUnknown() {
	xxx_messageInfo_GenesisState.DiscardUnknown(m)
}

var xxx_messageInfo_GenesisState proto.InternalMessageInfo

func (m *GenesisState) GetQueries() []Query {
	if m != nil {
		return m.Queries
	}
	return nil
}

func init() {
	proto.RegisterEnum("stride.interchainquery.v1.TimeoutPolicy", TimeoutPolicy_name, TimeoutPolicy_value)
	proto.RegisterType((*Query)(nil), "stride.interchainquery.v1.Query")
	proto.RegisterType((*DataPoint)(nil), "stride.interchainquery.v1.DataPoint")
	proto.RegisterType((*GenesisState)(nil), "stride.interchainquery.v1.GenesisState")
}

func init() {
	proto.RegisterFile("stride/interchainquery/v1/genesis.proto", fileDescriptor_74cd646eb05658fd)
}

var fileDescriptor_74cd646eb05658fd = []byte{
	// 719 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0x53, 0xc1, 0x52, 0xd3, 0x5c,
	0x14, 0x6e, 0x4a, 0xf9, 0xdb, 0xde, 0xa6, 0xa5, 0xff, 0x15, 0x35, 0x65, 0xc6, 0x36, 0xe2, 0x8c,
	0x66, 0x50, 0x12, 0xc1, 0xad, 0x0b, 0x69, 0xc9, 0x48, 0x15, 0xa1, 0xa4, 0x65, 0x46, 0xdc, 0x64,
	0xd2, 0xe4, 0xda, 0xde, 0x21, 0xc9, 0x2d, 0xb9, 0x37, 0x48, 0x9f, 0xc1, 0x8d, 0x4b, 0x17, 0x3e,
	0x86, 0x0f, 0xc1, 0x92, 0x71, 0xe5, 0xb8, 0x40, 0x07, 0x76, 0x3e, 0x85, 0x93, 0x7b, 0x13, 0x40,
	0x1d, 0x5d, 0xb9, 0x4a, 0xce, 0xf9, 0xbe, 0x73, 0xce, 0x97, 0xf3, 0x9d, 0x80, 0x7b, 0x94, 0x45,
	0xd8, 0x43, 0x06, 0x0e, 0x19, 0x8a, 0xdc, 0xb1, 0x83, 0xc3, 0x83, 0x18, 0x45, 0x53, 0xe3, 0x70,
	0xc5, 0x18, 0xa1, 0x10, 0x51, 0x4c, 0xf5, 0x49, 0x44, 0x18, 0x81, 0x0d, 0x41, 0xd4, 0x7f, 0x21,
	0xea, 0x87, 0x2b, 0x0b, 0xf3, 0x23, 0x32, 0x22, 0x9c, 0x65, 0x24, 0x6f, 0xa2, 0x60, 0xa1, 0xe1,
	0x12, 0x1a, 0x10, 0x6a, 0x0b, 0x40, 0x04, 0x29, 0xd4, 0x1c, 0x11, 0x32, 0xf2, 0x91, 0xc1, 0xa3,
	0x61, 0xfc, 0xda, 0xf0, 0xe2, 0xc8, 0x61, 0x98, 0x84, 0x02, 0x5f, 0xfc, 0x50, 0x00, 0xb3, 0x3b,
	0x49, 0x77, 0x58, 0x03, 0x79, 0xec, 0x29, 0x92, 0x2a, 0x69, 0x65, 0x2b, 0x8f, 0x3d, 0x78, 0x07,
	0x54, 0x5d, 0x12, 0x86, 0xc8, 0x4d, 0xd8, 0x36, 0xf6, 0x94, 0x3c, 0x87, 0xe4, 0xcb, 0x64, 0xd7,
	0x83, 0x0d, 0x50, 0xe2, 0x02, 0x13, 0x7c, 0x86, 0xe3, 0x45, 0x1e, 0x77, 0x3d, 0x78, 0x0b, 0x00,
	0x2e, 0xdb, 0x66, 0xd3, 0x09, 0x52, 0x0a, 0x1c, 0x2c, 0xf3, 0xcc, 0x60, 0x3a, 0x41, 0x50, 0x01,
	0xc5, 0x08, 0x1d, 0xc4, 0x88, 0x32, 0x65, 0x56, 0x95, 0x34, 0xd9, 0xca, 0x42, 0xf8, 0x10, 0xcc,
	0xbb, 0x8e, 0xef, 0x0f, 0x1d, 0x77, 0xdf, 0x0e, 0x88, 0x17, 0xfb, 0xc8, 0x0e, 0x9d, 0x00, 0x29,
	0x45, 0xde, 0x02, 0x66, 0xd8, 0x0b, 0x0e, 0x6d, 0x39, 0x01, 0x82, 0x2d, 0x50, 0xb9, 0xa8, 0xc0,
	0x9e, 0x52, 0xe2, 0x44, 0x90, 0xa5, 0xba, 0x1e, 0xbc, 0x0f, 0xfe, 0x67, 0x38, 0x40, 0x24, 0x66,
	0x76, 0xf2, 0xa4, 0xcc, 0x09, 0x26, 0x4a, 0x59, 0x95, 0xb4, 0x82, 0x55, 0x4f, 0x81, 0x41, 0x96,
	0x87, 0xb7, 0x81, 0x9c, 0x4a, 0xb1, 0x29, 0x0a, 0x99, 0x52, 0x51, 0x25, 0xad, 0x64, 0x55, 0xd2,
	0x5c, 0x1f, 0x85, 0x2c, 0xf9, 0x6c, 0x74, 0xc4, 0x22, 0x27, 0x99, 0x26, 0x8b, 0xcf, 0xe6, 0x71,
	0xd7, 0x83, 0xdb, 0xa0, 0x96, 0x8d, 0x9a, 0x10, 0x1f, 0xbb, 0x53, 0xa5, 0xaa, 0x4a, 0x5a, 0x6d,
	0x55, 0xd3, 0xff, 0xe8, 0xaa, 0x3e, 0x10, 0x05, 0x3d, 0xce, 0xb7, 0xaa, 0xec, 0x6a, 0x08, 0xb7,
	0x40, 0x26, 0xd1, 0xce, 0xbc, 0x53, 0x6a, 0xaa, 0xa4, 0x55, 0x56, 0x1b, 0xba, 0x30, 0x57, 0xcf,
	0xcc, 0xd5, 0xd7, 0x53, 0x42, 0xbb, 0x74, 0x7c, 0xda, 0xca, 0xbd, 0xff, 0xda, 0x92, 0xac, 0xb9,
	0xb4, 0x38, 0x83, 0x92, 0x5d, 0xd0, 0x78, 0x18, 0x60, 0x4a, 0x13, 0x5f, 0xc7, 0x08, 0x8f, 0xc6,
	0x4c, 0x99, 0x13, 0xbb, 0xb8, 0x04, 0x36, 0x78, 0x7e, 0xf1, 0x6d, 0x1e, 0x94, 0xd7, 0x1d, 0xe6,
	0xf4, 0x08, 0x0e, 0xd9, 0x6f, 0x27, 0xe2, 0x80, 0x6a, 0x84, 0x02, 0xc2, 0x50, 0xd6, 0x86, 0x9f,
	0x48, 0xfb, 0x71, 0x32, 0xfc, 0xcb, 0x69, 0xeb, 0xee, 0x08, 0xb3, 0x71, 0x3c, 0xd4, 0x5d, 0x12,
	0xa4, 0x47, 0x99, 0x3e, 0x96, 0xa9, 0xb7, 0x6f, 0x24, 0x67, 0x41, 0xf5, 0x6e, 0xc8, 0x3e, 0x7d,
	0x5c, 0x06, 0xe9, 0xcd, 0x76, 0x43, 0x66, 0xc9, 0xa2, 0xa5, 0x10, 0x00, 0x6d, 0x20, 0xfb, 0xc4,
	0x75, 0xfc, 0x6c, 0xc2, 0xcc, 0x3f, 0x98, 0x50, 0xe1, 0x1d, 0xd3, 0x01, 0x4b, 0x60, 0xf6, 0xd0,
	0xf1, 0x63, 0x71, 0xa1, 0x72, 0x7b, 0xfe, 0xfb, 0x69, 0xab, 0x1e, 0x21, 0x1a, 0xfb, 0xec, 0x01,
	0x09, 0x30, 0x43, 0xc1, 0x84, 0x4d, 0x2d, 0x41, 0x59, 0xec, 0x01, 0xf9, 0xa9, 0xf8, 0x53, 0xfb,
	0xcc, 0x61, 0x08, 0x3e, 0x01, 0xc5, 0xc4, 0x43, 0x8c, 0xa8, 0x22, 0xa9, 0x33, 0x5a, 0x65, 0x55,
	0xfd, 0x8b, 0xc9, 0xfc, 0x2f, 0x6b, 0x17, 0x12, 0xe5, 0x56, 0x56, 0xb6, 0x64, 0x83, 0xea, 0x4f,
	0xe6, 0xc3, 0x06, 0xb8, 0x6e, 0x99, 0xcf, 0xcc, 0xce, 0xc0, 0xde, 0xd9, 0x35, 0xad, 0x3d, 0xdb,
	0x32, 0xfb, 0xbd, 0xed, 0xad, 0xbe, 0x59, 0xcf, 0xc1, 0x9b, 0xe0, 0x9a, 0x65, 0x0e, 0xac, 0xbd,
	0x0b, 0x64, 0x67, 0xd7, 0xec, 0x0f, 0xea, 0x12, 0x5c, 0x00, 0x37, 0xcc, 0x97, 0x66, 0x67, 0x77,
	0x60, 0xa6, 0x50, 0x67, 0x6d, 0x73, 0xb3, 0xbd, 0xd6, 0x79, 0x5e, 0xcf, 0xb7, 0x37, 0x8e, 0xcf,
	0x9a, 0xd2, 0xc9, 0x59, 0x53, 0xfa, 0x76, 0xd6, 0x94, 0xde, 0x9d, 0x37, 0x73, 0x27, 0xe7, 0xcd,
	0xdc, 0xe7, 0xf3, 0x66, 0xee, 0x95, 0x7e, 0x65, 0x77, 0x7e, 0x1c, 0x2c, 0x87, 0x88, 0xbd, 0x21,
	0xd1, 0xbe, 0xc1, 0x55, 0x1b, 0x47, 0x06, 0x76, 0x53, 0x91, 0x62, 0x8f, 0xc3, 0xff, 0xf8, 0x95,
	0x3d, 0xfa, 0x11, 0x00, 0x00, 0xff, 0xff, 0xb7, 0xa7, 0xb4, 0xee, 0xc7, 0x04, 0x00, 0x00,
}

func (m *Query) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Query) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Query) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.SubmissionHeight != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.SubmissionHeight))
		i--
		dAtA[i] = 0x78
	}
	n1, err1 := github_com_cosmos_gogoproto_types.StdDurationMarshalTo(m.TimeoutDuration, dAtA[i-github_com_cosmos_gogoproto_types.SizeOfStdDuration(m.TimeoutDuration):])
	if err1 != nil {
		return 0, err1
	}
	i -= n1
	i = encodeVarintGenesis(dAtA, i, uint64(n1))
	i--
	dAtA[i] = 0x72
	if m.TimeoutPolicy != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.TimeoutPolicy))
		i--
		dAtA[i] = 0x68
	}
	if len(m.ExtraId) > 0 {
		i -= len(m.ExtraId)
		copy(dAtA[i:], m.ExtraId)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.ExtraId)))
		i--
		dAtA[i] = 0x62
	}
	if m.RequestSent {
		i--
		if m.RequestSent {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x58
	}
	if m.TimeoutTimestamp != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.TimeoutTimestamp))
		i--
		dAtA[i] = 0x48
	}
	if len(m.CallbackId) > 0 {
		i -= len(m.CallbackId)
		copy(dAtA[i:], m.CallbackId)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.CallbackId)))
		i--
		dAtA[i] = 0x42
	}
	if len(m.CallbackModuleName) > 0 {
		i -= len(m.CallbackModuleName)
		copy(dAtA[i:], m.CallbackModuleName)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.CallbackModuleName)))
		i--
		dAtA[i] = 0x3a
	}
	if len(m.Request) > 0 {
		i -= len(m.Request)
		copy(dAtA[i:], m.Request)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.Request)))
		i--
		dAtA[i] = 0x2a
	}
	if len(m.QueryType) > 0 {
		i -= len(m.QueryType)
		copy(dAtA[i:], m.QueryType)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.QueryType)))
		i--
		dAtA[i] = 0x22
	}
	if len(m.ChainId) > 0 {
		i -= len(m.ChainId)
		copy(dAtA[i:], m.ChainId)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.ChainId)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.ConnectionId) > 0 {
		i -= len(m.ConnectionId)
		copy(dAtA[i:], m.ConnectionId)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.ConnectionId)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Id) > 0 {
		i -= len(m.Id)
		copy(dAtA[i:], m.Id)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.Id)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *DataPoint) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *DataPoint) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *DataPoint) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Value) > 0 {
		i -= len(m.Value)
		copy(dAtA[i:], m.Value)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.Value)))
		i--
		dAtA[i] = 0x22
	}
	{
		size := m.LocalHeight.Size()
		i -= size
		if _, err := m.LocalHeight.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintGenesis(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x1a
	{
		size := m.RemoteHeight.Size()
		i -= size
		if _, err := m.RemoteHeight.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintGenesis(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x12
	if len(m.Id) > 0 {
		i -= len(m.Id)
		copy(dAtA[i:], m.Id)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.Id)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *GenesisState) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *GenesisState) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *GenesisState) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Queries) > 0 {
		for iNdEx := len(m.Queries) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Queries[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func encodeVarintGenesis(dAtA []byte, offset int, v uint64) int {
	offset -= sovGenesis(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *Query) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Id)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	l = len(m.ConnectionId)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	l = len(m.ChainId)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	l = len(m.QueryType)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	l = len(m.Request)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	l = len(m.CallbackModuleName)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	l = len(m.CallbackId)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	if m.TimeoutTimestamp != 0 {
		n += 1 + sovGenesis(uint64(m.TimeoutTimestamp))
	}
	if m.RequestSent {
		n += 2
	}
	l = len(m.ExtraId)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	if m.TimeoutPolicy != 0 {
		n += 1 + sovGenesis(uint64(m.TimeoutPolicy))
	}
	l = github_com_cosmos_gogoproto_types.SizeOfStdDuration(m.TimeoutDuration)
	n += 1 + l + sovGenesis(uint64(l))
	if m.SubmissionHeight != 0 {
		n += 1 + sovGenesis(uint64(m.SubmissionHeight))
	}
	return n
}

func (m *DataPoint) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Id)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	l = m.RemoteHeight.Size()
	n += 1 + l + sovGenesis(uint64(l))
	l = m.LocalHeight.Size()
	n += 1 + l + sovGenesis(uint64(l))
	l = len(m.Value)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	return n
}

func (m *GenesisState) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.Queries) > 0 {
		for _, e := range m.Queries {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	return n
}

func sovGenesis(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozGenesis(x uint64) (n int) {
	return sovGenesis(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Query) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGenesis
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Query: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Query: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Id", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Id = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ConnectionId", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ConnectionId = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ChainId", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ChainId = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field QueryType", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.QueryType = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Request", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Request = append(m.Request[:0], dAtA[iNdEx:postIndex]...)
			if m.Request == nil {
				m.Request = []byte{}
			}
			iNdEx = postIndex
		case 7:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field CallbackModuleName", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.CallbackModuleName = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 8:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field CallbackId", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.CallbackId = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 9:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field TimeoutTimestamp", wireType)
			}
			m.TimeoutTimestamp = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.TimeoutTimestamp |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 11:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field RequestSent", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.RequestSent = bool(v != 0)
		case 12:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ExtraId", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ExtraId = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 13:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field TimeoutPolicy", wireType)
			}
			m.TimeoutPolicy = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.TimeoutPolicy |= TimeoutPolicy(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 14:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field TimeoutDuration", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := github_com_cosmos_gogoproto_types.StdDurationUnmarshal(&m.TimeoutDuration, dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 15:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field SubmissionHeight", wireType)
			}
			m.SubmissionHeight = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.SubmissionHeight |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipGenesis(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthGenesis
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *DataPoint) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGenesis
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: DataPoint: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: DataPoint: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Id", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Id = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field RemoteHeight", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.RemoteHeight.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field LocalHeight", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.LocalHeight.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Value", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Value = append(m.Value[:0], dAtA[iNdEx:postIndex]...)
			if m.Value == nil {
				m.Value = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipGenesis(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthGenesis
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *GenesisState) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGenesis
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: GenesisState: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: GenesisState: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Queries", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Queries = append(m.Queries, Query{})
			if err := m.Queries[len(m.Queries)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipGenesis(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthGenesis
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipGenesis(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowGenesis
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthGenesis
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupGenesis
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthGenesis
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthGenesis        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowGenesis          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupGenesis = fmt.Errorf("proto: unexpected end of group")
)
