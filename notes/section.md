# 实现一元的 gRPC api

## 需要做的

- 定义一个 proto 服务，包含创建一台便携计算机的一元 RPC
- 实现一个服务器来处理请求，并将信息保存到内存中
- 实现一个客户端去调用这个服务
- 为客户端和服务端的交互编写单元测试
- 学习如何处理错误，返回对应的错误码，和 gRPC 的 deadline

## 定义 proto 服务

- 在 proto 目录下新建 `laptop_service.proto`，定义一个 laptop 服务。

  包括了一个请求消息，一个响应消息，在服务里有一个创建 laptop 的声明

  ```go
  message CreateLaptopRequest {
      Laptop laptop = 1;
  }
  
  message CreateLaptopResponse {
      string id = 1;
  }
  
  service LaptopServices {
      rpc CreateLaptop(CreateLaptopRequest) returns (CreateLaptopResponse) {}
  }
  ```

  

- 执行 `make gen`，在 pd 目录下将生成一个 `laptop_service.pb.go` 文件，一些内容如下

  ```go
  type CreateLaptopRequest struct {
      // ...
  	Laptop *Laptop `protobuf:"bytes,1,opt,name=laptop,proto3" json:"laptop,omitempty"`
  }
  type CreateLaptopResponse struct {
  	// ...
  	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
  }
  // LaptopServicesServer is the server API for LaptopServices service.
  type LaptopServicesServer interface {
  	CreateLaptop(context.Context, *CreateLaptopRequest) (*CreateLaptopResponse, error)
  }
  ```

  

## 定义一个服务器

- 在 service 中新建一个 `laptop_server.go` 文件

- 定义一个 LaptopServer

  ```go
  type LaptopServer struct {
  	store LaptopStore
  }
  
  func NewLaptopServer() *LaptopServer {
  	return &LaptopServer{
  		store: NewInMemoryLaptopStore(),
  	}
  }
  ```

  

- LaptopServer 包含了一个 laptop 存储接口，在 service 中创建 `laptop_store.go`

  ```go
  type LaptopStore interface {
      Save(laptop *pb.Laptop) error
  }
  ```

  

- 在 `laptop_store.go` 中，新建 LaptopStore 的内存存储实现 `ImMemoryLaptopStore`

  ```go
  type InMemoryLaptopStore struct {
  	mu   sync.Mutex
  	data map[string]*pb.Laptop
  }
  
  func NewInMemoryLaptopStore() *InMemoryLaptopStore {
  	//...
  }
  
  func (store *InMemoryLaptopStore) Save(laptop *pb.Laptop) error {
  	// ...
  }
  ```

  

- 实现 RPC 的 `CreateLaptop` 方法

  ```go
  func (server *LaptopServer) CreateLaptop(ctx context.Context, req *pb.CreateLaptopRequest) (*pb.CreateLaptopResponse, error) {
  	// ...
  }
  ```

  

## 编写单元测试

- 在 service 中创建 `laptop_server_test.go` ，编写测试用例

  ```go
  laptopNoID := &pb.Laptop{}
  laptopNoID.Id = ""
  
  laptopInvalidID := &pb.Laptop{}
  laptopInvalidID.Id = "invalid-id"
  
  laptopDuplicateID := sample.NewLaptop()
  storeDuplicateID := service.NewInMemoryLaptopStore()
  err := storeDuplicateID.Save(laptopDuplicateID)
  require.NoError(t, err)
  
  testCases := []struct {
  		name   string
  		laptop *pb.Laptop
  		store  service.LaptopStore
  		code   codes.Code
  	}{
  		{
  			name:   "success_with_id",
  			laptop: sample.NewLaptop(),
  			store:  service.NewInMemoryLaptopStore(),
  			code:   codes.OK,
  		},
  		{
  			name:   "success_no_id",
  			laptop: laptopNoID,
  			store:  service.NewInMemoryLaptopStore(),
  			code:   codes.OK,
  		},
  		{
  			name:   "failure_invalid_id",
  			laptop: laptopInvalidID,
  			store:  service.NewInMemoryLaptopStore(),
  			code:   codes.InvalidArgument,
  		},
  		{
  			name:   "failure_duplicate_id",
  			laptop: laptopDuplicateID,
  			store:  storeDuplicateID,
  			code:   codes.AlreadyExists,
  		},
  	}
  ```

  

