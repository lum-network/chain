// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: lum-network/icqueries/genesis.proto

package types

import (
	fmt "fmt"
	_ "github.com/cosmos/cosmos-proto"
	github_com_cosmos_cosmos_sdk_types "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

type Query struct {
	Id           string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	ConnectionId string `protobuf:"bytes,2,opt,name=connection_id,json=connectionId,proto3" json:"connection_id,omitempty"`
	ChainId      string `protobuf:"bytes,3,opt,name=chain_id,json=chainId,proto3" json:"chain_id,omitempty"`
	QueryType    string `protobuf:"bytes,4,opt,name=query_type,json=queryType,proto3" json:"query_type,omitempty"`
	Request      []byte `protobuf:"bytes,5,opt,name=request,proto3" json:"request,omitempty"`
	CallbackId   string `protobuf:"bytes,8,opt,name=callback_id,json=callbackId,proto3" json:"callback_id,omitempty"`
	Ttl          uint64 `protobuf:"varint,9,opt,name=ttl,proto3" json:"ttl,omitempty"`
	RequestSent  bool   `protobuf:"varint,11,opt,name=request_sent,json=requestSent,proto3" json:"request_sent,omitempty"`
	ExtraId      string `protobuf:"bytes,12,opt,name=extra_id,json=extraId,proto3" json:"extra_id,omitempty"`
}

func (m *Query) Reset()         { *m = Query{} }
func (m *Query) String() string { return proto.CompactTextString(m) }
func (*Query) ProtoMessage()    {}
func (*Query) Descriptor() ([]byte, []int) {
	return fileDescriptor_399613b68058de67, []int{0}
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

func (m *Query) GetCallbackId() string {
	if m != nil {
		return m.CallbackId
	}
	return ""
}

func (m *Query) GetTtl() uint64 {
	if m != nil {
		return m.Ttl
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
	return fileDescriptor_399613b68058de67, []int{1}
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
	return fileDescriptor_399613b68058de67, []int{2}
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
	proto.RegisterType((*Query)(nil), "stride.interchainquery.v1.Query")
	proto.RegisterType((*DataPoint)(nil), "stride.interchainquery.v1.DataPoint")
	proto.RegisterType((*GenesisState)(nil), "stride.interchainquery.v1.GenesisState")
}

func init() {
	proto.RegisterFile("lum-network/icqueries/genesis.proto", fileDescriptor_399613b68058de67)
}

var fileDescriptor_399613b68058de67 = []byte{
	// 500 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0x53, 0xb1, 0x8e, 0xd3, 0x40,
	0x10, 0x8d, 0x93, 0x0b, 0x49, 0x36, 0x3e, 0x74, 0x5a, 0x5d, 0xe1, 0x9c, 0x84, 0x63, 0x72, 0x12,
	0x8a, 0x10, 0x71, 0x04, 0xb4, 0x14, 0x28, 0x42, 0xe2, 0xdc, 0x1d, 0x3e, 0x2a, 0x1a, 0xcb, 0xb1,
	0x47, 0xce, 0x2a, 0xf6, 0x6e, 0xce, 0x3b, 0x3e, 0x2e, 0x25, 0x35, 0x0d, 0x1f, 0xc3, 0x47, 0x5c,
	0x79, 0xa2, 0x42, 0x14, 0x11, 0x4a, 0x3a, 0xbe, 0x02, 0xed, 0xae, 0x23, 0x22, 0x21, 0x3a, 0xaa,
	0xec, 0xcc, 0x9b, 0x79, 0xf3, 0x66, 0xf2, 0x4c, 0xce, 0xf3, 0xaa, 0x98, 0x70, 0xc0, 0x8f, 0xa2,
	0x5c, 0x4e, 0x59, 0x72, 0x5d, 0x41, 0xc9, 0x40, 0x4e, 0x33, 0xe0, 0x20, 0x99, 0xf4, 0x57, 0xa5,
	0x40, 0x41, 0x07, 0x12, 0x4b, 0x96, 0x82, 0xcf, 0x38, 0x42, 0x99, 0x2c, 0x62, 0xc6, 0x55, 0xdd,
	0xda, 0xbf, 0x79, 0x7e, 0x76, 0x9a, 0x89, 0x4c, 0xe8, 0xaa, 0xa9, 0x7a, 0x99, 0x86, 0xb3, 0x41,
	0x22, 0x64, 0x21, 0x64, 0x64, 0x00, 0x13, 0x18, 0x68, 0xf4, 0xa9, 0x49, 0xda, 0xef, 0x54, 0x37,
	0x7d, 0x48, 0x9a, 0x2c, 0x75, 0x2c, 0xcf, 0x1a, 0xf7, 0xc2, 0x26, 0x4b, 0xe9, 0x39, 0x39, 0x4e,
	0x04, 0xe7, 0x90, 0x20, 0x13, 0x3c, 0x62, 0xa9, 0xd3, 0xd4, 0x90, 0xfd, 0x27, 0x19, 0xa4, 0x74,
	0x40, 0xba, 0x5a, 0x80, 0xc2, 0x5b, 0x1a, 0xef, 0xe8, 0x38, 0x48, 0xe9, 0x23, 0x42, 0xb4, 0xac,
	0x08, 0xd7, 0x2b, 0x70, 0x8e, 0x34, 0xd8, 0xd3, 0x99, 0xf7, 0xeb, 0x15, 0x50, 0x87, 0x74, 0x4a,
	0xb8, 0xae, 0x40, 0xa2, 0xd3, 0xf6, 0xac, 0xb1, 0x1d, 0xee, 0x43, 0x3a, 0x24, 0xfd, 0x24, 0xce,
	0xf3, 0x79, 0x9c, 0x2c, 0x15, 0x6d, 0x57, 0x77, 0x92, 0x7d, 0x2a, 0x48, 0xe9, 0x09, 0x69, 0x21,
	0xe6, 0x4e, 0xcf, 0xb3, 0xc6, 0x47, 0xa1, 0x7a, 0xd2, 0xc7, 0xc4, 0xae, 0xbb, 0x23, 0x09, 0x1c,
	0x9d, 0xbe, 0x67, 0x8d, 0xbb, 0x61, 0xbf, 0xce, 0x5d, 0x01, 0x47, 0xa5, 0x14, 0x6e, 0xb1, 0x8c,
	0x15, 0xa5, 0x6d, 0x94, 0xea, 0x38, 0x48, 0x47, 0x9f, 0x9b, 0xa4, 0xf7, 0x26, 0xc6, 0xf8, 0x52,
	0x30, 0x8e, 0x7f, 0xdd, 0x21, 0x26, 0xc7, 0x25, 0x14, 0x02, 0x21, 0x5a, 0x00, 0xcb, 0x16, 0x68,
	0xee, 0x30, 0x7b, 0x75, 0xb7, 0x19, 0x36, 0x7e, 0x6c, 0x86, 0x4f, 0x32, 0x86, 0x8b, 0x6a, 0xee,
	0x27, 0xa2, 0xa8, 0x2f, 0x5b, 0xff, 0x4c, 0x64, 0xba, 0x9c, 0xaa, 0xdd, 0xa5, 0x1f, 0x70, 0xfc,
	0xf6, 0x75, 0x42, 0xea, 0xc3, 0x07, 0x1c, 0x43, 0xdb, 0x50, 0x5e, 0x68, 0x46, 0x1a, 0x11, 0x3b,
	0x17, 0x49, 0x9c, 0xef, 0x27, 0xb4, 0xfe, 0xc3, 0x84, 0xbe, 0x66, 0xac, 0x07, 0x3c, 0x25, 0xed,
	0x9b, 0x38, 0xaf, 0xcc, 0xdf, 0x60, 0xcf, 0x4e, 0x7f, 0x6d, 0x86, 0x27, 0x25, 0xc8, 0x2a, 0xc7,
	0x67, 0xa2, 0x60, 0x08, 0xc5, 0x0a, 0xd7, 0xa1, 0x29, 0x19, 0x5d, 0x12, 0xfb, 0xad, 0xb1, 0xdb,
	0x15, 0xc6, 0x08, 0xf4, 0x35, 0xe9, 0xd4, 0x36, 0x74, 0x2c, 0xaf, 0x35, 0xee, 0xbf, 0xf0, 0xfc,
	0x7f, 0xfa, 0xcf, 0xd7, 0x56, 0x9a, 0x1d, 0x29, 0xe5, 0xe1, 0xbe, 0x6d, 0x76, 0x71, 0xb7, 0x75,
	0xad, 0xfb, 0xad, 0x6b, 0xfd, 0xdc, 0xba, 0xd6, 0x97, 0x9d, 0xdb, 0xb8, 0xdf, 0xb9, 0x8d, 0xef,
	0x3b, 0xb7, 0xf1, 0xc1, 0x3f, 0x58, 0xed, 0xd0, 0xf9, 0x9a, 0x74, 0x7a, 0x7b, 0xf0, 0x05, 0xe8,
	0x35, 0xe7, 0x0f, 0xb4, 0x69, 0x5f, 0xfe, 0x0e, 0x00, 0x00, 0xff, 0xff, 0xe7, 0x68, 0xa8, 0x29,
	0x27, 0x03, 0x00, 0x00,
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
	if m.Ttl != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.Ttl))
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
	l = len(m.CallbackId)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	if m.Ttl != 0 {
		n += 1 + sovGenesis(uint64(m.Ttl))
	}
	if m.RequestSent {
		n += 2
	}
	l = len(m.ExtraId)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
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
				return fmt.Errorf("proto: wrong wireType = %d for field Ttl", wireType)
			}
			m.Ttl = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Ttl |= uint64(b&0x7F) << shift
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
		default:
			iNdEx = preIndex
			skippy, err := skipGenesis(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthGenesis
			}
			if (iNdEx + skippy) < 0 {
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
			if skippy < 0 {
				return ErrInvalidLengthGenesis
			}
			if (iNdEx + skippy) < 0 {
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
			if skippy < 0 {
				return ErrInvalidLengthGenesis
			}
			if (iNdEx + skippy) < 0 {
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
