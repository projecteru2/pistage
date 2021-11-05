// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.17.3
// source: apiserver/grpc/proto/pistage.proto

package proto

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type ApplyPistageRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Content string `protobuf:"bytes,1,opt,name=content,proto3" json:"content,omitempty"`
}

func (x *ApplyPistageRequest) Reset() {
	*x = ApplyPistageRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_apiserver_grpc_proto_pistage_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ApplyPistageRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ApplyPistageRequest) ProtoMessage() {}

func (x *ApplyPistageRequest) ProtoReflect() protoreflect.Message {
	mi := &file_apiserver_grpc_proto_pistage_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ApplyPistageRequest.ProtoReflect.Descriptor instead.
func (*ApplyPistageRequest) Descriptor() ([]byte, []int) {
	return file_apiserver_grpc_proto_pistage_proto_rawDescGZIP(), []int{0}
}

func (x *ApplyPistageRequest) GetContent() string {
	if x != nil {
		return x.Content
	}
	return ""
}

type ApplyPistageOnewayReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	WorkflowNamespace  string `protobuf:"bytes,1,opt,name=workflowNamespace,proto3" json:"workflowNamespace,omitempty"`
	WorkflowIdentifier string `protobuf:"bytes,2,opt,name=workflowIdentifier,proto3" json:"workflowIdentifier,omitempty"`
	Success            bool   `protobuf:"varint,3,opt,name=success,proto3" json:"success,omitempty"`
}

func (x *ApplyPistageOnewayReply) Reset() {
	*x = ApplyPistageOnewayReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_apiserver_grpc_proto_pistage_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ApplyPistageOnewayReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ApplyPistageOnewayReply) ProtoMessage() {}

func (x *ApplyPistageOnewayReply) ProtoReflect() protoreflect.Message {
	mi := &file_apiserver_grpc_proto_pistage_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ApplyPistageOnewayReply.ProtoReflect.Descriptor instead.
func (*ApplyPistageOnewayReply) Descriptor() ([]byte, []int) {
	return file_apiserver_grpc_proto_pistage_proto_rawDescGZIP(), []int{1}
}

func (x *ApplyPistageOnewayReply) GetWorkflowNamespace() string {
	if x != nil {
		return x.WorkflowNamespace
	}
	return ""
}

func (x *ApplyPistageOnewayReply) GetWorkflowIdentifier() string {
	if x != nil {
		return x.WorkflowIdentifier
	}
	return ""
}

func (x *ApplyPistageOnewayReply) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

type ApplyPistageStreamReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	WorkflowNamespace  string `protobuf:"bytes,1,opt,name=workflowNamespace,proto3" json:"workflowNamespace,omitempty"`
	WorkflowIdentifier string `protobuf:"bytes,2,opt,name=workflowIdentifier,proto3" json:"workflowIdentifier,omitempty"`
	Logtype            int64  `protobuf:"varint,3,opt,name=logtype,proto3" json:"logtype,omitempty"`
	Log                string `protobuf:"bytes,4,opt,name=log,proto3" json:"log,omitempty"`
}

func (x *ApplyPistageStreamReply) Reset() {
	*x = ApplyPistageStreamReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_apiserver_grpc_proto_pistage_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ApplyPistageStreamReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ApplyPistageStreamReply) ProtoMessage() {}

func (x *ApplyPistageStreamReply) ProtoReflect() protoreflect.Message {
	mi := &file_apiserver_grpc_proto_pistage_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ApplyPistageStreamReply.ProtoReflect.Descriptor instead.
func (*ApplyPistageStreamReply) Descriptor() ([]byte, []int) {
	return file_apiserver_grpc_proto_pistage_proto_rawDescGZIP(), []int{2}
}

func (x *ApplyPistageStreamReply) GetWorkflowNamespace() string {
	if x != nil {
		return x.WorkflowNamespace
	}
	return ""
}

func (x *ApplyPistageStreamReply) GetWorkflowIdentifier() string {
	if x != nil {
		return x.WorkflowIdentifier
	}
	return ""
}

func (x *ApplyPistageStreamReply) GetLogtype() int64 {
	if x != nil {
		return x.Logtype
	}
	return 0
}

func (x *ApplyPistageStreamReply) GetLog() string {
	if x != nil {
		return x.Log
	}
	return ""
}

type RollbackPistageRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Content string `protobuf:"bytes,1,opt,name=content,proto3" json:"content,omitempty"`
}

