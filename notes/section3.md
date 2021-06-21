# 流式 RPC 上传文件

## 需要做的

- 定义一个包含流式请求的 RPC 服务调用，用以上传指定便携计算机的图片
- 实现这个调用的服务端代码，保存图片到本地指定目录
- 实现客户端代码，上传图片
- 完善超时控制，完成单元测试

## 定义服务调用

- 修改 `proto/laptop_service.proto`

  ```protobuf
  // proto/laptop_service.proto
  
  message UploadImageRequest {
      oneof data {
          ImageInfo info = 1;
          bytes chunk_data = 2;
      }
  }
  
  message ImageInfo {
      string laptop_id = 1;
      string image_type = 2;
  }
  
  message UploadImageResponse {
      string id = 1;
      uint32 size = 2;
  }
  
  service LaptopServices {
  	//...
  	rpc UploadImage(stream UploadImageRequest) returns (UploadImageResponse) {};
  }
  ```

  

- 生成 RPC 代码，`pb/laptop_service.pb.go` 中代码将会更新

  ```
  > make gen
  ```



## 实现服务端代码

- 在 `service` 目录中新建 `image_store.go` 提供保存图片接口，并实现本地磁盘存储

  ```go
  type ImageStore interface {
      Save(laptopID string, imageType string, imageData bytes.Buffer) (string, error)
  }
  
  type DiskImageStore struct {
      mutex       sync.RWMutex
      imageFolder string
      images      map[string]*ImageInfo
  }
  
  type ImageInfo struct {
      LaptopID string
      Type     string
      Path     string
  }
  
  func NewDiskImageStore(imageFolder string) *DiskImageStore {
  	// ...
  }
  
  func (store *DiskImageStore) Save(laptopID string, imageType string, imageData bytes.Buffer) (string, error) {
      imageID, err := uuid.NewRandom()
      // ...
      return imageID.String(), nil
  }
  ```

  

- 在初始化服务时，需要初始化本地图片存储

  ```go
  // service/laptop_server.go
  type LaptopServer struct {
      laptopStore LaptopStore
      imageStore  ImageStore
  }
  
  func NewLaptopServer(laptopStore LaptopStore, imageStore ImageStore) *LaptopServer {
      return &LaptopServer{
          laptopStore: laptopStore,
          imageStore:  imageStore,
      }
  }
  ```

  

- 实现服务端的上传图片功能

  ```go
  // service/laptop_server.go
  func (server *LaptopServer) UploadImage(stream pb.LaptopServices_UploadImageServer) error {
      req, err := stream.Recv()
      // ...
      return nil
  }
  ```



## 客户端代码

- 修改客户端代码，封装创建 `laptop` 方法

  ```go
  // cmd/client/main.go
  func createLaptop(laptopClient pb.LaptopServicesClient, laptop *pb.Laptop) {
      // ...
  }
  ```

  

- 修改客户端代码，封装搜索 `laptop` 方法

  ```go
  // cmd/client/main.go
  func searchTaplop(laptopClient pb.LaptopServicesClient, filter *pb.Filter) {
  	// ...
  }
  ```

  

- 针对创建和搜索的测试代码

  ```go
  // cmd/client/main.go
  func testCreateLaptop(laptopClient pb.LaptopServicesClient) {
      createLaptop(laptopClient, sample.NewLaptop())
  }
  
  func testSearchLaptop(laptopClient pb.LaptopServicesClient) {
      for i := 0; i < 10; i++ {
          createLaptop(laptopClient, sample.NewLaptop())
      }
  
      filter := &pb.Filter{
          MaxPriceUsd: 1000,
          MinCpuCores: 4,
          MinCpuGhz:   2.0,
          MinRam:      &pb.Memory{Value: 6, Unit: pb.Memory_GIGABYTE},
      }
  
      searchTaplop(laptopClient, filter)
  }
  ```

  

- 封装上传图片的方法

  ```go
  // cmd/client/main.go
  func uploadImage(laptopClient pb.LaptopServicesClient, laptopID string, imagePath string) {
      file, err := os.Open(imagePath)
  	// ...
      log.Printf("image uploaded with id: %s, size: %d", res.GetId(), res.GetSize())
  }
  ```

  

- 测试图片上传

  ```go
  // cmd/client/main.go
  func testUploadImage(laptopClient pb.LaptopServicesClient) {
      laptop := sample.NewLaptop()
      createLaptop(laptopClient, laptop)
      uploadImage(laptopClient, laptop.GetId(), "tmp/pc.png")
  }
  ```

  

- 主函数调用

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
  
      testUploadImage(laptopClient)
  }
  ```

  

- 开启服务，执行客户端程序，将在 img 目录下出现图片



## 超时和测试

- 这里同样存在服务端超时和客户端取消，在对数据进行读取和写入前要对上下文判断

  ```go
  func (server *LaptopServer) UploadImage(stream pb.LaptopServices_UploadImageServer) error {
      // ...
      for {
          log.Print("wait to receive more data")
          if stream.Context().Err() == context.Canceled {
              log.Print("context is canceled")
              return fmt.Errorf("context is canceled")
          }
  
          if stream.Context().Err() == context.DeadlineExceeded {
              log.Print("deadline is exceeded")
              return fmt.Errorf("deadline is exceeded")
          }
  
          req, err := stream.Recv()
          // ...
      }
      // ...
  }       
  ```

  

- 测试同样在 `service/laptop_client_test.go` 中进行

  ```go
  func TestUploadImage(t *testing.T) {
      // 初始化
      testImageFolder := "../tmp"
  
      laptopStore := service.NewInMemoryLaptopStore()
      imageStore := service.NewDiskImageStore(testImageFolder)
  
      laptop := sample.NewLaptop()
      err := laptopStore.Save(laptop)
      require.NoError(t, err)
  
      serverAddr := startTestLaptopServer(t, laptopStore, imageStore)
      laptopClient := newTestLaptopClient(t, serverAddr)
  
      imagePath := fmt.Sprintf("%s/pc.png", testImageFolder)
      file, err := os.Open(imagePath)
      require.NoError(t, err)
      defer file.Close()
      
      // 基本与客户端中的代码一致（cmd/client/main.go）
      // ...
      
      // 对响应进行验证
      res, err := stream.CloseAndRecv()
      require.NoError(t, err)
      require.NotZero(t, res.Id)
      require.EqualValues(t, res.GetSize(), size)
  
      saveImagePath := fmt.Sprintf("%s/%s%s", testImageFolder, res.GetId(), imageType)
      require.FileExists(t, saveImagePath)
      require.NoError(t, os.Remove(saveImagePath))
  }
  ```

**--本节结束--**