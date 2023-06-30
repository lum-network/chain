// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: lum/network/dfract/proposal.proto

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

// DEPRECATED:
// For easier management, we moved the WithdrawAndMintProposal to tx based
// minting. The withdrawal address specified in the dFract module parameters is
// the one authorized to withdraw and mint udfr tokens based on the
// MicroMintRate
type WithdrawAndMintProposal struct {
	Title             string `protobuf:"bytes,1,opt,name=title,proto3" json:"title,omitempty"`
	Description       string `protobuf:"bytes,2,opt,name=description,proto3" json:"description,omitempty"`
	WithdrawalAddress string `protobuf:"bytes,3,opt,name=withdrawal_address,json=withdrawalAddress,proto3" json:"withdrawal_address,omitempty"`
	MicroMintRate     int64  `protobuf:"varint,4,opt,name=micro_mint_rate,json=microMintRate,proto3" json:"micro_mint_rate,omitempty"`
}

func (m *WithdrawAndMintProposal) Reset()      { *m = WithdrawAndMintProposal{} }
func (*WithdrawAndMintProposal) ProtoMessage() {}
func (*WithdrawAndMintProposal) Descriptor() ([]byte, []int) {
	return fileDescriptor_2f2bef36c5201cb4, []int{0}
}
func (m *WithdrawAndMintProposal) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *WithdrawAndMintProposal) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_WithdrawAndMintProposal.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *WithdrawAndMintProposal) XXX_Merge(src proto.Message) {
	xxx_messageInfo_WithdrawAndMintProposal.Merge(m, src)
}
func (m *WithdrawAndMintProposal) XXX_Size() int {
	return m.Size()
}
func (m *WithdrawAndMintProposal) XXX_DiscardUnknown() {
	xxx_messageInfo_WithdrawAndMintProposal.DiscardUnknown(m)
}

var xxx_messageInfo_WithdrawAndMintProposal proto.InternalMessageInfo

func init() {
	proto.RegisterType((*WithdrawAndMintProposal)(nil), "lum.network.dfract.WithdrawAndMintProposal")
}

func init() { proto.RegisterFile("lum/network/dfract/proposal.proto", fileDescriptor_2f2bef36c5201cb4) }

