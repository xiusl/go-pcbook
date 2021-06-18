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

  

- a

- asas

- sda

- sada

- 