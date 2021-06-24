# RPC 拦截器

## 需要做的
- 了解拦截器
- 实现 RPC 服务端拦截器校验授权信息
- 使用 JWT 生成令牌
- 实现 RPC 客户端拦截器自动附加令牌

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

## 新增用户模块和存储
- 在 service 目录下创建 `user.go`，用来保存用户信息

  ```go
  type User struct {
      Username       string
      HashedPassword string
      Role           string
  }
  func NewUser(username, password, role string) (*User, error) {
      // ...
  }
  func (user *User) IsCorrentPassword(password string) bool {
      // ...
  }
  func (user *User) Clone() *User {
      //...
  }
  ```

- 在 service 目录下穿件 `user_store.go`，用户的持久化存储，并实现内存存储

  ```go
  // UserStore 存储用户接口
  type UserStore interface {
      Save(user *User) error
      Find(username string) (*User, error)
  }
  type InMemoryUserStore struct {
      mutex sync.RWMutex
      users map[string]*User
  }
  func NewInMemoryUserStore() *InMemoryUserStore {}
  func (store *InMemoryUserStore) Save(user *User) error {}
  func (store *InMemoryUserStore) Find(username string) (*User, error) {}
  ```

## 定义授权 RPC 服务
- 在 proto 目录下新建 `auth_service.proto`

  ```protobuf
  message LoginRequest {
      string username = 1;
      string password = 2;
  }
  message LoginResponse {
      string access_token = 1;
  }
  service AuthService {
      rpc Login(LoginRequest) returns (LoginResponse) {}
  }
  ```

- 生成代码，`pb` 目录下自动创建 `auth_service.pb.go`

  ```
  make gen
  ```

## 使用 JWT 进行令牌生成和校验

- 在 service 中新建 `jwt_manager.go`，定义管理对象，Claim

  ```go
  type JWTManager struct {
      secretKey     string
      tokenDuration time.Duration
  }
  type UserClaims struct {
      jwt.StandardClaims
      Username string `json:"username"`
      Role     string `json:"role"`
  }
  ```

- 实现工厂函数，令牌生成方法，令牌校验方法

  ```go
  func NewJWTManager(secretKey string, duration time.Duration) *JWTManager {}
  func (manager *JWTManager) Generate(user *User) (string, error) {}
  func (manager *JWTManager) Verify(tokenString string) (*UserClaims, error) {}
  ```

## 实现 RPC 授权服务
- 在 service 中新建 `auth_server.go`

- 定义 `AuthServer` 对象，包括一个存储接口和 `jwt` 管理对象
  ```go
  type AuthServer struct {
      userStore  UserStore
      jwtManager *JWTManager
  }
  ```

- 实现 RPC 的登录方法
  ```go
  func (server *AuthServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
    // 根据用户名获取用户
    user, err := server.userStore.Find(req.GetUsername())
    // 校验用户密码
    if user == nil || !user.IsCorrentPassword(req.GetPassword()) {}
    // 生成令牌
    token, err := server.jwtManager.Generate(user)
    // 返回响应
    resp := &pb.LoginResponse{AccessToken: token}
    return resp, nil
  }
  ```

## 实现 RPC 服务的拦截器

- 在 service 中创建 `auth_interceptor.go`
- 定义 `AuthInterceptor` 结构体

  `accessibleRoles` 针对不同服务方法，限制用户的访问
  ```go
  type AuthInterceptor struct {
    jwtManager      *JWTManager
    accessibleRoles map[string][]string
  }
  func NewAuthInterceptor(jwtManager *JWTManager, accessibleRoles map[string][]string) *AuthInterceptor {}
  ```
- 实现一元和流式服务的拦截器
  ```go
  func (interceptor *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {}
  func (interceptor *AuthInterceptor) Stream() grpc.StreamServerInterceptor {}
  ```
- 增加统一的方法来判断权限
  ```go
  // service/auth_interceptor.go 部分代码有省略
  func (interceptor *AuthInterceptor) authorize(ctx context.Context, method string) error {
      // 根据服务方法，获取可调用该方法的用户角色
      accessibleRoles, ok := interceptor.accessibleRoles[method]
      if !ok { return nil } // 没有定义的话就是不需要授权、
      // 从上下文中获取令牌数据
      md, ok := metadata.FromIncomingContext(ctx)
      values := md["authorization"]
      accessToken := values[0]
      // 校验并获取相关信息
      claims, err := interceptor.jwtManager.Verify(accessToken)
      // 验证角色
      for _, role := range accessibleRoles {
          if role == claims.Role {
              return nil
          }
      }
      return status.Errorf(codes.PermissionDenied, "no permission to access this RPC")
  }
  ```
- 在服务初始化时，添加定义好的拦截器
  ```go
  // cmd/server/main.go
  const (
      secretKey     = "njkandsiaud"  // 需要在环境变量中获取
      tokenDuration = 10 * time.Minute
  )
  // 创建两个种子用户
  func seedUser(userStroe service.UserStore) error {
      err := createUser(userStroe, "admin", "abc", "admin")
      if err != nil {
          return err
      }
      return createUser(userStroe, "user1", "abc", "user")
  }

  func createUser(userStroe service.UserStore, username, password, role string) error {
      user, err := service.NewUser(username, password, role)
      if err != nil {
          return err
      }
      return userStroe.Save(user)
  }

  // 配置服务方法的权限
  func accessibleRoles() map[string][]string {
      const latopServicePath = "/xiusl.pcbook.LaptopServices/"
      return map[string][]string{
          latopServicePath + "CreateLaptop": {"admin"},
          latopServicePath + "UploadImage":  {"admin"},
          latopServicePath + "RateLaptop":   {"admin", "user"},
      }
  }

  func main() {
      // ...
      userStore := service.NewInMemoryUserStore()
      if err := seedUser(userStore); err != nil {
          log.Fatal("cannot create seed users: %w", err)
      }
      jwtManager := service.NewJWTManager(secretKey, tokenDuration)
      authServer := service.NewAuthServer(userStore, jwtManager)

      // ...

      interceptor := service.NewAuthInterceptor(jwtManager, accessibleRoles())
      grpcServer := grpc.NewServer(
          grpc.UnaryInterceptor(interceptor.Unary()),
          grpc.StreamInterceptor(interceptor.Stream()),
      )
      pb.RegisterAuthServiceServer(grpcServer, authServer)
      pb.RegisterLaptopServicesServer(grpcServer, laptopServer)
      reflection.Register(grpcServer)

      // ...
  }
  ```

