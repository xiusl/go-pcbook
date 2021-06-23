# RPC 拦截器

## 需要做的
- 了解拦截器
- 实现 RPC 授权访问

## 定义服务拦截器

```go
func unaryInterceptor(
    ctx context.Context,
    req interface{},
    info *grpc.UnaryServerInfo,
    handler grpc.UnaryHandler,
) (interface{}, error) {

    log.Println("---> unary interceptor: ", info.FullMethod)
    return handler(ctx, req)
}

func streamInterceptor(
    srv interface{},
    ss grpc.ServerStream,
    info *grpc.StreamServerInfo,
    handler grpc.StreamHandler,
) error {

    log.Println("---> stream interceptor: ", info.FullMethod)
    return handler(srv, ss)
}

func main() {
    // ...
    grpcServer := grpc.NewServer(
        grpc.UnaryInterceptor(unaryInterceptor),
        grpc.StreamInterceptor(streamInterceptor),
    )
    // ...
}
```
