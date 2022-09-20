// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: dfract/tx.proto

package types

import (
	fmt "fmt"
	types "github.com/cosmos/cosmos-sdk/types"
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

type MsgDeposit struct {
	DepositorAddress string     `protobuf:"bytes,1,opt,name=depositor_address,json=depositorAddress,proto3" json:"depositor_address,omitempty"`
	Amount           types.Coin `protobuf:"bytes,2,opt,name=amount,proto3" json:"amount"`
}

func (m *MsgDeposit) Reset()         { *m = MsgDeposit{} }
func (m *MsgDeposit) String() string { return proto.CompactTextString(m) }
func (*MsgDeposit) ProtoMessage()    {}
func (*MsgDeposit) Descriptor() ([]byte, []int) {
	return fileDescriptor_2488f78627331760, []int{0}
}
func (m *MsgDeposit) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgDeposit) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgDeposit.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgDeposit) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgDeposit.Merge(m, src)
}
func (m *MsgDeposit) XXX_Size() int {
	return m.Size()
}
func (m *MsgDeposit) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgDeposit.DiscardUnknown(m)
}

var xxx_messageInfo_MsgDeposit proto.InternalMessageInfo

func (m *MsgDeposit) GetDepositorAddress() string {
	if m != nil {
		return m.DepositorAddress
	}
	return ""
}

func (m *MsgDeposit) GetAmount() types.Coin {
	if m != nil {
		return m.Amount
	}
	return types.Coin{}
}

func init() {
	proto.RegisterType((*MsgDeposit)(nil), "lum.network.dfract.MsgDeposit")
}

func init() { proto.RegisterFile("dfract/tx.proto", fileDescriptor_2488f78627331760) }

var fileDescriptor_2488f78627331760 = []byte{
	// 248 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x3c, 0x90, 0xbd, 0x4e, 0xc3, 0x30,
	0x14, 0x85, 0x63, 0x84, 0x2a, 0x61, 0x06, 0x20, 0x62, 0x28, 0x1d, 0x4c, 0xc5, 0x54, 0xa9, 0xc2,
	0x57, 0x85, 0x81, 0x99, 0x02, 0x23, 0x4b, 0x47, 0x16, 0xe4, 0x38, 0x26, 0xb5, 0xa8, 0x7d, 0x23,
	0xff, 0x40, 0x79, 0x0b, 0x1e, 0xab, 0x63, 0x47, 0x26, 0x84, 0x92, 0x17, 0x41, 0x8d, 0x03, 0xdb,
	0xd1, 0xfd, 0x8e, 0xf4, 0xe9, 0x5c, 0x7a, 0x54, 0xbe, 0x38, 0x21, 0x03, 0x84, 0x35, 0xaf, 0x1d,
	0x06, 0xcc, 0xf3, 0x55, 0x34, 0xdc, 0xaa, 0xf0, 0x8e, 0xee, 0x95, 0x27, 0x38, 0x3a, 0xad, 0xb0,
	0xc2, 0x0e, 0xc3, 0x2e, 0xa5, 0xe6, 0x88, 0x49, 0xf4, 0x06, 0x3d, 0x14, 0xc2, 0x2b, 0x78, 0x9b,
	0x15, 0x2a, 0x88, 0x19, 0x48, 0xd4, 0x36, 0xf1, 0x0b, 0x47, 0xe9, 0xa3, 0xaf, 0xee, 0x55, 0x8d,
	0x5e, 0x87, 0x7c, 0x4a, 0x4f, 0xca, 0x14, 0xd1, 0x3d, 0x8b, 0xb2, 0x74, 0xca, 0xfb, 0x21, 0x19,
	0x93, 0xc9, 0xc1, 0xe2, 0xf8, 0x1f, 0xdc, 0xa6, 0x7b, 0x7e, 0x43, 0x07, 0xc2, 0x60, 0xb4, 0x61,
	0xb8, 0x37, 0x26, 0x93, 0xc3, 0xab, 0x33, 0x9e, 0x5c, 0x7c, 0xe7, 0xe2, 0xbd, 0x8b, 0xdf, 0xa1,
	0xb6, 0xf3, 0xfd, 0xcd, 0xf7, 0x79, 0xb6, 0xe8, 0xeb, 0xf3, 0x87, 0x4d, 0xc3, 0xc8, 0xb6, 0x61,
	0xe4, 0xa7, 0x61, 0xe4, 0xb3, 0x65, 0xd9, 0xb6, 0x65, 0xd9, 0x57, 0xcb, 0xb2, 0xa7, 0x69, 0xa5,
	0xc3, 0x32, 0x16, 0x5c, 0xa2, 0x81, 0x55, 0x34, 0x97, 0xfd, 0x44, 0x90, 0x4b, 0xa1, 0x2d, 0xac,
	0xe1, 0xef, 0x0f, 0x1f, 0xb5, 0xf2, 0xc5, 0xa0, 0x5b, 0x70, 0xfd, 0x1b, 0x00, 0x00, 0xff, 0xff,
	0xd3, 0x5c, 0xdc, 0x1e, 0x1e, 0x01, 0x00, 0x00,
}

func (m *MsgDeposit) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgDeposit) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgDeposit) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size, err := m.Amount.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintTx(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x12
	if len(m.DepositorAddress) > 0 {
		i -= len(m.DepositorAddress)
		copy(dAtA[i:], m.DepositorAddress)
		i = encodeVarintTx(dAtA, i, uint64(len(m.DepositorAddress)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintTx(dAtA []byte, offset int, v uint64) int {
	offset -= sovTx(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *MsgDeposit) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.DepositorAddress)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = m.Amount.Size()
	n += 1 + l + sovTx(uint64(l))
	return n
}

func sovTx(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozTx(x uint64) (n int) {
	return sovTx(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *MsgDeposit) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
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
			return fmt.Errorf("proto: MsgDeposit: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgDeposit: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field DepositorAddress", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.DepositorAddress = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Amount", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Amount.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthTx
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthTx
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
func skipTx(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowTx
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
					return 0, ErrIntOverflowTx
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
					return 0, ErrIntOverflowTx
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
				return 0, ErrInvalidLengthTx
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupTx
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthTx
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthTx        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowTx          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupTx = fmt.Errorf("proto: unexpected end of group")
)
