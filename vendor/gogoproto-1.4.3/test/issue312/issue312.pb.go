// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: issue312.proto

package issue312

import (
	fmt "fmt"
	_ "github.com/cosmos/gogoproto/gogoproto"
	proto "github.com/cosmos/gogoproto/proto"
	math "math"
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

type TaskState int32

const (
	TaskState_TASK_STAGING  TaskState = 6
	TaskState_TASK_STARTING TaskState = 0
	TaskState_TASK_RUNNING  TaskState = 1
)

var TaskState_name = map[int32]string{
	6: "TASK_STAGING",
	0: "TASK_STARTING",
	1: "TASK_RUNNING",
}

var TaskState_value = map[string]int32{
	"TASK_STAGING":  6,
	"TASK_STARTING": 0,
	"TASK_RUNNING":  1,
}

func (x TaskState) Enum() *TaskState {
	p := new(TaskState)
	*p = x
	return p
}

func (x TaskState) String() string {
	return proto.EnumName(TaskState_name, int32(x))
}

func (x *TaskState) UnmarshalJSON(data []byte) error {
	value, err := proto.UnmarshalJSONEnum(TaskState_value, data, "TaskState")
	if err != nil {
		return err
	}
	*x = TaskState(value)
	return nil
}

func (TaskState) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_8a64932ccacef062, []int{0}
}

func init() {
	proto.RegisterEnum("issue312.TaskState", TaskState_name, TaskState_value)
}

func init() { proto.RegisterFile("issue312.proto", fileDescriptor_8a64932ccacef062) }

var fileDescriptor_8a64932ccacef062 = []byte{
	// 129 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0xcb, 0x2c, 0x2e, 0x2e,
	0x4d, 0x35, 0x36, 0x34, 0xd2, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0xe2, 0x80, 0xf1, 0xa5, 0x44,
	0xd2, 0xf3, 0xd3, 0xf3, 0xc1, 0x82, 0xfa, 0x20, 0x16, 0x44, 0x5e, 0xcb, 0x89, 0x8b, 0x33, 0x24,
	0xb1, 0x38, 0x3b, 0xb8, 0x24, 0xb1, 0x24, 0x55, 0x48, 0x80, 0x8b, 0x27, 0xc4, 0x31, 0xd8, 0x3b,
	0x3e, 0x38, 0xc4, 0xd1, 0xdd, 0xd3, 0xcf, 0x5d, 0x80, 0x4d, 0x48, 0x90, 0x8b, 0x17, 0x26, 0x12,
	0x14, 0x02, 0x12, 0x62, 0x80, 0x2b, 0x0a, 0x0a, 0xf5, 0xf3, 0x03, 0x89, 0x30, 0x3a, 0x49, 0x7d,
	0x78, 0x28, 0xc7, 0xf8, 0xe3, 0xa1, 0x1c, 0xe3, 0x8a, 0x47, 0x72, 0x8c, 0x3b, 0x1e, 0xc9, 0x31,
	0x46, 0xc1, 0x6d, 0x05, 0x04, 0x00, 0x00, 0xff, 0xff, 0xa6, 0x1b, 0x46, 0x67, 0x90, 0x00, 0x00,
	0x00,
}
