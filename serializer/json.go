package serializer

import (
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
)

// ConvertProtobufToJSON 将 protobuf buffer message 转换为 json 字符串
func ConvertProtobufToJSON(message proto.Message) (string, error) {
	marshaler := jsonpb.Marshaler{
		EnumsAsInts:  false,
		EmitDefaults: true,
		Indent:       "  ",
		OrigName:     true,
	}
	return marshaler.MarshalToString(message)
}

// ConvertJSONToProtobuf 将 json 字符串转换为 protobuf buffer message
func ConvertJSONToProtobuf(data string, message proto.Message) error {
	return jsonpb.UnmarshalString(data, message)
}
