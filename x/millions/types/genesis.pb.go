// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: lum/network/millions/genesis.proto

package types

import (
	fmt "fmt"
	io "io"
	math "math"
	math_bits "math/bits"

	_ "github.com/cosmos/gogoproto/gogoproto"
	proto "github.com/cosmos/gogoproto/proto"
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

type GenesisState struct {
	Params           Params           `protobuf:"bytes,1,opt,name=params,proto3" json:"params" yaml:"params"`
	NextPoolId       uint64           `protobuf:"varint,2,opt,name=next_pool_id,json=nextPoolId,proto3" json:"next_pool_id,omitempty"`
	NextDepositId    uint64           `protobuf:"varint,3,opt,name=next_deposit_id,json=nextDepositId,proto3" json:"next_deposit_id,omitempty"`
	NextPrizeId      uint64           `protobuf:"varint,4,opt,name=next_prize_id,json=nextPrizeId,proto3" json:"next_prize_id,omitempty"`
	NextWithdrawalId uint64           `protobuf:"varint,5,opt,name=next_withdrawal_id,json=nextWithdrawalId,proto3" json:"next_withdrawal_id,omitempty"`
	Pools            []Pool           `protobuf:"bytes,6,rep,name=pools,proto3" json:"pools"`
	Deposits         []Deposit        `protobuf:"bytes,7,rep,name=deposits,proto3" json:"deposits"`
	Draws            []Draw           `protobuf:"bytes,8,rep,name=draws,proto3" json:"draws"`
	Prizes           []Prize          `protobuf:"bytes,9,rep,name=prizes,proto3" json:"prizes"`
	Withdrawals      []Withdrawal     `protobuf:"bytes,10,rep,name=withdrawals,proto3" json:"withdrawals"`
	EpochTrackers    []EpochTracker   `protobuf:"bytes,11,rep,name=epoch_trackers,json=epochTrackers,proto3" json:"epoch_trackers"`
	EpochUnbondings  []EpochUnbonding `protobuf:"bytes,12,rep,name=epoch_unbondings,json=epochUnbondings,proto3" json:"epoch_unbondings"`
}

func (m *GenesisState) Reset()         { *m = GenesisState{} }
func (m *GenesisState) String() string { return proto.CompactTextString(m) }
func (*GenesisState) ProtoMessage()    {}
func (*GenesisState) Descriptor() ([]byte, []int) {
	return fileDescriptor_ff43331bdfdad888, []int{0}
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

func (m *GenesisState) GetParams() Params {
	if m != nil {
		return m.Params
	}
	return Params{}
}

func (m *GenesisState) GetNextPoolId() uint64 {
	if m != nil {
		return m.NextPoolId
	}
	return 0
}

func (m *GenesisState) GetNextDepositId() uint64 {
	if m != nil {
		return m.NextDepositId
	}
	return 0
}

func (m *GenesisState) GetNextPrizeId() uint64 {
	if m != nil {
		return m.NextPrizeId
	}
	return 0
}

func (m *GenesisState) GetNextWithdrawalId() uint64 {
	if m != nil {
		return m.NextWithdrawalId
	}
	return 0
}

func (m *GenesisState) GetPools() []Pool {
	if m != nil {
		return m.Pools
	}
	return nil
}

func (m *GenesisState) GetDeposits() []Deposit {
	if m != nil {
		return m.Deposits
	}
	return nil
}

func (m *GenesisState) GetDraws() []Draw {
	if m != nil {
		return m.Draws
	}
	return nil
}

func (m *GenesisState) GetPrizes() []Prize {
	if m != nil {
		return m.Prizes
	}
	return nil
}

func (m *GenesisState) GetWithdrawals() []Withdrawal {
	if m != nil {
		return m.Withdrawals
	}
	return nil
}

func (m *GenesisState) GetEpochTrackers() []EpochTracker {
	if m != nil {
		return m.EpochTrackers
	}
	return nil
}

func (m *GenesisState) GetEpochUnbondings() []EpochUnbonding {
	if m != nil {
		return m.EpochUnbondings
	}
	return nil
}

func init() {
	proto.RegisterType((*GenesisState)(nil), "lum.network.millions.GenesisState")
}

func init() {
	proto.RegisterFile("lum/network/millions/genesis.proto", fileDescriptor_ff43331bdfdad888)
}

var fileDescriptor_ff43331bdfdad888 = []byte{
	// 515 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x93, 0x4f, 0x6b, 0x13, 0x41,
	0x18, 0x87, 0xb3, 0xb6, 0x89, 0x75, 0x36, 0xb1, 0x65, 0xa8, 0x30, 0x44, 0xdd, 0xac, 0x8b, 0x4a,
	0x0e, 0x76, 0x17, 0x2a, 0x08, 0x7a, 0x11, 0x42, 0xa5, 0x06, 0x0f, 0x96, 0x68, 0x11, 0xbc, 0x84,
	0x49, 0x76, 0xd8, 0x0c, 0x9d, 0xdd, 0x59, 0x76, 0x66, 0x49, 0xeb, 0xa7, 0xf0, 0x63, 0xf5, 0xd8,
	0x8b, 0xe0, 0xa9, 0x48, 0xf2, 0x0d, 0xfc, 0x04, 0x32, 0x7f, 0xd2, 0x54, 0x98, 0xe6, 0xb6, 0xbc,
	0xf3, 0xfc, 0x9e, 0x79, 0xdf, 0x97, 0x1d, 0x10, 0xb1, 0x3a, 0x4f, 0x0a, 0x22, 0xe7, 0xbc, 0x3a,
	0x4b, 0x72, 0xca, 0x18, 0xe5, 0x85, 0x48, 0x32, 0x52, 0x10, 0x41, 0x45, 0x5c, 0x56, 0x5c, 0x72,
	0xb8, 0xcf, 0xea, 0x3c, 0xb6, 0x4c, 0xbc, 0x62, 0xba, 0xfb, 0x19, 0xcf, 0xb8, 0x06, 0x12, 0xf5,
	0x65, 0xd8, 0xae, 0xdb, 0x97, 0x92, 0x92, 0x0b, 0x2a, 0x2d, 0xd3, 0x73, 0x33, 0x15, 0x9e, 0x5b,
	0xe0, 0x99, 0x13, 0x28, 0x71, 0x85, 0x73, 0xb1, 0xd1, 0x51, 0x72, 0xce, 0x2c, 0x10, 0xba, 0x81,
	0x8a, 0xfe, 0x20, 0x96, 0x78, 0xe1, 0x24, 0xe6, 0x54, 0xce, 0x54, 0x2b, 0x78, 0xb3, 0x88, 0x94,
	0x7c, 0x3a, 0x33, 0x44, 0xf4, 0xab, 0x09, 0xda, 0xc7, 0x66, 0x63, 0x5f, 0x24, 0x96, 0x04, 0x7e,
	0x02, 0x2d, 0xd3, 0x2c, 0xf2, 0x42, 0xaf, 0xef, 0x1f, 0x3e, 0x89, 0x5d, 0x1b, 0x8c, 0x4f, 0x34,
	0x33, 0x78, 0x74, 0x79, 0xdd, 0x6b, 0xfc, 0xbd, 0xee, 0x75, 0x2e, 0x70, 0xce, 0xde, 0x45, 0x26,
	0x19, 0x8d, 0xac, 0x02, 0x86, 0xa0, 0x5d, 0x90, 0x73, 0x39, 0x56, 0xb3, 0x8d, 0x69, 0x8a, 0xee,
	0x85, 0x5e, 0x7f, 0x7b, 0x04, 0x54, 0xed, 0x84, 0x73, 0x36, 0x4c, 0xe1, 0x4b, 0xb0, 0xab, 0x09,
	0xbb, 0x65, 0x05, 0x6d, 0x69, 0xa8, 0xa3, 0xca, 0x47, 0xa6, 0x3a, 0x4c, 0x61, 0x04, 0x3a, 0xc6,
	0xa4, 0x96, 0xa0, 0xa8, 0x6d, 0x4d, 0xf9, 0x5a, 0xa5, 0x6a, 0xc3, 0x14, 0xbe, 0x02, 0x50, 0x33,
	0xeb, 0x35, 0x28, 0xb0, 0xa9, 0xc1, 0x3d, 0x75, 0xf2, 0xed, 0xe6, 0x60, 0x98, 0xc2, 0x37, 0xa0,
	0xa9, 0xda, 0x12, 0xa8, 0x15, 0x6e, 0xf5, 0xfd, 0xc3, 0xee, 0x1d, 0x73, 0x72, 0xce, 0x06, 0xdb,
	0x6a, 0xca, 0x91, 0xc1, 0xe1, 0x7b, 0xb0, 0x63, 0x9b, 0x15, 0xe8, 0xbe, 0x8e, 0x3e, 0x75, 0x47,
	0x6d, 0xf3, 0x36, 0x7d, 0x13, 0x52, 0x17, 0xab, 0x26, 0x04, 0xda, 0xd9, 0x74, 0xf1, 0x51, 0x85,
	0xe7, 0xab, 0x8b, 0x35, 0x0e, 0xdf, 0x82, 0x96, 0x9e, 0x5e, 0xa0, 0x07, 0x3a, 0xf8, 0xf8, 0x8e,
	0x8e, 0x15, 0x63, 0x93, 0x36, 0x00, 0x3f, 0x02, 0x7f, 0xbd, 0x14, 0x81, 0x80, 0xce, 0x87, 0xee,
	0xfc, 0x7a, 0x49, 0x56, 0x72, 0x3b, 0x0a, 0x3f, 0x83, 0x87, 0xfa, 0xf7, 0x19, 0xcb, 0x0a, 0x4f,
	0xcf, 0x48, 0x25, 0x90, 0xaf, 0x65, 0x91, 0x5b, 0xf6, 0x41, 0xb1, 0x5f, 0x0d, 0x6a, 0x75, 0x1d,
	0x72, 0xab, 0x26, 0xe0, 0x29, 0xd8, 0x33, 0xc2, 0xba, 0x98, 0xf0, 0x22, 0xa5, 0x45, 0x26, 0x50,
	0x5b, 0x2b, 0x9f, 0x6f, 0x50, 0x9e, 0xae, 0x60, 0x2b, 0xdd, 0x25, 0xff, 0x55, 0xc5, 0xe0, 0xf8,
	0x72, 0x11, 0x78, 0x57, 0x8b, 0xc0, 0xfb, 0xb3, 0x08, 0xbc, 0x9f, 0xcb, 0xa0, 0x71, 0xb5, 0x0c,
	0x1a, 0xbf, 0x97, 0x41, 0xe3, 0xfb, 0x41, 0x46, 0xe5, 0xac, 0x9e, 0xc4, 0x53, 0x9e, 0x27, 0xac,
	0xce, 0x0f, 0x56, 0xcf, 0x63, 0x3a, 0xc3, 0xb4, 0x48, 0xce, 0xd7, 0xcf, 0x44, 0x5e, 0x94, 0x44,
	0x4c, 0x5a, 0xfa, 0x9d, 0xbc, 0xfe, 0x17, 0x00, 0x00, 0xff, 0xff, 0x25, 0xb2, 0x31, 0xe9, 0x6d,
	0x04, 0x00, 0x00,
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
	if len(m.EpochUnbondings) > 0 {
		for iNdEx := len(m.EpochUnbondings) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.EpochUnbondings[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x62
		}
	}
	if len(m.EpochTrackers) > 0 {
		for iNdEx := len(m.EpochTrackers) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.EpochTrackers[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x5a
		}
	}
	if len(m.Withdrawals) > 0 {
		for iNdEx := len(m.Withdrawals) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Withdrawals[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x52
		}
	}
	if len(m.Prizes) > 0 {
		for iNdEx := len(m.Prizes) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Prizes[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x4a
		}
	}
	if len(m.Draws) > 0 {
		for iNdEx := len(m.Draws) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Draws[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x42
		}
	}
	if len(m.Deposits) > 0 {
		for iNdEx := len(m.Deposits) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Deposits[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x3a
		}
	}
	if len(m.Pools) > 0 {
		for iNdEx := len(m.Pools) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Pools[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x32
		}
	}
	if m.NextWithdrawalId != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.NextWithdrawalId))
		i--
		dAtA[i] = 0x28
	}
	if m.NextPrizeId != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.NextPrizeId))
		i--
		dAtA[i] = 0x20
	}
	if m.NextDepositId != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.NextDepositId))
		i--
		dAtA[i] = 0x18
	}
	if m.NextPoolId != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.NextPoolId))
		i--
		dAtA[i] = 0x10
	}
	{
		size, err := m.Params.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintGenesis(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0xa
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
func (m *GenesisState) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.Params.Size()
	n += 1 + l + sovGenesis(uint64(l))
	if m.NextPoolId != 0 {
		n += 1 + sovGenesis(uint64(m.NextPoolId))
	}
	if m.NextDepositId != 0 {
		n += 1 + sovGenesis(uint64(m.NextDepositId))
	}
	if m.NextPrizeId != 0 {
		n += 1 + sovGenesis(uint64(m.NextPrizeId))
	}
	if m.NextWithdrawalId != 0 {
		n += 1 + sovGenesis(uint64(m.NextWithdrawalId))
	}
	if len(m.Pools) > 0 {
		for _, e := range m.Pools {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.Deposits) > 0 {
		for _, e := range m.Deposits {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.Draws) > 0 {
		for _, e := range m.Draws {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.Prizes) > 0 {
		for _, e := range m.Prizes {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.Withdrawals) > 0 {
		for _, e := range m.Withdrawals {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.EpochTrackers) > 0 {
		for _, e := range m.EpochTrackers {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.EpochUnbondings) > 0 {
		for _, e := range m.EpochUnbondings {
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
				return fmt.Errorf("proto: wrong wireType = %d for field Params", wireType)
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
			if err := m.Params.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field NextPoolId", wireType)
			}
			m.NextPoolId = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.NextPoolId |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field NextDepositId", wireType)
			}
			m.NextDepositId = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.NextDepositId |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field NextPrizeId", wireType)
			}
			m.NextPrizeId = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.NextPrizeId |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 5:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field NextWithdrawalId", wireType)
			}
			m.NextWithdrawalId = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.NextWithdrawalId |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Pools", wireType)
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
			m.Pools = append(m.Pools, Pool{})
			if err := m.Pools[len(m.Pools)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 7:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Deposits", wireType)
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
			m.Deposits = append(m.Deposits, Deposit{})
			if err := m.Deposits[len(m.Deposits)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 8:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Draws", wireType)
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
			m.Draws = append(m.Draws, Draw{})
			if err := m.Draws[len(m.Draws)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 9:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Prizes", wireType)
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
			m.Prizes = append(m.Prizes, Prize{})
			if err := m.Prizes[len(m.Prizes)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 10:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Withdrawals", wireType)
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
			m.Withdrawals = append(m.Withdrawals, Withdrawal{})
			if err := m.Withdrawals[len(m.Withdrawals)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 11:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field EpochTrackers", wireType)
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
			m.EpochTrackers = append(m.EpochTrackers, EpochTracker{})
			if err := m.EpochTrackers[len(m.EpochTrackers)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 12:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field EpochUnbondings", wireType)
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
			m.EpochUnbondings = append(m.EpochUnbondings, EpochUnbonding{})
			if err := m.EpochUnbondings[len(m.EpochUnbondings)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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
