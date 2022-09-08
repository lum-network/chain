// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: dfract/proposal.proto

package types

import (
	fmt "fmt"
	_ "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	_ "github.com/regen-network/cosmos-proto"
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

type SpendAndAdjustProposal struct {
	Title            string `protobuf:"bytes,1,opt,name=title,proto3" json:"title,omitempty"`
	Description      string `protobuf:"bytes,2,opt,name=description,proto3" json:"description,omitempty"`
	SpendDestination string `protobuf:"bytes,3,opt,name=spend_destination,json=spendDestination,proto3" json:"spend_destination,omitempty"`
	MintRate         int64  `protobuf:"varint,4,opt,name=mint_rate,json=mintRate,proto3" json:"mint_rate,omitempty"`
}

func (m *SpendAndAdjustProposal) Reset()      { *m = SpendAndAdjustProposal{} }
func (*SpendAndAdjustProposal) ProtoMessage() {}
func (*SpendAndAdjustProposal) Descriptor() ([]byte, []int) {
	return fileDescriptor_e7847e6fdc37fddb, []int{0}
}
func (m *SpendAndAdjustProposal) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *SpendAndAdjustProposal) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_SpendAndAdjustProposal.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *SpendAndAdjustProposal) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SpendAndAdjustProposal.Merge(m, src)
}
func (m *SpendAndAdjustProposal) XXX_Size() int {
	return m.Size()
}
func (m *SpendAndAdjustProposal) XXX_DiscardUnknown() {
	xxx_messageInfo_SpendAndAdjustProposal.DiscardUnknown(m)
}

var xxx_messageInfo_SpendAndAdjustProposal proto.InternalMessageInfo

func init() {
	proto.RegisterType((*SpendAndAdjustProposal)(nil), "lum.network.dfract.SpendAndAdjustProposal")
}

func init() { proto.RegisterFile("dfract/proposal.proto", fileDescriptor_e7847e6fdc37fddb) }

var fileDescriptor_e7847e6fdc37fddb = []byte{
	// 321 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x4c, 0x90, 0xb1, 0x4e, 0xc3, 0x30,
	0x18, 0x84, 0x63, 0x0a, 0x88, 0x86, 0x05, 0xa2, 0x82, 0x4a, 0x91, 0xdc, 0x8a, 0x01, 0x55, 0xaa,
	0x1a, 0xab, 0x62, 0x63, 0x2b, 0xb0, 0xb0, 0xa1, 0xb2, 0xb1, 0x54, 0x4e, 0x62, 0x52, 0x43, 0xe2,
	0x3f, 0x8a, 0xff, 0x14, 0x78, 0x03, 0x46, 0x46, 0xc6, 0x8e, 0x3c, 0x00, 0x4f, 0xc0, 0xc4, 0x58,
	0x31, 0x31, 0xa2, 0xf4, 0x45, 0x50, 0x62, 0x4b, 0xb0, 0xf9, 0xee, 0x3b, 0x59, 0xff, 0x9d, 0xbb,
	0x17, 0xdd, 0xe6, 0x3c, 0x44, 0x96, 0xe5, 0x90, 0x81, 0xe6, 0x89, 0x9f, 0xe5, 0x80, 0xe0, 0x79,
	0x49, 0x91, 0xfa, 0x4a, 0xe0, 0x03, 0xe4, 0xf7, 0xbe, 0x89, 0x74, 0x5a, 0x31, 0xc4, 0x50, 0x63,
	0x56, 0xbd, 0x4c, 0xb2, 0x73, 0x10, 0x82, 0x4e, 0x41, 0x4f, 0x0d, 0x30, 0xc2, 0x22, 0x6a, 0x14,
	0x0b, 0xb8, 0x16, 0x6c, 0x3e, 0x0a, 0x04, 0xf2, 0x11, 0x0b, 0x41, 0x2a, 0xc3, 0x8f, 0x3e, 0x88,
	0xbb, 0x7f, 0x9d, 0x09, 0x15, 0x8d, 0x55, 0x34, 0x8e, 0xee, 0x0a, 0x8d, 0x57, 0xf6, 0x0a, 0xaf,
	0xe5, 0x6e, 0xa0, 0xc4, 0x44, 0xb4, 0x49, 0x8f, 0xf4, 0x9b, 0x13, 0x23, 0xbc, 0x9e, 0xbb, 0x1d,
	0x09, 0x1d, 0xe6, 0x32, 0x43, 0x09, 0xaa, 0xbd, 0x56, 0xb3, 0xff, 0x96, 0x37, 0x70, 0x77, 0x75,
	0xf5, 0xe3, 0x34, 0x12, 0x1a, 0xa5, 0xe2, 0x75, 0xae, 0x51, 0xe7, 0x76, 0x6a, 0x70, 0xf1, 0xe7,
	0x7b, 0x87, 0x6e, 0x33, 0x95, 0x0a, 0xa7, 0x39, 0x47, 0xd1, 0x5e, 0xef, 0x91, 0x7e, 0x63, 0xb2,
	0x55, 0x19, 0x13, 0x8e, 0xe2, 0xf4, 0xf8, 0x79, 0xd1, 0x75, 0x5e, 0x17, 0x5d, 0xe7, 0xeb, 0x7d,
	0xd8, 0xb1, 0xb5, 0x62, 0x98, 0xfb, 0xb6, 0x87, 0x7f, 0x0e, 0x0a, 0x85, 0xc2, 0xb3, 0xcb, 0xb7,
	0x92, 0x92, 0xcf, 0x92, 0x92, 0x65, 0x49, 0xc9, 0x4f, 0x49, 0xc9, 0xcb, 0x8a, 0x3a, 0xcb, 0x15,
	0x75, 0xbe, 0x57, 0xd4, 0xb9, 0x19, 0xc4, 0x12, 0x67, 0x45, 0xe0, 0x87, 0x90, 0xb2, 0xa4, 0x48,
	0x87, 0x76, 0x52, 0x16, 0xce, 0xb8, 0x54, 0xec, 0x91, 0xd9, 0xf5, 0xf1, 0x29, 0x13, 0x3a, 0xd8,
	0xac, 0x67, 0x39, 0xf9, 0x0d, 0x00, 0x00, 0xff, 0xff, 0xb5, 0xdd, 0x74, 0x75, 0x94, 0x01, 0x00,
	0x00,
}

