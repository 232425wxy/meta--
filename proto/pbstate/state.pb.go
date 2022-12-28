// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: state.proto

package pbstate

import (
	fmt "fmt"
	pbtypes "github.com/232425wxy/meta--/proto/pbtypes"
	_ "github.com/cosmos/gogoproto/gogoproto"
	proto "github.com/cosmos/gogoproto/proto"
	github_com_cosmos_gogoproto_types "github.com/cosmos/gogoproto/types"
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

type State struct {
	InitialHeight   int64                 `protobuf:"varint,1,opt,name=initial_height,json=initialHeight,proto3" json:"initial_height,omitempty"`
	LastBlockHeight int64                 `protobuf:"varint,2,opt,name=last_block_height,json=lastBlockHeight,proto3" json:"last_block_height,omitempty"`
	LastBlock       pbtypes.SimpleBlock   `protobuf:"bytes,3,opt,name=last_block,json=lastBlock,proto3" json:"last_block"`
	LastBlockTime   time.Time             `protobuf:"bytes,4,opt,name=last_block_time,json=lastBlockTime,proto3,stdtime" json:"last_block_time"`
	Validators      *pbtypes.ValidatorSet `protobuf:"bytes,5,opt,name=validators,proto3" json:"validators,omitempty"`
}

func (m *State) Reset()         { *m = State{} }
func (m *State) String() string { return proto.CompactTextString(m) }
func (*State) ProtoMessage()    {}
func (*State) Descriptor() ([]byte, []int) {
	return fileDescriptor_a888679467bb7853, []int{0}
}
func (m *State) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *State) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_State.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *State) XXX_Merge(src proto.Message) {
	xxx_messageInfo_State.Merge(m, src)
}
func (m *State) XXX_Size() int {
	return m.Size()
}
func (m *State) XXX_DiscardUnknown() {
	xxx_messageInfo_State.DiscardUnknown(m)
}

var xxx_messageInfo_State proto.InternalMessageInfo

func (m *State) GetInitialHeight() int64 {
	if m != nil {
		return m.InitialHeight
	}
	return 0
}

func (m *State) GetLastBlockHeight() int64 {
	if m != nil {
		return m.LastBlockHeight
	}
	return 0
}

func (m *State) GetLastBlock() pbtypes.SimpleBlock {
	if m != nil {
		return m.LastBlock
	}
	return pbtypes.SimpleBlock{}
}

func (m *State) GetLastBlockTime() time.Time {
	if m != nil {
		return m.LastBlockTime
	}
	return time.Time{}
}

func (m *State) GetValidators() *pbtypes.ValidatorSet {
	if m != nil {
		return m.Validators
	}
	return nil
}

func init() {
	proto.RegisterType((*State)(nil), "pbstate.State")
}

func init() { proto.RegisterFile("state.proto", fileDescriptor_a888679467bb7853) }