func (x *RollbackPistageRequest) Reset() {
	*x = RollbackPistageRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_apiserver_grpc_proto_pistage_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RollbackPistageRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RollbackPistageRequest) ProtoMessage() {}

func (x *RollbackPistageRequest) ProtoReflect() protoreflect.Message {
	mi := &file_apiserver_grpc_proto_pistage_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RollbackPistageRequest.ProtoReflect.Descriptor instead.
func (*RollbackPistageRequest) Descriptor() ([]byte, []int) {
	return file_apiserver_grpc_proto_pistage_proto_rawDescGZIP(), []int{3}
}

func (x *RollbackPistageRequest) GetContent() string {
	if x != nil {
		return x.Content
	}
	return ""
}

type RollbackReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	WorkflowNamespace  string `protobuf:"bytes,1,opt,name=workflowNamespace,proto3" json:"workflowNamespace,omitempty"`
	WorkflowIdentifier string `protobuf:"bytes,2,opt,name=workflowIdentifier,proto3" json:"workflowIdentifier,omitempty"`
	Success            bool   `protobuf:"varint,3,opt,name=success,proto3" json:"success,omitempty"`
}

func (x *RollbackReply) Reset() {
	*x = RollbackReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_apiserver_grpc_proto_pistage_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RollbackReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RollbackReply) ProtoMessage() {}

func (x *RollbackReply) ProtoReflect() protoreflect.Message {
	mi := &file_apiserver_grpc_proto_pistage_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RollbackReply.ProtoReflect.Descriptor instead.
func (*RollbackReply) Descriptor() ([]byte, []int) {
	return file_apiserver_grpc_proto_pistage_proto_rawDescGZIP(), []int{4}
}

func (x *RollbackReply) GetWorkflowNamespace() string {
	if x != nil {
		return x.WorkflowNamespace
	}
	return ""
}

func (x *RollbackReply) GetWorkflowIdentifier() string {
	if x != nil {
		return x.WorkflowIdentifier
	}
	return ""
}

func (x *RollbackReply) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

type RollbackPistageStreamReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	WorkflowNamespace  string `protobuf:"bytes,1,opt,name=workflowNamespace,proto3" json:"workflowNamespace,omitempty"`
	WorkflowIdentifier string `protobuf:"bytes,2,opt,name=workflowIdentifier,proto3" json:"workflowIdentifier,omitempty"`
	Logtype            int64  `protobuf:"varint,3,opt,name=logtype,proto3" json:"logtype,omitempty"`
	Log                string `protobuf:"bytes,4,opt,name=log,proto3" json:"log,omitempty"`
}

func (x *RollbackPistageStreamReply) Reset() {
	*x = RollbackPistageStreamReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_apiserver_grpc_proto_pistage_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RollbackPistageStreamReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RollbackPistageStreamReply) ProtoMessage() {}

