# 实现服务器流式 RPC

## 需要做的

- 定义一个流式 RPC 文件，去搜索指定的便携电脑
- 实现一个服务去处理搜索 RPC 请求
- 在客户端中去调用这个搜索流 RPC 请求
- 写一些单元测试



## 定义 RPC

- 在 proto 中新建 `filter_message.proto`  文件，定义过滤器的一些信息

  ```protobuf
  message Filter {
      double max_price_usd = 1;
      uint32 min_cpu_cores = 2;
      double min_cpu_ghz = 3;
  		Memory min_ram = 4;
  }
  ```

  

- 在 `laptop_server.proto` 定义搜索

  ```protobuf
  message SearchLaptopRequest { Filter filter = 1; }
  message SearchLaptopResponse { Laptop laptop = 1; }
  
  service LaptopServices {
      // ...
      rpc SearchLaptop(SearchLaptopRequest) returns (stream SearchLaptopResponse) {}
  }
  ```

  

- 生成 proto 代码

  ```
  make gen
  ```



## 实现搜索服务

- 在 `laptop_store.go` 新增搜索接口

  ```go
  type LaptopStore interface {
      // ...
      Search(filter *pb.Filter, found func(laptop *pb.Laptop) error) error
  }
  ```

  

- 增加对 `LaptopStore` 的内存实现

  ```go
  // laptop_store.go
  
  func (store *InMemoryLaptopStore) Search(filter *pb.Filter, found func(laptop *pb.Laptop) error) error {
      // ...
      
  }
  
  // 提供一些辅助函数
  func isQualified(filter *pb.Filter, laptop *pb.Laptop) bool {
      // ...
  }
  func toBit(memory *pb.Memory) uint64 {
      // ...
  }
  ```

  

- 在 `laptop_server.go` 实现 `SearchLaptop`

  ```go
  func (server *LaptopServer) SearchLaptop(req *pb.SearchLaptopRequest, stream pb.LaptopServices_SearchLaptopServer) error {
      // ...
  }
  ```



## 在客户端调用搜索服务

- 重构客户端代码，封装创建方法

  ```go
  // cmd/client/main.go
  
  func createLaptop(laptopClient pb.LaptopServicesClient) {
      laptop := sample.NewLaptop()
      req := &pb.CreateLaptopRequest{
          Laptop: laptop,
      }
      // ...
      log.Printf("created laptop success, id: %v", res.Id)
  }
  ```

  