- 测试

  ```go
  for _, tc := range testCases {
  		t.Run(tc.name, func(t *testing.T) {
  
  			req := &pb.CreateLaptopRequest{
  				Laptop: tc.laptop,
  			}
  
  			srv := service.NewLaptopServer(tc.store)
  
  			resp, err := srv.CreateLaptop(context.Background(), req)
  
  			if tc.code == codes.OK {
  				require.NoError(t, err)
  				require.NotNil(t, resp)
  				fmt.Println(tc.laptop.Id)
  				fmt.Println(resp)
  				require.NotEmpty(t, resp.Id)
  				if len(tc.laptop.Id) > 0 {
  					require.Equal(t, resp.Id, tc.laptop.Id)
  				}
  			} else {
  				require.Error(t, err)
  				require.Nil(t, resp)
  				st, ok := status.FromError(err)
  				require.True(t, ok)
  				require.Equal(t, tc.code, st.Code())
  			}
  		})
  }
  ```



## 编写 客户端测试

- 在 service 创建 `laptop_client_test.go` ，创建测试方法

  ```go
  func TestClientCreateLaptop(t *testing.T) {
  		// ...
  }
  ```

  

- 启动一个 grpc 服务器，添加 `startTestLaptopServer` 函数，返回一个 kaotop 服务，和服务的地址

  ```go
  func startTestLaptopServer(t *testing.T)  (*service.LaptopServer, string) {
      laptopServer := service.NewLaptopServer(service.NewInMemoryLaptopStore())
  
      grpcServer := grpc.NewServer()
      pb.RegisterLaptopServicesServer(grpcServer, laptopServer)
  
      listen, err := net.Listen("tcp", ":0")
      require.NoError(t, err)
  
      go grpcServer.Serve(listen)
  
      return laptopServer, listen.Addr().String()
  }
  ```

  

- 创建一个 grpc 客户端，添加 `newTestLaptopClient` 函数，根据 addr 地址，返回 `LaptopServicesClient` 对象

  ```go
  func newTestLaptopClient(t *testing.T, addr string) pb.LaptopServicesClient {
      conn, err := grpc.Dial(addr, grpc.WithInsecure())
      require.NoError(t, err)
      return pb.NewLaptopServicesClient(conn)
  }
  ```

  

- 完善测试代码

  ```go
  func TestClientCreateLaptop(t *testing.T) {
      laptopServer, serverAddr := startTestLaptopServer(t)
      laptopClient := newTestLaptopClient(t, serverAddr)
  
      laptop := sample.NewLaptop()
      expectedID := laptop.Id
  
      req := &pb.CreateLaptopRequest{
        Laptop: laptop,
      }
  
      resp, err := laptopClient.CreateLaptop(context.Background(), req)
      require.NoError(t, err)
      require.NotNil(t, resp.Id)
      require.Equal(t, resp.Id, expectedID)
  
      other, err := laptopServer.Store.FindByID(expectedID)
      require.NoError(t, err)
      require.NotNil(t, other.Id)
      require.Equal(t, other.Id, expectedID)
  
      requireSameLaptop(t, other, laptop)
  }
  ```

  

- `LaptopStore` 需要新增一个函数 `FindByID`，来验证是否存储成功，编辑 `laptop_store.go`

  ```go
  type LaptopStore interface {
  	// ...
  	FindByID(id string) (*pb.Laptop, error)
  }
  
  func (store *InMemoryLaptopStore) FindByID(id string) (*pb.Laptop, error) {
    //...
  }
  ```

  

