package serializer_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xiusl/pcbook/pb"
	"github.com/xiusl/pcbook/sample"
	"github.com/xiusl/pcbook/serializer"
	"google.golang.org/protobuf/proto"
)

func TestFileSerializer(t *testing.T) {

	binaryFile := "../tmp/laptop.bin"

	laptop1 := sample.MewLaptop()

	err := serializer.WriteProtobufToBinaryFile(laptop1, binaryFile)
	require.NoError(t, err)

	laptop2 := &pb.Laptop{}
	err = serializer.ReadProtobufFromBinaryFile(binaryFile, laptop2)
	require.NoError(t, err)
	require.True(t, proto.Equal(laptop1, laptop2))
}