## 在 RPC 客户端拦截器中附加 Token

- 首先对 `cmd/client/main.go` 中的代码重构
  - 新建 `client` 目录，增加 `laptop_client.go` 文件，定义 `LaptopClient` 对象

    ```go
    type LaptopClient struct {
        server pb.LaptopServicesClient
    }
    func NewLaptopClient(cc *grpc.ClientConn) *LaptopClient {}
    ```
  - 实现 Laptop 的一些方法

    ```go
    func (clien *LaptopClient) CreateLaptop(laptop *pb.Laptop) {}
    func (client *LaptopClient) SearchLaptop(filter *pb.Filter) {}
    func (clien *LaptopClient) UploadImage(laptopID, imagePath string) {}
    func (client *LaptopClient) RateLaptop(laptopIDs []string, scores []float64) error {}
    ```
  - 修复 `cmd/client/main.go` 的错误
    ```go
    func testCreateLaptop(laptopClient *client.LaptopClient) {
        laptopClient.CreateLaptop(sample.NewLaptop())
    }

    func testSearchLaptop(laptopClient *client.LaptopClient) {
        // laptopClient.SearchLaptop(filter)
    }
    func testUploadImage(laptopClient *client.LaptopClient) {
        laptop := sample.NewLaptop()
        laptopClient.CreateLaptop(laptop)
        laptopClient.UploadImage(laptop.GetId(), "tmp/pc.png")
    }
    func testRatingLaptop(laptopClient *client.LaptopClient) {
        // laptopClient.CreateLaptop(laptop)
        // err := laptopClient.RateLaptop(laptopIDs, scores)
    }
    func main() {
        // ...
        laptopClient := client.NewLaptopClient(conn)
        testRatingLaptop(laptopClient)
    }
    ```

- 此时在没有为客户端增加登录的时候测试，将会出现权限错误的问题
- 新增授权客户端，在 client 目录新建 `auth_client.go`
- `auth_client.go` 创建 `AuthClient` 结构体

  ```go
  type AuthClient struct {
      server   pb.AuthServiceClient
      username string
      password string
  }
  func NewAuthClient(cc *grpc.ClientConn, username, password string) *AuthClient {}
  func (client *AuthClient) Login() (string, error) {}
  ```
- 新增客户端授权拦截器，在 client 目录新建 `auth_interceptor.go`
  ```go
  type AuthInterceptor struct {
      authClient  *AuthClient
      authMethods map[string]bool
      accessToken string
  }
  func NewAuthInterceptor(
      authClient *AuthClient,
      authMethods map[string]bool,
      refreshDuration time.Duration,
  ) (*AuthInterceptor, error) {
      // 获取 Token 并定时刷新 Token
      err := interceptor.scheduleRefreshToken(refreshDuration)
  }
  func (interceptor *AuthInterceptor) Unary() grpc.UnaryClientInterceptor {
      // 向上下文注入Token
      // return invoker(interceptor.attachToken(ctx), method, req, reply, cc, opts...)
  }
  func (interceptor *AuthInterceptor) Stream() grpc.StreamClientInterceptor {
      // 向上下文注入Token
      // return streamer(interceptor.attachToken(ctx), desc, cc, method, opts...)
  }

  func (interceptor *AuthInterceptor) attachToken(ctx context.Context) context.Context {}
  func (interceptor *AuthInterceptor) scheduleRefreshToken(duration time.Duration) error {}
  func (interceptor *AuthInterceptor) refreshToken() error {}
  ```
- 在客户端初始化时，增加拦截器
  ```go
  // cmd/client/main.go
  const (
      username        = "user1"
      password        = "abc"
      refreshDuration = 30 * time.Second
  )

  // 定义需要增加 token 的服务方法
  func authMethods() map[string]bool {
      const latopServicePath = "/xiusl.pcbook.LaptopServices/"
      return map[string]bool{
          latopServicePath + "CreateLaptop": true,
          latopServicePath + "UploadImage":  true,
          latopServicePath + "RateLaptop":   true,
      }
  }
  func main() {
      // ...

      // 初始化授权客户端
      authClient := client.NewAuthClient(conn, username, password)

      // 初始化授权客户端拦截器
      interceptor, err := client.NewAuthInterceptor(authClient, authMethods(), refreshDuration)

      // 创建一个新的连接，这个连接拥有两个拦截器
      conn1, err := grpc.Dial(
          *addr,
          grpc.WithInsecure(),
          grpc.WithUnaryInterceptor(interceptor.Unary()),
          grpc.WithStreamInterceptor(interceptor.Stream()),
      )
      laptopClient := client.NewLaptopClient(conn1)
      testRatingLaptop(laptopClient)
  }

- 简单测试

  修改 `cmd/client/main.go` 的 username 和 password，当 username = admin 并密码正确的情况，运行无误，其它情况会返回错误


**--本节结束--**