func (x *RollbackPistageStreamReply) ProtoReflect() protoreflect.Message {
	mi := &file_apiserver_grpc_proto_pistage_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RollbackPistageStreamReply.ProtoReflect.Descriptor instead.
func (*RollbackPistageStreamReply) Descriptor() ([]byte, []int) {
	return file_apiserver_grpc_proto_pistage_proto_rawDescGZIP(), []int{5}
}

func (x *RollbackPistageStreamReply) GetWorkflowNamespace() string {
	if x != nil {
		return x.WorkflowNamespace
	}
	return ""
}

func (x *RollbackPistageStreamReply) GetWorkflowIdentifier() string {
	if x != nil {
		return x.WorkflowIdentifier
	}
	return ""
}

func (x *RollbackPistageStreamReply) GetLogtype() int64 {
	if x != nil {
		return x.Logtype
	}
	return 0
}

func (x *RollbackPistageStreamReply) GetLog() string {
	if x != nil {
		return x.Log
	}
	return ""
}

var File_apiserver_grpc_proto_pistage_proto protoreflect.FileDescriptor

var file_apiserver_grpc_proto_pistage_proto_rawDesc = []byte{
	0x0a, 0x22, 0x61, 0x70, 0x69, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2f, 0x67, 0x72, 0x70, 0x63,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x70, 0x69, 0x73, 0x74, 0x61, 0x67, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x12, 0x05, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x2f, 0x0a, 0x13, 0x41,
	0x70, 0x70, 0x6c, 0x79, 0x50, 0x69, 0x73, 0x74, 0x61, 0x67, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x22, 0x91, 0x01, 0x0a,
	0x17, 0x41, 0x70, 0x70, 0x6c, 0x79, 0x50, 0x69, 0x73, 0x74, 0x61, 0x67, 0x65, 0x4f, 0x6e, 0x65,
	0x77, 0x61, 0x79, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x12, 0x2c, 0x0a, 0x11, 0x77, 0x6f, 0x72, 0x6b,
	0x66, 0x6c, 0x6f, 0x77, 0x4e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x11, 0x77, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x4e, 0x61, 0x6d,
	0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x12, 0x2e, 0x0a, 0x12, 0x77, 0x6f, 0x72, 0x6b, 0x66, 0x6c,
	0x6f, 0x77, 0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x66, 0x69, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x12, 0x77, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x49, 0x64, 0x65, 0x6e,
	0x74, 0x69, 0x66, 0x69, 0x65, 0x72, 0x12, 0x18, 0x0a, 0x07, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73,
	0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73,
	0x22, 0xa3, 0x01, 0x0a, 0x17, 0x41, 0x70, 0x70, 0x6c, 0x79, 0x50, 0x69, 0x73, 0x74, 0x61, 0x67,
	0x65, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x12, 0x2c, 0x0a, 0x11,
	0x77, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x4e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63,
	0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x11, 0x77, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f,
	0x77, 0x4e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x12, 0x2e, 0x0a, 0x12, 0x77, 0x6f,
	0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x66, 0x69, 0x65, 0x72,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x12, 0x77, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77,
	0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x66, 0x69, 0x65, 0x72, 0x12, 0x18, 0x0a, 0x07, 0x6c, 0x6f,
	0x67, 0x74, 0x79, 0x70, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x03, 0x52, 0x07, 0x6c, 0x6f, 0x67,
	0x74, 0x79, 0x70, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x6c, 0x6f, 0x67, 0x18, 0x04, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x03, 0x6c, 0x6f, 0x67, 0x22, 0x32, 0x0a, 0x16, 0x52, 0x6f, 0x6c, 0x6c, 0x62, 0x61,
	0x63, 0x6b, 0x50, 0x69, 0x73, 0x74, 0x61, 0x67, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x12, 0x18, 0x0a, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x22, 0x87, 0x01, 0x0a, 0x0d, 0x52,
	0x6f, 0x6c, 0x6c, 0x62, 0x61, 0x63, 0x6b, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x12, 0x2c, 0x0a, 0x11,
	0x77, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x4e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63,
	0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x11, 0x77, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f,
	0x77, 0x4e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x12, 0x2e, 0x0a, 0x12, 0x77, 0x6f,
	0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x66, 0x69, 0x65, 0x72,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x12, 0x77, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77,
	0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x66, 0x69, 0x65, 0x72, 0x12, 0x18, 0x0a, 0x07, 0x73, 0x75,
	0x63, 0x63, 0x65, 0x73, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x73, 0x75, 0x63,
	0x63, 0x65, 0x73, 0x73, 0x22, 0xa6, 0x01, 0x0a, 0x1a, 0x52, 0x6f, 0x6c, 0x6c, 0x62, 0x61, 0x63,
	0x6b, 0x50, 0x69, 0x73, 0x74, 0x61, 0x67, 0x65, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x52, 0x65,
	0x70, 0x6c, 0x79, 0x12, 0x2c, 0x0a, 0x11, 0x77, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x4e,
	0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x11,
	0x77, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x4e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63,
	0x65, 0x12, 0x2e, 0x0a, 0x12, 0x77, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x49, 0x64, 0x65,
	0x6e, 0x74, 0x69, 0x66, 0x69, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x12, 0x77,
	0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x66, 0x69, 0x65,
	0x72, 0x12, 0x18, 0x0a, 0x07, 0x6c, 0x6f, 0x67, 0x74, 0x79, 0x70, 0x65, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x03, 0x52, 0x07, 0x6c, 0x6f, 0x67, 0x74, 0x79, 0x70, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x6c,
	0x6f, 0x67, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6c, 0x6f, 0x67, 0x32, 0xc6, 0x02,
	0x0a, 0x07, 0x50, 0x69, 0x73, 0x74, 0x61, 0x67, 0x65, 0x12, 0x4b, 0x0a, 0x0b, 0x41, 0x70, 0x70,
	0x6c, 0x79, 0x4f, 0x6e, 0x65, 0x77, 0x61, 0x79, 0x12, 0x1a, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x2e, 0x41, 0x70, 0x70, 0x6c, 0x79, 0x50, 0x69, 0x73, 0x74, 0x61, 0x67, 0x65, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x1e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x41, 0x70, 0x70,
	0x6c, 0x79, 0x50, 0x69, 0x73, 0x74, 0x61, 0x67, 0x65, 0x4f, 0x6e, 0x65, 0x77, 0x61, 0x79, 0x52,
	0x65, 0x70, 0x6c, 0x79, 0x22, 0x00, 0x12, 0x4d, 0x0a, 0x0b, 0x41, 0x70, 0x70, 0x6c, 0x79, 0x53,
	0x74, 0x72, 0x65, 0x61, 0x6d, 0x12, 0x1a, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x41, 0x70,
	0x70, 0x6c, 0x79, 0x50, 0x69, 0x73, 0x74, 0x61, 0x67, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x1e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x41, 0x70, 0x70, 0x6c, 0x79, 0x50,
	0x69, 0x73, 0x74, 0x61, 0x67, 0x65, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x52, 0x65, 0x70, 0x6c,
	0x79, 0x22, 0x00, 0x30, 0x01, 0x12, 0x47, 0x0a, 0x0e, 0x52, 0x6f, 0x6c, 0x6c, 0x62, 0x61, 0x63,
	0x6b, 0x4f, 0x6e, 0x65, 0x77, 0x61, 0x79, 0x12, 0x1d, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e,
	0x52, 0x6f, 0x6c, 0x6c, 0x62, 0x61, 0x63, 0x6b, 0x50, 0x69, 0x73, 0x74, 0x61, 0x67, 0x65, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x14, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x52,
	0x6f, 0x6c, 0x6c, 0x62, 0x61, 0x63, 0x6b, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x00, 0x12, 0x56,
	0x0a, 0x0e, 0x52, 0x6f, 0x6c, 0x6c, 0x62, 0x61, 0x63, 0x6b, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d,
	0x12, 0x1d, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x52, 0x6f, 0x6c, 0x6c, 0x62, 0x61, 0x63,
	0x6b, 0x50, 0x69, 0x73, 0x74, 0x61, 0x67, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a,
	0x21, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x52, 0x6f, 0x6c, 0x6c, 0x62, 0x61, 0x63, 0x6b,
	0x50, 0x69, 0x73, 0x74, 0x61, 0x67, 0x65, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x52, 0x65, 0x70,
	0x6c, 0x79, 0x22, 0x00, 0x30, 0x01, 0x42, 0x35, 0x5a, 0x33, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62,
	0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x65, 0x72, 0x75, 0x32,
	0x2f, 0x70, 0x69, 0x73, 0x74, 0x61, 0x67, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x73, 0x65, 0x72, 0x76,
	0x65, 0x72, 0x2f, 0x67, 0x72, 0x70, 0x63, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_apiserver_grpc_proto_pistage_proto_rawDescOnce sync.Once
	file_apiserver_grpc_proto_pistage_proto_rawDescData = file_apiserver_grpc_proto_pistage_proto_rawDesc
)

func file_apiserver_grpc_proto_pistage_proto_rawDescGZIP() []byte {
	file_apiserver_grpc_proto_pistage_proto_rawDescOnce.Do(func() {
		file_apiserver_grpc_proto_pistage_proto_rawDescData = protoimpl.X.CompressGZIP(file_apiserver_grpc_proto_pistage_proto_rawDescData)
	})
	return file_apiserver_grpc_proto_pistage_proto_rawDescData
}

var file_apiserver_grpc_proto_pistage_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_apiserver_grpc_proto_pistage_proto_goTypes = []interface{}{
	(*ApplyPistageRequest)(nil),        // 0: proto.ApplyPistageRequest
	(*ApplyPistageOnewayReply)(nil),    // 1: proto.ApplyPistageOnewayReply
	(*ApplyPistageStreamReply)(nil),    // 2: proto.ApplyPistageStreamReply
	(*RollbackPistageRequest)(nil),     // 3: proto.RollbackPistageRequest
	(*RollbackReply)(nil),              // 4: proto.RollbackReply
	(*RollbackPistageStreamReply)(nil), // 5: proto.RollbackPistageStreamReply
}
var file_apiserver_grpc_proto_pistage_proto_depIdxs = []int32{
	0, // 0: proto.Pistage.ApplyOneway:input_type -> proto.ApplyPistageRequest
	0, // 1: proto.Pistage.ApplyStream:input_type -> proto.ApplyPistageRequest
	3, // 2: proto.Pistage.RollbackOneway:input_type -> proto.RollbackPistageRequest
	3, // 3: proto.Pistage.RollbackStream:input_type -> proto.RollbackPistageRequest
	1, // 4: proto.Pistage.ApplyOneway:output_type -> proto.ApplyPistageOnewayReply
	2, // 5: proto.Pistage.ApplyStream:output_type -> proto.ApplyPistageStreamReply
	4, // 6: proto.Pistage.RollbackOneway:output_type -> proto.RollbackReply
	5, // 7: proto.Pistage.RollbackStream:output_type -> proto.RollbackPistageStreamReply
	4, // [4:8] is the sub-list for method output_type
	0, // [0:4] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_apiserver_grpc_proto_pistage_proto_init() }
func file_apiserver_grpc_proto_pistage_proto_init() {
	if File_apiserver_grpc_proto_pistage_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_apiserver_grpc_proto_pistage_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ApplyPistageRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_apiserver_grpc_proto_pistage_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ApplyPistageOnewayReply); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_apiserver_grpc_proto_pistage_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ApplyPistageStreamReply); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_apiserver_grpc_proto_pistage_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RollbackPistageRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_apiserver_grpc_proto_pistage_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RollbackReply); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_apiserver_grpc_proto_pistage_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RollbackPistageStreamReply); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_apiserver_grpc_proto_pistage_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_apiserver_grpc_proto_pistage_proto_goTypes,
		DependencyIndexes: file_apiserver_grpc_proto_pistage_proto_depIdxs,
		MessageInfos:      file_apiserver_grpc_proto_pistage_proto_msgTypes,
	}.Build()
	File_apiserver_grpc_proto_pistage_proto = out.File
	file_apiserver_grpc_proto_pistage_proto_rawDesc = nil
	file_apiserver_grpc_proto_pistage_proto_goTypes = nil
	file_apiserver_grpc_proto_pistage_proto_depIdxs = nil
}
