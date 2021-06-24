package service

import (
    "context"
    "log"

    "google.golang.org/grpc"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/metadata"
    "google.golang.org/grpc/status"
)

// AuthInterceptor 服务授权拦截器
type AuthInterceptor struct {
    jwtManager      *JWTManager
    accessibleRoles map[string][]string
}

// NewAuthInterceptor 新建一个服务授权拦截器
func NewAuthInterceptor(jwtManager *JWTManager, accessibleRoles map[string][]string) *AuthInterceptor {
    return &AuthInterceptor{jwtManager, accessibleRoles}
}

// Unary 一元 RPC 授权拦截器
func (interceptor *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
    return func(
        ctx context.Context,
        req interface{},
        info *grpc.UnaryServerInfo,
        handler grpc.UnaryHandler,
    ) (interface{}, error) {

        log.Println("---> unary interceptor: ", info.FullMethod)

        if err := interceptor.authorize(ctx, info.FullMethod); err != nil {
            return nil, err
        }
        return handler(ctx, req)
    }
}

// Stream 流式 RPC 授权拦截器
func (interceptor *AuthInterceptor) Stream() grpc.StreamServerInterceptor {
    return func(
        srv interface{},
        ss grpc.ServerStream,
        info *grpc.StreamServerInfo,
        handler grpc.StreamHandler,
    ) error {
        log.Println("---> stream interceptor: ", info.FullMethod)
        if err := interceptor.authorize(ss.Context(), info.FullMethod); err != nil {
            return err
        }
        return handler(srv, ss)
    }
}

func (interceptor *AuthInterceptor) authorize(ctx context.Context, method string) error {
    accessibleRoles, ok := interceptor.accessibleRoles[method]
    if !ok {
        return nil
    }

    md, ok := metadata.FromIncomingContext(ctx)
    if !ok {
        return status.Errorf(codes.Unauthenticated, "metadata is not provided")
    }

    values := md["authorization"]
    if len(values) == 0 {
        return status.Errorf(codes.Unauthenticated, "authorization token is not provided")
    }

    accessToken := values[0]
    claims, err := interceptor.jwtManager.Verify(accessToken)
    if err != nil {
        return status.Errorf(codes.Unauthenticated, "authorization token is invalid")
    }

    for _, role := range accessibleRoles {
        if role == claims.Role {
            return nil
        }
    }
    return status.Errorf(codes.PermissionDenied, "no permission to access this RPC")
}
