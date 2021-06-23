package service

import (
    "context"

    "github.com/xiusl/pcbook/pb"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

// AuthServer 授权服务
type AuthServer struct {
    userStore  UserStore
    jwtManager *JWTManager
}

// NewAuthServer 创建一个授权服务
func NewAuthServer(userStore UserStore, jwtManager *JWTManager) *AuthServer {
    return &AuthServer{
        userStore:  userStore,
        jwtManager: jwtManager,
    }
}

// Login 用户登录 RPC
func (server *AuthServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
    user, err := server.userStore.Find(req.GetUsername())
    if err != nil {
        return nil, status.Errorf(codes.Internal, "cannot find user: %v", err)
    }

    if user == nil || !user.IsCorrentPassword(req.GetPassword()) {
        return nil, status.Errorf(codes.NotFound, "incorrect username/password")
    }

    token, err := server.jwtManager.Generate(user)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "cannot generate access token")
    }

    resp := &pb.LoginResponse{AccessToken: token}
    return resp, nil
}
