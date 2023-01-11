// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: lum-network/dfract/staking.proto

package types

import (
	fmt "fmt"
	types "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/gogoproto/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	_ "github.com/gogo/protobuf/types"
	github_com_gogo_protobuf_types "github.com/gogo/protobuf/types"
	io "io"
	math "math"
	math_bits "math/bits"
	time "time"
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

type StakingState int32

const (
	StakingState_StateBonded    StakingState = 0
	StakingState_StateUnbonding StakingState = 1
)

var StakingState_name = map[int32]string{
	0: "BONDED",
	1: "UNBONDING",
}

var StakingState_value = map[string]int32{
	"BONDED":    0,
	"UNBONDING": 1,
}

func (x StakingState) String() string {
	return proto.EnumName(StakingState_name, int32(x))
}

func (StakingState) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_38e8b605f1cc6c4c, []int{0}
}

type Stake struct {
	DelegatorAddress string        `protobuf:"bytes,1,opt,name=delegator_address,json=delegatorAddress,proto3" json:"delegator_address,omitempty"`
	StakingToken     types.Coin    `protobuf:"bytes,2,opt,name=staking_token,json=stakingToken,proto3" json:"staking_token"`
	Status           StakingState  `protobuf:"varint,3,opt,name=status,proto3,enum=lum.network.dfract.StakingState" json:"status,omitempty"`
	UnbondingTime    time.Duration `protobuf:"bytes,4,opt,name=unbonding_time,json=unbondingTime,proto3,stdduration" json:"unbonding_time,omitempty" yaml:"unbonding_time"`
	UnbondedAt       time.Time     `protobuf:"bytes,5,opt,name=unbonded_at,json=unbondedAt,proto3,stdtime" json:"unbonded_at"`
	CreatedAt        time.Time     `protobuf:"bytes,6,opt,name=created_at,json=createdAt,proto3,stdtime" json:"created_at"`
}

func (m *Stake) Reset()         { *m = Stake{} }
func (m *Stake) String() string { return proto.CompactTextString(m) }
func (*Stake) ProtoMessage()    {}
func (*Stake) Descriptor() ([]byte, []int) {
	return fileDescriptor_38e8b605f1cc6c4c, []int{0}
}
func (m *Stake) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Stake) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Stake.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Stake) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Stake.Merge(m, src)
}
func (m *Stake) XXX_Size() int {
	return m.Size()
}
func (m *Stake) XXX_DiscardUnknown() {
	xxx_messageInfo_Stake.DiscardUnknown(m)
}

var xxx_messageInfo_Stake proto.InternalMessageInfo

func (m *Stake) GetDelegatorAddress() string {
	if m != nil {
		return m.DelegatorAddress
	}
	return ""
}

func (m *Stake) GetStakingToken() types.Coin {
	if m != nil {
		return m.StakingToken
	}
	return types.Coin{}
}

func (m *Stake) GetStatus() StakingState {
	if m != nil {
		return m.Status
	}
	return StakingState_StateBonded
}

func (m *Stake) GetUnbondingTime() time.Duration {
	if m != nil {
		return m.UnbondingTime
	}
	return 0
}

func (m *Stake) GetUnbondedAt() time.Time {
	if m != nil {
		return m.UnbondedAt
	}
	return time.Time{}
}

func (m *Stake) GetCreatedAt() time.Time {
	if m != nil {
		return m.CreatedAt
	}
	return time.Time{}
}

func init() {
	proto.RegisterEnum("lum.network.dfract.StakingState", StakingState_name, StakingState_value)
	proto.RegisterType((*Stake)(nil), "lum.network.dfract.Stake")
}

func init() { proto.RegisterFile("lum-network/dfract/staking.proto", fileDescriptor_38e8b605f1cc6c4c) }