var fileDescriptor_2f2bef36c5201cb4 = []byte{
	// 288 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x52, 0xcc, 0x29, 0xcd, 0xd5,
	0xcf, 0x4b, 0x2d, 0x29, 0xcf, 0x2f, 0xca, 0xd6, 0x4f, 0x49, 0x2b, 0x4a, 0x4c, 0x2e, 0xd1, 0x2f,
	0x28, 0xca, 0x2f, 0xc8, 0x2f, 0x4e, 0xcc, 0xd1, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x12, 0xca,
	0x29, 0xcd, 0xd5, 0x83, 0x2a, 0xd1, 0x83, 0x28, 0x91, 0x12, 0x49, 0xcf, 0x4f, 0xcf, 0x07, 0x4b,
	0xeb, 0x83, 0x58, 0x10, 0x95, 0x4a, 0xdb, 0x18, 0xb9, 0xc4, 0xc3, 0x33, 0x4b, 0x32, 0x52, 0x8a,
	0x12, 0xcb, 0x1d, 0xf3, 0x52, 0x7c, 0x33, 0xf3, 0x4a, 0x02, 0xa0, 0x66, 0x09, 0x89, 0x70, 0xb1,
	0x96, 0x64, 0x96, 0xe4, 0xa4, 0x4a, 0x30, 0x2a, 0x30, 0x6a, 0x70, 0x06, 0x41, 0x38, 0x42, 0x0a,
	0x5c, 0xdc, 0x29, 0xa9, 0xc5, 0xc9, 0x45, 0x99, 0x05, 0x25, 0x99, 0xf9, 0x79, 0x12, 0x4c, 0x60,
	0x39, 0x64, 0x21, 0x21, 0x5d, 0x2e, 0xa1, 0x72, 0xa8, 0x91, 0x89, 0x39, 0xf1, 0x89, 0x29, 0x29,
	0x45, 0xa9, 0xc5, 0xc5, 0x12, 0xcc, 0x60, 0x85, 0x82, 0x08, 0x19, 0x47, 0x88, 0x84, 0x90, 0x1a,
	0x17, 0x7f, 0x6e, 0x66, 0x72, 0x51, 0x7e, 0x7c, 0x6e, 0x66, 0x5e, 0x49, 0x7c, 0x51, 0x62, 0x49,
	0xaa, 0x04, 0x8b, 0x02, 0xa3, 0x06, 0x73, 0x10, 0x2f, 0x58, 0x18, 0xe4, 0xa4, 0xa0, 0xc4, 0x92,
	0x54, 0x2b, 0x9e, 0x8e, 0x05, 0xf2, 0x0c, 0x33, 0x16, 0xc8, 0x33, 0xbc, 0x58, 0x20, 0xcf, 0xe8,
	0xe4, 0x7a, 0xe2, 0x91, 0x1c, 0xe3, 0x85, 0x47, 0x72, 0x8c, 0x0f, 0x1e, 0xc9, 0x31, 0x4e, 0x78,
	0x2c, 0xc7, 0x70, 0xe1, 0xb1, 0x1c, 0xc3, 0x8d, 0xc7, 0x72, 0x0c, 0x51, 0xda, 0xe9, 0x99, 0x25,
	0x19, 0xa5, 0x49, 0x7a, 0xc9, 0xf9, 0xb9, 0xfa, 0x39, 0xa5, 0xb9, 0xba, 0xb0, 0xa0, 0x4a, 0xce,
	0x48, 0xcc, 0xcc, 0xd3, 0xaf, 0x80, 0x05, 0x59, 0x49, 0x65, 0x41, 0x6a, 0x71, 0x12, 0x1b, 0x38,
	0x18, 0x8c, 0x01, 0x01, 0x00, 0x00, 0xff, 0xff, 0x0f, 0xe1, 0x90, 0xf8, 0x55, 0x01, 0x00, 0x00,
}

func (this *WithdrawAndMintProposal) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*WithdrawAndMintProposal)
	if !ok {
		that2, ok := that.(WithdrawAndMintProposal)
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
	if this.WithdrawalAddress != that1.WithdrawalAddress {
		return false
	}
	if this.MicroMintRate != that1.MicroMintRate {
		return false
	}
	return true
}
func (m *WithdrawAndMintProposal) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *WithdrawAndMintProposal) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *WithdrawAndMintProposal) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.MicroMintRate != 0 {
		i = encodeVarintProposal(dAtA, i, uint64(m.MicroMintRate))
		i--
		dAtA[i] = 0x20
	}
	if len(m.WithdrawalAddress) > 0 {
		i -= len(m.WithdrawalAddress)
		copy(dAtA[i:], m.WithdrawalAddress)
		i = encodeVarintProposal(dAtA, i, uint64(len(m.WithdrawalAddress)))
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
func (m *WithdrawAndMintProposal) Size() (n int) {
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
	l = len(m.WithdrawalAddress)
	if l > 0 {
		n += 1 + l + sovProposal(uint64(l))
	}
	if m.MicroMintRate != 0 {
		n += 1 + sovProposal(uint64(m.MicroMintRate))
	}
	return n
}

func sovProposal(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozProposal(x uint64) (n int) {
	return sovProposal(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *WithdrawAndMintProposal) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: WithdrawAndMintProposal: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: WithdrawAndMintProposal: illegal tag %d (wire type %d)", fieldNum, wire)
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
				return fmt.Errorf("proto: wrong wireType = %d for field WithdrawalAddress", wireType)
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
			m.WithdrawalAddress = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field MicroMintRate", wireType)
			}
			m.MicroMintRate = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProposal
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.MicroMintRate |= int64(b&0x7F) << shift
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
			if (skippy < 0) || (iNdEx+skippy) < 0 {
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
