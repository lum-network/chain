// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: lum/network/dfract/genesis.proto

package types

import (
	fmt "fmt"
	types "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/gogoproto/gogoproto"
	proto "github.com/cosmos/gogoproto/proto"
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

type GenesisState struct {
	ModuleAccountBalance      []types.Coin `protobuf:"bytes,1,rep,name=module_account_balance,json=moduleAccountBalance,proto3" json:"module_account_balance" yaml:"module_account_balance"`
	Params                    Params       `protobuf:"bytes,2,opt,name=params,proto3" json:"params" yaml:"params"`
	DepositsPendingWithdrawal []*Deposit   `protobuf:"bytes,3,rep,name=deposits_pending_withdrawal,json=depositsPendingWithdrawal,proto3" json:"deposits_pending_withdrawal,omitempty"`
	DepositsPendingMint       []*Deposit   `protobuf:"bytes,4,rep,name=deposits_pending_mint,json=depositsPendingMint,proto3" json:"deposits_pending_mint,omitempty"`
	DepositsMinted            []*Deposit   `protobuf:"bytes,5,rep,name=deposits_minted,json=depositsMinted,proto3" json:"deposits_minted,omitempty"`
}

func (m *GenesisState) Reset()         { *m = GenesisState{} }
func (m *GenesisState) String() string { return proto.CompactTextString(m) }
func (*GenesisState) ProtoMessage()    {}
func (*GenesisState) Descriptor() ([]byte, []int) {
	return fileDescriptor_1e0774872ec49cc5, []int{0}
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

func (m *GenesisState) GetModuleAccountBalance() []types.Coin {
	if m != nil {
		return m.ModuleAccountBalance
	}
	return nil
}

func (m *GenesisState) GetParams() Params {
	if m != nil {
		return m.Params
	}
	return Params{}
}

func (m *GenesisState) GetDepositsPendingWithdrawal() []*Deposit {
	if m != nil {
		return m.DepositsPendingWithdrawal
	}
	return nil
}

func (m *GenesisState) GetDepositsPendingMint() []*Deposit {
	if m != nil {
		return m.DepositsPendingMint
	}
	return nil
}

func (m *GenesisState) GetDepositsMinted() []*Deposit {
	if m != nil {
		return m.DepositsMinted
	}
	return nil
}

func init() {
	proto.RegisterType((*GenesisState)(nil), "lum.network.dfract.GenesisState")
}

func init() { proto.RegisterFile("lum/network/dfract/genesis.proto", fileDescriptor_1e0774872ec49cc5) }

var fileDescriptor_1e0774872ec49cc5 = []byte{
	// 401 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x92, 0xcd, 0xea, 0xd3, 0x40,
	0x14, 0xc5, 0x13, 0xab, 0x5d, 0xa4, 0x7e, 0x40, 0x6c, 0x25, 0x4d, 0x31, 0x2d, 0x01, 0xa1, 0x20,
	0xce, 0xd0, 0xba, 0x73, 0x67, 0xac, 0x88, 0x8b, 0x62, 0xa9, 0x0b, 0x41, 0x17, 0x61, 0x32, 0x19,
	0xd3, 0xc1, 0xcc, 0x4c, 0xc8, 0x4c, 0x5a, 0xfb, 0x16, 0xbe, 0x94, 0xd0, 0x65, 0x97, 0xae, 0x8a,
	0xb4, 0x6f, 0xe0, 0x13, 0x48, 0x33, 0x49, 0x17, 0xff, 0x06, 0xba, 0x0b, 0x73, 0xcf, 0xf9, 0x9d,
	0x73, 0xc3, 0xb5, 0x46, 0x69, 0xc1, 0x20, 0x27, 0x6a, 0x23, 0xf2, 0x1f, 0x30, 0xfe, 0x9e, 0x23,
	0xac, 0x60, 0x42, 0x38, 0x91, 0x54, 0x82, 0x2c, 0x17, 0x4a, 0xd8, 0x76, 0x5a, 0x30, 0x50, 0x29,
	0x80, 0x56, 0xb8, 0xdd, 0x44, 0x24, 0xa2, 0x1c, 0xc3, 0xf3, 0x97, 0x56, 0xba, 0x1e, 0x16, 0x92,
	0x09, 0x09, 0x23, 0x24, 0x09, 0x5c, 0x4f, 0x22, 0xa2, 0xd0, 0x04, 0x62, 0x41, 0x79, 0x35, 0x1f,
	0x36, 0x64, 0x65, 0x28, 0x47, 0xac, 0x8a, 0x72, 0x9b, 0xca, 0xc4, 0x24, 0x13, 0x92, 0x2a, 0xad,
	0xf0, 0x7f, 0xb7, 0xac, 0x87, 0x1f, 0x74, 0xbd, 0xcf, 0x0a, 0x29, 0x62, 0xaf, 0xad, 0x67, 0x4c,
	0xc4, 0x45, 0x4a, 0x42, 0x84, 0xb1, 0x28, 0xb8, 0x0a, 0x23, 0x94, 0x22, 0x8e, 0x89, 0x63, 0x8e,
	0x5a, 0xe3, 0xce, 0xb4, 0x0f, 0x74, 0x29, 0x70, 0x2e, 0x05, 0xaa, 0x52, 0xe0, 0x9d, 0xa0, 0x3c,
	0x78, 0xb1, 0x3b, 0x0c, 0x8d, 0x7f, 0x87, 0xe1, 0xf3, 0x2d, 0x62, 0xe9, 0x1b, 0xbf, 0x19, 0xe3,
	0x2f, 0xbb, 0x7a, 0xf0, 0x56, 0xbf, 0x07, 0xfa, 0xd9, 0xfe, 0x68, 0xb5, 0x75, 0x75, 0xe7, 0xde,
	0xc8, 0x1c, 0x77, 0xa6, 0x2e, 0xb8, 0xfe, 0x4d, 0x60, 0x51, 0x2a, 0x82, 0x5e, 0x15, 0xf4, 0x48,
	0x07, 0x69, 0x9f, 0xbf, 0xac, 0x00, 0xf6, 0x37, 0x6b, 0x50, 0x2d, 0x29, 0xc3, 0x8c, 0xf0, 0x98,
	0xf2, 0x24, 0xdc, 0x50, 0xb5, 0x8a, 0x73, 0xb4, 0x41, 0xa9, 0xd3, 0x2a, 0xf7, 0x18, 0x34, 0xf1,
	0x67, 0xda, 0xb6, 0xec, 0xd7, 0xfe, 0x85, 0xb6, 0x7f, 0xb9, 0xb8, 0xed, 0x4f, 0x56, 0xef, 0x0a,
	0xce, 0x28, 0x57, 0xce, 0xfd, 0xdb, 0xd8, 0xa7, 0x77, 0xb0, 0x73, 0xca, 0x95, 0x3d, 0xb3, 0x9e,
	0x5c, 0x80, 0x67, 0x10, 0x89, 0x9d, 0x07, 0xb7, 0x51, 0x8f, 0x6b, 0xcf, 0xbc, 0xb4, 0x04, 0xef,
	0x77, 0x47, 0xcf, 0xdc, 0x1f, 0x3d, 0xf3, 0xef, 0xd1, 0x33, 0x7f, 0x9d, 0x3c, 0x63, 0x7f, 0xf2,
	0x8c, 0x3f, 0x27, 0xcf, 0xf8, 0xfa, 0x32, 0xa1, 0x6a, 0x55, 0x44, 0x00, 0x0b, 0x06, 0xd3, 0x82,
	0xbd, 0xaa, 0xcf, 0x01, 0xaf, 0x10, 0xe5, 0xf0, 0x67, 0x7d, 0x16, 0x6a, 0x9b, 0x11, 0x19, 0xb5,
	0xcb, 0xab, 0x78, 0xfd, 0x3f, 0x00, 0x00, 0xff, 0xff, 0xf5, 0x23, 0xf7, 0x8c, 0xc6, 0x02, 0x00,
	0x00,
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
	if len(m.DepositsMinted) > 0 {
		for iNdEx := len(m.DepositsMinted) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.DepositsMinted[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x2a
		}
	}
	if len(m.DepositsPendingMint) > 0 {
		for iNdEx := len(m.DepositsPendingMint) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.DepositsPendingMint[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x22
		}
	}
	if len(m.DepositsPendingWithdrawal) > 0 {
		for iNdEx := len(m.DepositsPendingWithdrawal) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.DepositsPendingWithdrawal[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x1a
		}
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
	dAtA[i] = 0x12
	if len(m.ModuleAccountBalance) > 0 {
		for iNdEx := len(m.ModuleAccountBalance) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.ModuleAccountBalance[iNdEx].MarshalToSizedBuffer(dAtA[:i])
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
func (m *GenesisState) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.ModuleAccountBalance) > 0 {
		for _, e := range m.ModuleAccountBalance {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	l = m.Params.Size()
	n += 1 + l + sovGenesis(uint64(l))
	if len(m.DepositsPendingWithdrawal) > 0 {
		for _, e := range m.DepositsPendingWithdrawal {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.DepositsPendingMint) > 0 {
		for _, e := range m.DepositsPendingMint {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.DepositsMinted) > 0 {
		for _, e := range m.DepositsMinted {
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
				return fmt.Errorf("proto: wrong wireType = %d for field ModuleAccountBalance", wireType)
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
			m.ModuleAccountBalance = append(m.ModuleAccountBalance, types.Coin{})
			if err := m.ModuleAccountBalance[len(m.ModuleAccountBalance)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
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
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field DepositsPendingWithdrawal", wireType)
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
			m.DepositsPendingWithdrawal = append(m.DepositsPendingWithdrawal, &Deposit{})
			if err := m.DepositsPendingWithdrawal[len(m.DepositsPendingWithdrawal)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field DepositsPendingMint", wireType)
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
			m.DepositsPendingMint = append(m.DepositsPendingMint, &Deposit{})
			if err := m.DepositsPendingMint[len(m.DepositsPendingMint)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field DepositsMinted", wireType)
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
			m.DepositsMinted = append(m.DepositsMinted, &Deposit{})
			if err := m.DepositsMinted[len(m.DepositsMinted)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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