var fileDescriptor_38e8b605f1cc6c4c = []byte{
	// 490 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x52, 0x3f, 0x6f, 0xd3, 0x40,
	0x14, 0x8f, 0x69, 0x1b, 0x91, 0x4b, 0x1b, 0xc2, 0x09, 0x24, 0x13, 0x24, 0xc7, 0x64, 0x8a, 0x28,
	0xdc, 0xa9, 0x65, 0x41, 0x48, 0x0c, 0x71, 0x13, 0x21, 0x96, 0x20, 0x25, 0xe9, 0xc2, 0x12, 0x9d,
	0xed, 0xab, 0x7b, 0x4a, 0xce, 0x17, 0xf9, 0x9e, 0x81, 0x4c, 0xec, 0x9d, 0x3a, 0xb2, 0xf4, 0xc3,
	0xb0, 0x75, 0xec, 0xc8, 0x14, 0x50, 0xb2, 0x31, 0xf2, 0x09, 0x90, 0xed, 0x73, 0xd5, 0xd2, 0x89,
	0xcd, 0x7e, 0xbf, 0x7f, 0xef, 0x7e, 0x77, 0xc8, 0x9d, 0xa7, 0xf2, 0x65, 0xcc, 0xe1, 0xb3, 0x4a,
	0x66, 0x34, 0x3c, 0x49, 0x58, 0x00, 0x54, 0x03, 0x9b, 0x89, 0x38, 0x22, 0x8b, 0x44, 0x81, 0xc2,
	0x78, 0x9e, 0x4a, 0x62, 0x18, 0xa4, 0x60, 0xb4, 0x1e, 0x45, 0x2a, 0x52, 0x39, 0x4c, 0xb3, 0xaf,
	0x82, 0xd9, 0x72, 0x22, 0xa5, 0xa2, 0x39, 0xa7, 0xf9, 0x9f, 0x9f, 0x9e, 0xd0, 0x30, 0x4d, 0x18,
	0x08, 0x15, 0x1b, 0xbc, 0xfd, 0x2f, 0x0e, 0x42, 0x72, 0x0d, 0x4c, 0x2e, 0x4a, 0x83, 0x40, 0x69,
	0xa9, 0x34, 0xf5, 0x99, 0xe6, 0xf4, 0xd3, 0x81, 0xcf, 0x81, 0x1d, 0xd0, 0x40, 0x09, 0x63, 0xd0,
	0xf9, 0xbe, 0x85, 0x76, 0xc6, 0xc0, 0x66, 0x1c, 0xef, 0xa3, 0x87, 0x21, 0x9f, 0xf3, 0x88, 0x81,
	0x4a, 0xa6, 0x2c, 0x0c, 0x13, 0xae, 0xb5, 0x6d, 0xb9, 0x56, 0xb7, 0x36, 0x6a, 0x5e, 0x03, 0xbd,
	0x62, 0x8e, 0xfb, 0x68, 0xcf, 0x1c, 0x69, 0x0a, 0x6a, 0xc6, 0x63, 0xfb, 0x9e, 0x6b, 0x75, 0xeb,
	0x87, 0x4f, 0x48, 0x11, 0x47, 0xb2, 0x38, 0x62, 0xe2, 0xc8, 0x91, 0x12, 0xb1, 0xb7, 0x7d, 0xb9,
	0x6a, 0x57, 0x46, 0xbb, 0x46, 0x35, 0xc9, 0x44, 0xf8, 0x35, 0xaa, 0x6a, 0x60, 0x90, 0x6a, 0x7b,
	0xcb, 0xb5, 0xba, 0x8d, 0x43, 0x97, 0xdc, 0x2d, 0x86, 0x8c, 0x0b, 0xc5, 0x18, 0x18, 0xf0, 0x91,
	0xe1, 0xe3, 0xaf, 0xa8, 0x91, 0xc6, 0xbe, 0x8a, 0xc3, 0x7c, 0x03, 0x21, 0xb9, 0xbd, 0x6d, 0x16,
	0x28, 0x0a, 0x21, 0x65, 0x21, 0xa4, 0x6f, 0x0a, 0xf3, 0xde, 0x66, 0x0b, 0xfc, 0x5e, 0xb5, 0xed,
	0xdb, 0xc2, 0x17, 0x4a, 0x0a, 0xe0, 0x72, 0x01, 0xcb, 0x3f, 0xab, 0xf6, 0xe3, 0x25, 0x93, 0xf3,
	0x37, 0x9d, 0xdb, 0x8c, 0xce, 0xb7, 0x9f, 0x6d, 0x6b, 0xb4, 0x77, 0x3d, 0x9c, 0x08, 0xc9, 0xf1,
	0x00, 0xd5, 0x8b, 0x01, 0x0f, 0xa7, 0x0c, 0xec, 0x9d, 0x3c, 0xbd, 0x75, 0x27, 0x7d, 0x52, 0x5e,
	0x87, 0x77, 0x3f, 0x8b, 0x3f, 0xcf, 0x9c, 0x50, 0x29, 0xec, 0x01, 0x3e, 0x42, 0x28, 0x48, 0x38,
	0x83, 0xc2, 0xa5, 0xfa, 0x1f, 0x2e, 0x35, 0xa3, 0xeb, 0xc1, 0xf3, 0x21, 0xda, 0xbd, 0x59, 0x12,
	0x7e, 0x8a, 0xaa, 0xde, 0x87, 0x61, 0x7f, 0xd0, 0x6f, 0x56, 0x5a, 0x0f, 0xce, 0x2e, 0xdc, 0x7a,
	0x3e, 0xf6, 0xf2, 0x4c, 0xfc, 0x0c, 0xd5, 0x8e, 0x87, 0x19, 0xfc, 0x7e, 0xf8, 0xae, 0x69, 0xb5,
	0xf0, 0xd9, 0x85, 0xdb, 0xc8, 0xf1, 0xe3, 0xf2, 0x7c, 0xde, 0xe0, 0x72, 0xed, 0x58, 0x57, 0x6b,
	0xc7, 0xfa, 0xb5, 0x76, 0xac, 0xf3, 0x8d, 0x53, 0xb9, 0xda, 0x38, 0x95, 0x1f, 0x1b, 0xa7, 0xf2,
	0x71, 0x3f, 0x12, 0x70, 0x9a, 0xfa, 0x24, 0x50, 0x92, 0xde, 0x7c, 0xe5, 0xc1, 0x29, 0x13, 0x31,
	0xfd, 0x52, 0xbe, 0x76, 0x58, 0x2e, 0xb8, 0xf6, 0xab, 0xf9, 0xfe, 0xaf, 0xfe, 0x06, 0x00, 0x00,
	0xff, 0xff, 0xff, 0x73, 0xa2, 0x56, 0x10, 0x03, 0x00, 0x00,
}