- 运行测试

  ```
  Running tool: /usr/local/go/bin/go test -timeout 30s -run ^TestClientCreateLaptop$ github.com/xiusl/pcbook/service
  
  ok  	github.com/xiusl/pcbook/service	0.595s
  ```

## 实现真正的 Server & Client

- 删除主目录的 `main.go`

- 新建 `cmd/server` 和 `cmd/client` 目录，在两个目录下创建 `main.go`

  ```
  - cmd/
    - server/
      - main.go
    - client/
      - main.go
  ```

  

- 修改 `Makefile`

  ```makefile
  server:
    go run cmd/server/main.go
  client:
    go run cmd/client/main.go
  ```

  

- 完善 `cmd/server/main.go`

  ```go
  func main() {
      port := flag.String("port", "", "server port")
      flag.Parse()
      log.Printf("start server on port: %s", *port)
  
      laptopServer := service.NewLaptopServer(service.NewInMemoryLaptopStore())
      grpcServer := grpc.NewServer()
      pb.RegisterLaptopServicesServer(grpcServer, laptopServer)
  
      address := fmt.Sprintf("0.0.0.0:%s", *port)
      listener, err := net.Listen("tcp", address)
      if err != nil {
        log.Fatalf("cannot start server: %v", err)
      }
  
      err = grpcServer.Serve(listener)
      if err != nil {
        log.Fatalf("cannot start server: %v", err)
      }
  }
  ```

  

- 完善 `cmd/client/main.go`

  ```go
  func main() {
      addr := flag.String("addr", "", "the server address")
      flag.Parse()
      log.Printf("dial server: %s", *addr)
  
      conn, err := grpc.Dial(*addr, grpc.WithInsecure())
      if err != nil {
        log.Fatalf("cannot dial server: %v", err)
      }
  
      laptopClient := pb.NewLaptopServicesClient(conn)
  
      laptop := sample.NewLaptop()
      req := &pb.CreateLaptopRequest{
        Laptop: laptop,
      }
  
      res, err := laptopClient.CreateLaptop(context.Background(), req)
      if err != nil {
        st, ok := status.FromError(err)
        if ok && st.Code() == codes.AlreadyExists {
          log.Println("laptop already exists.")
        } else {
          log.Printf("laptop create error: %v", err)
        }
        return
      }
  
      log.Printf("created laptop success, id: %v", res.Id)
  }
  
  ```

  

- 更新 `Makefile`

  ```make
  server:
  	go run cmd/server/main.go -port=8080
  client:
  	go run cmd/client/main.go -addr="0.0.0.0:8080"
  ```

  

- 运行服务

  ```shell
  $ make server
  > go run cmd/server/main.go -port=8080
  > 2021/06/18 23:03:19 start server on port: 8080
  ```

  

- 运行客户端

  ```shell
  $ make client
  > go run cmd/client/main.go -addr="0.0.0.:8080"
  > 2021/06/18 23:13:26 dial server: 0.0.0.0:8080
  > 2021/06/18 23:13:26 created laptop success, id: d53003bc-0118-4e91-93d8-e7d04c9b7a75
  
  // 服务端日志
  > 2021/06/18 23:13:26 receive a create-laptop request with id:d53003bc-0118-4e91-93d8-e7d04c9b7a75.
  ```

  

  - 修改客户端代码，验证 laptop exist

    ```
    laptop := sample.NewLaptop()
    laptop.Id = "d53003bc-0118-4e91-93d8-e7d04c9b7a75"
    
    // 运行
    $ make client
    > go run cmd/client/main.go -addr="0.0.0.0:8080"
    > 2021/06/18 23:14:00 dial server: 0.0.0.0:8080
    > 2021/06/18 23:14:00 laptop already exists.
    ```

    

  - 修改客户端代码，验证 valid laptop id

    ```
    laptop := sample.NewLaptop()
    laptop.Id = "invalid-id"
    
    // 运行
    ￥ make client
    > go run cmd/client/main.go -addr="0.0.0.0:8080"
    > 2021/06/18 23:14:10 dial server: 0.0.0.0:8080
    > 2021/06/18 23:14:10 laptop create error: rpc error: code = InvalidArgument desc = laptap ID is not a valid UUID: invalid UUID length: 10
    ```

