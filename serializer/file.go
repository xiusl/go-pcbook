package serializer

import (
    "fmt"
    "io/ioutil"

    "github.com/golang/protobuf/proto"
)

// WriteProtobufToJSONFile 将 proto 消息写入本地 json 文件
func WriteProtobufToJSONFile(message proto.Message, filename string) error {
    data, err := ConvertProtobufToJSON(message)
    if err != nil {
        return fmt.Errorf("cannot marshal protobuf to json string: %w", err)
    }

    err = ioutil.WriteFile(filename, []byte(data), 0644)
    if err != nil {
        return fmt.Errorf("cannot write data to file: %w", err)
    }
    return nil
}

// ReadProtobufFromJSONFile 从本地 json 文件读取 proto 消息
func ReadProtobufFromJSONFile(filename string, message proto.Message) error {
    data, err := ioutil.ReadFile(filename)
    if err != nil {
        return fmt.Errorf("cannot read data from file: %w", err)
    }

    err = ConvertJSONToProtobuf(string(data), message)
    if err != nil {
        return fmt.Errorf("cannot ummarshal data: %w", err)
    }
    return nil
}

// WriteProtobufToBinaryFile 将 proto 消息写入二进制文件
// 这里指的就是 Laptop 对象
func WriteProtobufToBinaryFile(message proto.Message, filename string) error {
    data, err := proto.Marshal(message)
    if err != nil {
        return fmt.Errorf("cannot marshal message to binary: %w", err)
    }

    err = ioutil.WriteFile(filename, data, 0644)
    if err != nil {
        return fmt.Errorf("cannot write data to file: %w", err)
    }
    return nil
}

// ReadProtobufFromBinaryFile 从二进制文件中读取 proto 消息
func ReadProtobufFromBinaryFile(filename string, message proto.Message) error {
    data, err := ioutil.ReadFile(filename)
    if err != nil {
        return fmt.Errorf("cannot read data from file: %w", err)
    }

    err = proto.Unmarshal(data, message)
    if err != nil {
        return fmt.Errorf("cannot ummarshal data: %w", err)
    }
    return nil
}
