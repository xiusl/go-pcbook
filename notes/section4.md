# 双向流式 RPC

## 需要做的
- 定义一个双向流式 RPC，去给便携电脑评分
- 实现服务端逻辑
- 实现客户端调用
- 测试

## 定义 PRC

- 在 `proto/laptop_service.proto` 中定义 proto

  ```protobuf
  message RateLaptopRequest {
      string laptop_id = 1;
      double score = 2;
  }

  message RateLaptopResponse {
      string laptop_id = 1;
      uint32 rated_count = 2;
      double average_scote = 3;
  }

  service LaptopServices {
      // ...
      rpc RateLaptop(stream RateLaptopRequest) returns (stream RateLaptopResponse) {};
  }

  ```



- 生成 go 代码

  ```shell
  make gen
  ```



## 评分的存储

- 在 service 中新建 `rating_store.go`

  ```go
  // RatingStore 分数存储接口
  type RatingStore interface {
      Add(laptopID string, score float64) (*Rating, error)
  }

  // Rating 分数对象
  type Rating struct {
      Count uint32
      Sum   float64
  }

  // InMemoryRatingStore 分数存储的内存实现
  type InMemoryRatingStore struct {
      mutex  sync.RWMutex
      rating map[string]*Rating
  }
  ```



- 实现存储接口的方法
  ```go
  // service/rating_store.go
  // NewInMemoryRatingStore 工厂方法
  func NewInMemoryRatingStore() *InMemoryRatingStore {
      return &InMemoryRatingStore{
          rating: make(map[string]*Rating),
      }
  }

  // Add 内存分数存储新增
  func (store *InMemoryRatingStore) Add(laptopID string, score float64) (*Rating, error) {
      store.mutex.Lock()
      defer store.mutex.Unlock()
      // ...
      return rating, nil
  }
  ```



## 完善 RPC 服务端代码

- `LaptopService` 对象中新增评分存储

  ```go
  // LaptopServer 提供 laptop 服务的服务器
  type LaptopServer struct {
      laptopStore LaptopStore
      imageStore  ImageStore
      ratingStore RatingStore
  }

  // NewLaptopServer 创建一个 laptop 服务器
  func NewLaptopServer(laptopStore LaptopStore, imageStore ImageStore, ratingStore RatingStore) *LaptopServer {
      return &LaptopServer{
          laptopStore: laptopStore,
          imageStore:  imageStore,
          ratingStore: ratingStore,
      }
  }
  ```



- 修复之前的一些测试代码

  ```go
  // service/laptop_client_test.go
  // 新增 ratingStore service.RatingStore
  func startTestLaptopServer(t *testing.T, laptopstroe service.LaptopStore, imageStore service.ImageStore, ratingStore service.RatingStore) string {
  	// ...
  }

  func TestUploadImage(t *testing.T) {
      // ...
      // 测试上传图像是不需要评分存储，可以传入 nil
      serverAddr := startTestLaptopServer(t, laptopStore, imageStore, nil)
      // ...
  }

  // 测试创建和测试搜索的函数中同样处理
  ```



- 修复服务端运行代码

  ```go
  // cmd/server/main.go
  func main() {
      // ...
      laptopStore := service.NewInMemoryLaptopStore()
      imageStore := service.NewDiskImageStore("img")
      ratingStore := service.NewInMemoryRatingStore()
      laptopServer := service.NewLaptopServer(laptopStore, imageStore, ratingStore)
      // ...
  }
  ```



- 在 `laptop_server.go` 中实现 RPC 调用

  ```go
  func (server *LaptopServer) RateLaptop(stream pb.LaptopServices_RateLaptopServer) error {
      // 循环读取内容
      for {
          // 对上下文进行判断
          // ...

          // 接收请求
          req, err := stream.Recv()

          // 查询便携电脑
          laptap, err := server.laptopStore.FindByID(laptopID)

  		// 打分
          rating, err := server.ratingStore.Add(laptopID, scroe)

          // 发送本次的响应
          err = stream.Send(res)

      }
      return nil
  }

  ```



## 实现客户端调用



- `cmd/client/main.go` 中新增 `ratingLaptop` 函数

  ```go
  func ratingLaptop(laptopClient pb.LaptopServicesClient, laptopIDS []string, scores []float64) error {
  	// 建立上下文
      ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)

      // 建立流式请求
      stream, err := laptopClient.RateLaptop(ctx)

      // 启动协程，不断读取响应
      withResp := make(chan error)
      go for() {
          for {
              res, err := stream.Recv()
              // ...
          }
      }()

      // 依次发送多个评分请求
      for i, laptopID := range laptopIDS {
          err = stream.Send(req)
      }

      // 关闭发送
      err = stream.CloseSend()

      // 返回
      err = <-withResp
      return err
  }
  ```



- 对 `ratingLaptop` 进行测试

  ```go
  func testRatingLaptop(laptopClient pb.LaptopServicesClient) {
      n := 3
      laptopIDS := make([]string, n)

      for i := 0; i < n; i++ {
          laptop := sample.NewLaptop()
          createLaptop(laptopClient, laptop)
          laptopIDS[i] = laptop.GetId()
      }

      scores := make([]float64, n)
      for {
          fmt.Println("rate laptop (y/n)?:")
          var ans string
          fmt.Scan(&ans)

          if strings.ToLower(ans) != "y" {
              break
          }

          for i := 0; i < n; i++ {
              scores[i] = sample.RandomLaptopScore()
          }

          err := ratingLaptop(laptopClient, laptopIDS, scores)
          if err != nil {
              log.Fatal(err)
          }
      }
  }
  ```



-



## 完成测试

- 同样在 `laptop_client_test.go` 中进行测试

  ```go
  func TestRatingLaptop(t *testing.T) {
      laptopStore := service.NewInMemoryLaptopStore()
      ratingStore := service.NewInMemoryRatingStore()

      laptop := sample.NewLaptop()
      err := laptopStore.Save(laptop)
      require.NoError(t, err)

      serverAddr := startTestLaptopServer(t, laptopStore, nil, ratingStore)
      laptopClient := newTestLaptopClient(t, serverAddr)

      stream, err := laptopClient.RateLaptop(context.Background())
      require.NoError(t, err)

      scores := []float64{8, 7.5, 10}
      averages := []float64{8, 7.75, 8.5}

      n := len(scores)
      for i := 0; i < n; i++ {
          req := &pb.RateLaptopRequest{
              LaptopId: laptop.Id,
              Score:    scores[i],
          }

          err = stream.Send(req)
          require.NoError(t, err)
      }

      err = stream.CloseSend()
      require.NoError(t, err)

      for idx := 0; ; idx++ {
          res, err := stream.Recv()
          if err == io.EOF {
              require.Equal(t, n, idx)
              return
          }

          require.NoError(t, err)
          require.Equal(t, laptop.GetId(), res.GetLaptopId())
          require.Equal(t, uint32(idx+1), res.GetRatedCount())
          require.Equal(t, averages[idx], res.GetAverageScote())
      }
  }
  ```

**--本节结束--**
