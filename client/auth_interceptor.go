package client

import (
    "context"
    "log"
    "time"

    "google.golang.org/grpc"
    "google.golang.org/grpc/metadata"
)

// AuthInterceptor 客户端授权拦截器
type AuthInterceptor struct {
    authClient  *AuthClient
    authMethods map[string]bool
    accessToken string
}

// NewAuthInterceptor 创建新的客户端授权拦截器
func NewAuthInterceptor(
    authClient *AuthClient,
    authMethods map[string]bool,
    refreshDuration time.Duration,
) (*AuthInterceptor, error) {
    interceptor := &AuthInterceptor{
        authClient:  authClient,
        authMethods: authMethods,
    }

    err := interceptor.scheduleRefreshToken(refreshDuration)
    if err != nil {
        return nil, err
    }
    return interceptor, nil
}

// Unary 一元客户端授权拦截器
func (interceptor *AuthInterceptor) Unary() grpc.UnaryClientInterceptor {
    return func(
        ctx context.Context,
        method string,
        req, reply interface{},
        cc *grpc.ClientConn,
        invoker grpc.UnaryInvoker,
        opts ...grpc.CallOption,
    ) error {
        log.Printf("---> unary interceptor: %v", method)

        if interceptor.authMethods[method] {
            return invoker(interceptor.attachToken(ctx), method, req, reply, cc, opts...)
        }
        return invoker(ctx, method, req, reply, cc, opts...)
    }
}

// Stream 一元客户端授权拦截器
func (interceptor *AuthInterceptor) Stream() grpc.StreamClientInterceptor {
    return func(
        ctx context.Context,
        desc *grpc.StreamDesc,
        cc *grpc.ClientConn,
        method string,
        streamer grpc.Streamer,
        opts ...grpc.CallOption,
    ) (grpc.ClientStream, error) {
        log.Printf("---> stream interceptor: %v", method)
        if interceptor.authMethods[method] {
            return streamer(interceptor.attachToken(ctx), desc, cc, method, opts...)
        }
        return streamer(ctx, desc, cc, method, opts...)
    }
}

func (interceptor *AuthInterceptor) attachToken(ctx context.Context) context.Context {
    return metadata.AppendToOutgoingContext(ctx, "authorization", interceptor.accessToken)
}

func (interceptor *AuthInterceptor) scheduleRefreshToken(duration time.Duration) error {
    err := interceptor.refreshToken()
    if err != nil {
        return err
    }
    // 开启后台协程，定时刷新 Token
    go func() {
        wait := duration
        for {
            time.Sleep(wait)
            err := interceptor.refreshToken()
            if err != nil {
                wait = time.Second
            } else {
                wait = duration
            }
        }
    }()

    return nil
}

func (interceptor *AuthInterceptor) refreshToken() error {
    accessToken, err := interceptor.authClient.Login()
    if err != nil {
        return err
    }
    log.Printf("Token Refreshed: %v\n", accessToken)
    interceptor.accessToken = accessToken
    return nil
}