var fileDescriptor_a888679467bb7853 = []byte{
	// 318 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x5c, 0x90, 0x4d, 0x4b, 0x3b, 0x31,
	0x10, 0xc6, 0x37, 0x7d, 0xf9, 0xff, 0x35, 0xa5, 0x16, 0x17, 0x85, 0xb5, 0x60, 0x5a, 0x44, 0xa1,
	0x08, 0x66, 0xd1, 0x2a, 0xe8, 0x75, 0x4f, 0x1e, 0x3c, 0x6d, 0xc5, 0x6b, 0xc9, 0x6a, 0xdc, 0x06,
	0x53, 0x13, 0xba, 0x51, 0xf0, 0x5b, 0xf4, 0xec, 0x27, 0xea, 0xb1, 0x47, 0x4f, 0x2a, 0xbb, 0x5f,
	0x44, 0x32, 0xfb, 0xd2, 0xea, 0x6d, 0x66, 0x9e, 0x5f, 0x9e, 0x99, 0x27, 0xb8, 0x95, 0x18, 0x66,
	0x38, 0xd5, 0x33, 0x65, 0x94, 0xfb, 0x5f, 0x47, 0xd0, 0x76, 0xf7, 0xa0, 0xf7, 0x75, 0x64, 0xde,
	0x34, 0x4f, 0xfc, 0x48, 0xaa, 0xfb, 0xa7, 0x9c, 0xe9, 0xee, 0xff, 0x96, 0x5e, 0x99, 0x14, 0x0f,
	0xcc, 0xa8, 0x59, 0x21, 0x1f, 0xc6, 0x2a, 0x56, 0x50, 0x9e, 0x9c, 0xd2, 0x73, 0x3a, 0xf4, 0xab,
	0x1e, 0xaa, 0x82, 0xba, 0xfc, 0x4b, 0x41, 0x1d, 0xbd, 0x3c, 0xfa, 0xb1, 0x52, 0xb1, 0xe4, 0xab,
	0xde, 0x88, 0x29, 0x4f, 0x0c, 0x9b, 0xea, 0xfc, 0xe5, 0xc1, 0x7b, 0x0d, 0x37, 0x47, 0xf6, 0x46,
	0xf7, 0x08, 0x6f, 0x89, 0x67, 0x61, 0x04, 0x93, 0xe3, 0x09, 0x17, 0xf1, 0xc4, 0x78, 0xa8, 0x8f,
	0x06, 0xf5, 0xb0, 0x5d, 0x4c, 0xaf, 0x61, 0xe8, 0x1e, 0xe3, 0x6d, 0xc9, 0x12, 0x33, 0x86, 0x0c,
	0x25, 0x59, 0x03, 0xb2, 0x63, 0x85, 0xc0, 0xce, 0x0b, 0xf6, 0x0a, 0xe3, 0x15, 0xeb, 0xd5, 0xfb,
	0x68, 0xd0, 0x3a, 0xdb, 0xa1, 0x45, 0x54, 0x3a, 0x12, 0x53, 0x2d, 0x39, 0xf0, 0x41, 0x63, 0xf1,
	0xd9, 0x73, 0xc2, 0xcd, 0xca, 0xc0, 0xbd, 0xc1, 0x9d, 0xb5, 0x35, 0xf6, 0x6a, 0xaf, 0x01, 0xef,
	0xbb, 0x34, 0x8f, 0x44, 0xcb, 0x48, 0xf4, 0xb6, 0x8c, 0x14, 0x6c, 0x58, 0x97, 0xf9, 0x57, 0x0f,
	0x85, 0xed, 0xca, 0xc9, 0xaa, 0xee, 0x05, 0xc6, 0xd5, 0xc7, 0x26, 0x5e, 0x13, 0x8c, 0x76, 0xab,
	0x43, 0xee, 0x4a, 0x69, 0xc4, 0x4d, 0xb8, 0x06, 0x06, 0xde, 0x22, 0x25, 0x68, 0x99, 0x12, 0xf4,
	0x9d, 0x12, 0x34, 0xcf, 0x88, 0xb3, 0xcc, 0x88, 0xf3, 0x91, 0x11, 0x27, 0xfa, 0x07, 0xdb, 0x87,
	0x3f, 0x01, 0x00, 0x00, 0xff, 0xff, 0x7a, 0xcb, 0x11, 0xa2, 0xef, 0x01, 0x00, 0x00,
}

func (m *State) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *State) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *State) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Validators != nil {
		{
			size, err := m.Validators.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintState(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x2a
	}
	n2, err2 := github_com_cosmos_gogoproto_types.StdTimeMarshalTo(m.LastBlockTime, dAtA[i-github_com_cosmos_gogoproto_types.SizeOfStdTime(m.LastBlockTime):])
	if err2 != nil {
		return 0, err2
	}
	i -= n2
	i = encodeVarintState(dAtA, i, uint64(n2))
	i--
	dAtA[i] = 0x22
	{
		size, err := m.LastBlock.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintState(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x1a
	if m.LastBlockHeight != 0 {
		i = encodeVarintState(dAtA, i, uint64(m.LastBlockHeight))
		i--
		dAtA[i] = 0x10
	}
	if m.InitialHeight != 0 {
		i = encodeVarintState(dAtA, i, uint64(m.InitialHeight))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func encodeVarintState(dAtA []byte, offset int, v uint64) int {
	offset -= sovState(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *State) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.InitialHeight != 0 {
		n += 1 + sovState(uint64(m.InitialHeight))
	}
	if m.LastBlockHeight != 0 {
		n += 1 + sovState(uint64(m.LastBlockHeight))
	}
	l = m.LastBlock.Size()
	n += 1 + l + sovState(uint64(l))
	l = github_com_cosmos_gogoproto_types.SizeOfStdTime(m.LastBlockTime)
	n += 1 + l + sovState(uint64(l))
	if m.Validators != nil {
		l = m.Validators.Size()
		n += 1 + l + sovState(uint64(l))
	}
	return n
}

func sovState(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozState(x uint64) (n int) {
	return sovState(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *State) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowState
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
			return fmt.Errorf("proto: State: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: State: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field InitialHeight", wireType)
			}
			m.InitialHeight = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowState
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.InitialHeight |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field LastBlockHeight", wireType)
			}
			m.LastBlockHeight = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowState
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.LastBlockHeight |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field LastBlock", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowState
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
				return ErrInvalidLengthState
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthState
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.LastBlock.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field LastBlockTime", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowState
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
				return ErrInvalidLengthState
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthState
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := github_com_cosmos_gogoproto_types.StdTimeUnmarshal(&m.LastBlockTime, dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Validators", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowState
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
				return ErrInvalidLengthState
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthState
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Validators == nil {
				m.Validators = &pbtypes.ValidatorSet{}
			}
			if err := m.Validators.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipState(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthState
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
func skipState(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowState
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
					return 0, ErrIntOverflowState
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
					return 0, ErrIntOverflowState
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
				return 0, ErrInvalidLengthState
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupState
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthState
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthState        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowState          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupState = fmt.Errorf("proto: unexpected end of group")
)
