package client

import (
    "context"
    "time"

    "github.com/xiusl/pcbook/pb"
    "google.golang.org/grpc"
)

// AuthClient 调用授权 RPC 的客户端
type AuthClient struct {
    server   pb.AuthServiceClient
    username string
    password string
}

// NewAuthClient 创建一个新的授权客户端
func NewAuthClient(cc *grpc.ClientConn, username, password string) *AuthClient {
    server := pb.NewAuthServiceClient(cc)
    return &AuthClient{server, username, password}
}

// Login 用户登录并返回令牌 Token
func (client *AuthClient) Login() (string, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    req := &pb.LoginRequest{
        Username: client.username,
        Password: client.password,
    }

    res, err := client.server.Login(ctx, req)
    if err != nil {
        return "", err
    }
    return res.GetAccessToken(), nil
}