- 在主函数中创建多个 laptop

  ```go
  func main() {
  	// ...
  
  	for i := 0; i < 10; i++ {
  		createLaptop(laptopClient)
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
  
  
  
- 新增搜索函数

    ```go
    func searchTaplop(laptopClient pb.LaptopServicesClient, filter *pb.Filter) {
    	log.Printf("search filter: %v", filter)
    
    	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    	defer cancel()
    
    	req := &pb.SearchLaptopRequest{
    		Filter: filter,
    	}
    	stream, err := laptopClient.SearchLaptop(ctx, req)
    	if err != nil {
    		log.Fatalf("cannot search laptop: %v", err)
    	}
    
    	for {
    		res, err := stream.Recv()
    		if err == io.EOF {
    			return
    		}
    		if err != nil {
    			log.Fatalf("cannot receive response: %v", err)
    		}
    
    		laptop := res.GetLaptop()
    		log.Print("- found: ", laptop.GetId())
    		log.Print("  + brand: ", laptop.GetBrand())
    		log.Print("  + name: ", laptop.GetName())
    		log.Print("  + cpu cores: ", laptop.GetCpu().GetNumberCores())
    		log.Print("  + cpu min ghz: ", laptop.GetCpu().GetMinGhz())
    		log.Print("  + ram: ", laptop.GetRam().GetValue(), laptop.GetRam().GetUnit())
    		log.Print("  + price: ", laptop.GetPriceUsd())
    
    	}
    }
    ```

    

## 运行

-   服务端

```shell
> make server
$ go run cmd/server/main.go -port=8080
$ 2021/06/19 10:28:19 start server on port: 8080
$ 2021/06/19 10:28:25 receive a create-laptop request with id:f4581124-7044-4710-882d-98d75b557b4b.
$ 2021/06/19 10:28:25 store save success f4581124-7044-4710-882d-98d75b557b4b.
$ 2021/06/19 10:28:25 receive a create-laptop request with id:fa5ccc8a-9097-4bed-837a-d6e7dd049c1d.
$ 2021/06/19 10:28:25 store save success fa5ccc8a-9097-4bed-837a-d6e7dd049c1d.
$ 2021/06/19 10:28:25 receive a create-laptop request with id:75939d03-b6bd-4e06-93a9-c173d1c0f872.
$ 2021/06/19 10:28:25 store save success 75939d03-b6bd-4e06-93a9-c173d1c0f872.
$ 2021/06/19 10:28:25 receive a create-laptop request with id:d1ce927b-b77c-4699-b458-dd4965deb367.
$ 2021/06/19 10:28:25 store save success d1ce927b-b77c-4699-b458-dd4965deb367.
```

-   客户端

```shell
> make client
$ go run cmd/client/main.go -addr="0.0.0.0:8080"
$ 2021/06/19 10:28:25 dial server: 0.0.0.0:8080
$ 2021/06/19 10:28:25 created laptop success, id: f4581124-7044-4710-882d-98d75b557b4b
$ 2021/06/19 10:28:25 created laptop success, id: fa5ccc8a-9097-4bed-837a-d6e7dd049c1d
$ 2021/06/19 10:28:25 created laptop success, id: 75939d03-b6bd-4e06-93a9-c173d1c0f872
$ 2021/06/19 10:28:25 created laptop success, id: d1ce927b-b77c-4699-b458-dd4965deb367
$ 2021/06/19 10:28:25 created laptop success, id: 6f0b4461-d9e3-4ee5-b2a6-376979f874af
$ 2021/06/19 10:28:25 created laptop success, id: 857926eb-caa9-453f-8fd1-ba10fdfec1eb
$ 2021/06/19 10:28:25 created laptop success, id: 4775c90f-30e3-462d-b6fa-65766ad0be3c
$ 2021/06/19 10:28:25 created laptop success, id: c7f514da-696a-43e2-822e-d559f266190f
$ 2021/06/19 10:28:25 created laptop success, id: 48c34443-683c-4de6-bc67-5a300a7bc131
$ 2021/06/19 10:28:25 created laptop success, id: 5c4bc4d5-d86b-4164-a7ed-69f19e44be80
$ 2021/06/19 10:28:25 search filter: max_price_usd:1000 min_cpu_cores:4 min_cpu_ghz:2 min_ram:{value:6 unit:GIGABYTE}
$ 2021/06/19 10:28:25 - found: 48c34443-683c-4de6-bc67-5a300a7bc131
$ 2021/06/19 10:28:25   + brand: Dell
$ 2021/06/19 10:28:25   + name: Inspiron 14-5000
$ 2021/06/19 10:28:25   + cpu cores: 5
$ 2021/06/19 10:28:25   + cpu min ghz: 3.042457815634881
$ 2021/06/19 10:28:25   + ram: 9 GIGABYTE
$ 2021/06/19 10:28:25   + price: 851.5690577317878
```



## 超时和取消

-   首先在搜索时增加一个耗时任务

    ```go
    // service/laptop_store.go
    func (store *InMemoryLaptopStore) Search(filter *pb.Filter, found func(laptop *pb.Laptop) error) error {
    	store.mutex.Lock()
    	defer store.mutex.Unlock()
    
    	for _, laptop := range store.data {
    
    		time.Sleep(4 * time.Second)
    		log.Print("check laptop id ", laptop.Id)
        	if isQualified(filter, laptop) {
            	// ...
            }
        }
        // ...
    }
    ```

    

- 此时，再次执行，发现客户端已经退出，但搜索任务还在执行，因此在执行比较时需要判断上下文状态

    ```go
    // service/laptop_store.go
    func (store *InMemoryLaptopStore) Search(ctx context.Context, filter *pb.Filter, found func(laptop *pb.Laptop) error) error {
    	store.mutex.Lock()
    	defer store.mutex.Unlock()
    
    	for _, laptop := range store.data {
    
    		time.Sleep(4 * time.Second)
            
            if ctx.Err() == context.Canceled || ctx.Err() == context.DeadlineExceeded {
    			log.Print("context is canceled")
    			return errors.New("context is canceled")
    		}
            
    		log.Print("check laptop id ", laptop.Id)
       
            if isQualified(filter, laptop) {
            	// ...
            }
        }
        // ...
    }
    ```

    ```go
    // service/laptop_server.go
    func (server *LaptopServer) SearchLaptop(req *pb.SearchLaptopRequest, stream pb.LaptopServices_SearchLaptopServer) error {
    	filter := req.GetFilter()
    	log.Printf("receive a search-laptop request with filter: %v", filter)
        // add stream.Context()
    	err := server.Store.Search(stream.Context(), filter, func(laptop *pb.Laptop) error {
        	// ...
        })
    }
    ```

    再次运行发现问题被修复



## 进行测试

-   可以通过模拟 gPRC Server 的方式进行测试，但是需要 mock 太多方法，因此还是通过测试服务器的方法来测试

```go
// service/laptop_client_test.go
func TestClientSearchLaptop(t *testing.T) {

	filter := &pb.Filter{
		MaxPriceUsd: 1000,
		MinCpuCores: 4,
		MinCpuGhz:   2.0,
		MinRam:      &pb.Memory{Value: 6, Unit: pb.Memory_GIGABYTE},
	}

	store := service.NewInMemoryLaptopStore()
	expectedIDs := make(map[string]bool)

    // 创建 6 个 laptop，其中前四个为不符合条件的情况，后两个为预期结果
	for i := 0; i < 6; i++ {
		laptop := sample.NewLaptop()
		switch i {
		case 0:
			laptop.PriceUsd = 1500
		case 1:
			laptop.Cpu.NumberCores = 1
		case 2:
			laptop.Cpu.MinGhz = 1.0
		case 3:
			laptop.Ram = &pb.Memory{Value: 1, Unit: pb.Memory_GIGABYTE}
		case 4:
			laptop.PriceUsd = 900
			laptop.Cpu.NumberCores = 8
			laptop.Cpu.MinGhz = 2.5
			laptop.Ram = &pb.Memory{Value: 8, Unit: pb.Memory_GIGABYTE}
			expectedIDs[laptop.Id] = true
		case 5:
			laptop.PriceUsd = 888
			laptop.Cpu.NumberCores = 16
			laptop.Cpu.MinGhz = 3.5
			laptop.Ram = &pb.Memory{Value: 16, Unit: pb.Memory_GIGABYTE}
			expectedIDs[laptop.Id] = true
		}

		err := store.Save(laptop)
		require.NoError(t, err)
	}

    // 构建测试服务器
	_, serverAddr := startTestLaptopServer(t, store)
	laptopClient := newTestLaptopClient(t, serverAddr)

	req := &pb.SearchLaptopRequest{
		Filter: filter,
	}
	stream, err := laptopClient.SearchLaptop(context.Background(), req)
	require.NoError(t, err)

	found := 0
	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}

		require.NoError(t, err)
		require.Contains(t, expectedIDs, res.GetLaptop().GetId())
		found += 1
	}
	require.Equal(t, found, len(expectedIDs))
}
```

**--本节结束--**