## gRPC 超时和取消 

- 修改 `cmd/client/main.go` ，对上下文增加超时时间

  ```go
  func main() {
      // ...
    
      // 1 秒后将超时退出
      ctx, cancel := context.WithTimeout(context.Background(), time.Second)
      defer cancel()
  
      res, err := laptopClient.CreateLaptop(ctx, req)
      // ...
  }
  ```

- 为对超时进行测试，修改 `service/laptop_server.go` ，增加耗时操作

  ```go
  // service/laptop_server.go
  func (server *LaptopServer) CreateLaptop(ctx context.Context, req *pb.CreateLaptopRequest) (*pb.CreateLaptopResponse, error) {
  		// ...
    
    	// 新增此行代码测试
    	time.Sleep(3 * time.Second)
  
  		err := server.Store.Save(laptop)
  }
  ```

- 再次运行 server 和 client，发现在客户端超时退出后，相应数据还是写入到了内存中，这是不应该的

  ```
  > make server
  $ 2021/06/19 00:10:41 start server on port: 8080
  $ 2021/06/19 00:10:43 receive a create-laptop request with id:7154eb98-2459-495e-911d-9e70034094a2.
  $ 2021/06/19 00:10:46 store save success 7154eb98-2459-495e-911d-9e70034094a2.
  
  -----
  
  > make client
  $ 2021/06/19 00:10:43 dial server: 0.0.0.0:8080
  $ 2021/06/19 00:10:44 laptop create error: rpc error: code = DeadlineExceeded desc = context deadline exceeded
  ```

- 修复超时错误，在实际保存数据时，对上下文进行判断

  ```go
  // service/laptop_server.go
  func (server *LaptopServer) CreateLaptop(ctx context.Context, req *pb.CreateLaptopRequest) (*pb.CreateLaptopResponse, error) {
  		// ...
    
    	// 新增此行代码测试
    	time.Sleep(3 * time.Second)
  
      if ctx.Err() == context.DeadlineExceeded {
        log.Print("deadline is exceeded")
        return nil, fmt.Errorf("deadline is exceeded")
      }
  		err := server.Store.Save(laptop)
  }
  ```

  

- 再次运行

  ```
  > make server
  $ 2021/06/19 00:22:52 receive a create-laptop request with id:f39f76d9-414e-47e3-b38c-efe5a8cea5b5.
  $ 2021/06/19 00:22:55 deadline is exceeded
  
  -----
  
  > make client
  $ 2021/06/19 00:22:52 dial server: 0.0.0.0:8080
  $ 2021/06/19 00:22:53 laptop create error: rpc error: code = DeadlineExceeded desc = context deadline exceeded
  ```

  

- 同样，客户端中断执行，在超时前执行 `ctrl+c`，数据同样被写入，修复问题

  ```go
  // service/laptop_server.go
  func (server *LaptopServer) CreateLaptop(ctx context.Context, req *pb.CreateLaptopRequest) (*pb.CreateLaptopResponse, error) {
  		// ...
    
    	// 新增此行代码测试
    	time.Sleep(3 * time.Second)
  
      if ctx.Err() == context.Canceled {
        log.Print("context is canceled")
        return nil, fmt.Errorf("context is canceled")
      }
      if ctx.Err() == context.DeadlineExceeded {
        log.Print("deadline is exceeded")
        return nil, fmt.Errorf("deadline is exceeded")
      }
    
  		err := server.Store.Save(laptop)
  }
  ```

**--本节结束--**