func (m *Stake) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Stake) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Stake) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	n1, err1 := github_com_gogo_protobuf_types.StdTimeMarshalTo(m.CreatedAt, dAtA[i-github_com_gogo_protobuf_types.SizeOfStdTime(m.CreatedAt):])
	if err1 != nil {
		return 0, err1
	}
	i -= n1
	i = encodeVarintStaking(dAtA, i, uint64(n1))
	i--
	dAtA[i] = 0x32
	n2, err2 := github_com_gogo_protobuf_types.StdTimeMarshalTo(m.UnbondedAt, dAtA[i-github_com_gogo_protobuf_types.SizeOfStdTime(m.UnbondedAt):])
	if err2 != nil {
		return 0, err2
	}
	i -= n2
	i = encodeVarintStaking(dAtA, i, uint64(n2))
	i--
	dAtA[i] = 0x2a
	n3, err3 := github_com_gogo_protobuf_types.StdDurationMarshalTo(m.UnbondingTime, dAtA[i-github_com_gogo_protobuf_types.SizeOfStdDuration(m.UnbondingTime):])
	if err3 != nil {
		return 0, err3
	}
	i -= n3
	i = encodeVarintStaking(dAtA, i, uint64(n3))
	i--
	dAtA[i] = 0x22
	if m.Status != 0 {
		i = encodeVarintStaking(dAtA, i, uint64(m.Status))
		i--
		dAtA[i] = 0x18
	}
	{
		size, err := m.StakingToken.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintStaking(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x12
	if len(m.DelegatorAddress) > 0 {
		i -= len(m.DelegatorAddress)
		copy(dAtA[i:], m.DelegatorAddress)
		i = encodeVarintStaking(dAtA, i, uint64(len(m.DelegatorAddress)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintStaking(dAtA []byte, offset int, v uint64) int {
	offset -= sovStaking(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *Stake) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.DelegatorAddress)
	if l > 0 {
		n += 1 + l + sovStaking(uint64(l))
	}
	l = m.StakingToken.Size()
	n += 1 + l + sovStaking(uint64(l))
	if m.Status != 0 {
		n += 1 + sovStaking(uint64(m.Status))
	}
	l = github_com_gogo_protobuf_types.SizeOfStdDuration(m.UnbondingTime)
	n += 1 + l + sovStaking(uint64(l))
	l = github_com_gogo_protobuf_types.SizeOfStdTime(m.UnbondedAt)
	n += 1 + l + sovStaking(uint64(l))
	l = github_com_gogo_protobuf_types.SizeOfStdTime(m.CreatedAt)
	n += 1 + l + sovStaking(uint64(l))
	return n
}

func sovStaking(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozStaking(x uint64) (n int) {
	return sovStaking(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Stake) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowStaking
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
			return fmt.Errorf("proto: Stake: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Stake: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field DelegatorAddress", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowStaking
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
				return ErrInvalidLengthStaking
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthStaking
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.DelegatorAddress = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field StakingToken", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowStaking
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
				return ErrInvalidLengthStaking
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthStaking
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.StakingToken.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Status", wireType)
			}
			m.Status = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowStaking
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Status |= StakingState(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field UnbondingTime", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowStaking
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
				return ErrInvalidLengthStaking
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthStaking
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := github_com_gogo_protobuf_types.StdDurationUnmarshal(&m.UnbondingTime, dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field UnbondedAt", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowStaking
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
				return ErrInvalidLengthStaking
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthStaking
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := github_com_gogo_protobuf_types.StdTimeUnmarshal(&m.UnbondedAt, dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field CreatedAt", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowStaking
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
				return ErrInvalidLengthStaking
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthStaking
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := github_com_gogo_protobuf_types.StdTimeUnmarshal(&m.CreatedAt, dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipStaking(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthStaking
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthStaking
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
func skipStaking(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowStaking
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
					return 0, ErrIntOverflowStaking
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
					return 0, ErrIntOverflowStaking
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
				return 0, ErrInvalidLengthStaking
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupStaking
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthStaking
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthStaking        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowStaking          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupStaking = fmt.Errorf("proto: unexpected end of group")
)