func (this *SpendAndAdjustProposal) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*SpendAndAdjustProposal)
	if !ok {
		that2, ok := that.(SpendAndAdjustProposal)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if this.Title != that1.Title {
		return false
	}
	if this.Description != that1.Description {
		return false
	}
	if this.SpendDestination != that1.SpendDestination {
		return false
	}
	if this.MintRate != that1.MintRate {
		return false
	}
	return true
}
func (m *SpendAndAdjustProposal) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *SpendAndAdjustProposal) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *SpendAndAdjustProposal) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.MintRate != 0 {
		i = encodeVarintProposal(dAtA, i, uint64(m.MintRate))
		i--
		dAtA[i] = 0x20
	}
	if len(m.SpendDestination) > 0 {
		i -= len(m.SpendDestination)
		copy(dAtA[i:], m.SpendDestination)
		i = encodeVarintProposal(dAtA, i, uint64(len(m.SpendDestination)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.Description) > 0 {
		i -= len(m.Description)
		copy(dAtA[i:], m.Description)
		i = encodeVarintProposal(dAtA, i, uint64(len(m.Description)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Title) > 0 {
		i -= len(m.Title)
		copy(dAtA[i:], m.Title)
		i = encodeVarintProposal(dAtA, i, uint64(len(m.Title)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintProposal(dAtA []byte, offset int, v uint64) int {
	offset -= sovProposal(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *SpendAndAdjustProposal) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Title)
	if l > 0 {
		n += 1 + l + sovProposal(uint64(l))
	}
	l = len(m.Description)
	if l > 0 {
		n += 1 + l + sovProposal(uint64(l))
	}
	l = len(m.SpendDestination)
	if l > 0 {
		n += 1 + l + sovProposal(uint64(l))
	}
	if m.MintRate != 0 {
		n += 1 + sovProposal(uint64(m.MintRate))
	}
	return n
}

func sovProposal(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozProposal(x uint64) (n int) {
	return sovProposal(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *SpendAndAdjustProposal) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowProposal
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
			return fmt.Errorf("proto: SpendAndAdjustProposal: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: SpendAndAdjustProposal: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Title", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProposal
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
				return ErrInvalidLengthProposal
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthProposal
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Title = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Description", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProposal
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
				return ErrInvalidLengthProposal
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthProposal
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Description = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field SpendDestination", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProposal
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
				return ErrInvalidLengthProposal
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthProposal
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.SpendDestination = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field MintRate", wireType)
			}
			m.MintRate = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProposal
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.MintRate |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipProposal(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthProposal
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthProposal
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
func skipProposal(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowProposal
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
					return 0, ErrIntOverflowProposal
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
					return 0, ErrIntOverflowProposal
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
				return 0, ErrInvalidLengthProposal
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupProposal
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthProposal
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthProposal        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowProposal          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupProposal = fmt.Errorf("proto: unexpected end of group")
)
