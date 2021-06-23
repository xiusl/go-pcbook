## RPC 反射

## 如何使用反射

[server-reflection-tutorial](https://github.com/grpc/grpc-go/blob/master/Documentation/server-reflection-tutorial.md)

## 修改 `cmd/server/main.go`

```
// cmd/server/main.go
func main() {
    // ...

    pb.RegisterLaptopServicesServer(grpcServer, laptopServer)
    reflection.Register(grpcServer)

    // ...
}
```

## 安装 RPC cli 工具

- 这里使用的是 `ktr0731/evans`

```shell
> brew tap ktr0731/evans
> brew install evans
```

- 加载 evans

已经在服务点实现了反射，可以直接使用 `-r` 来加载 `evans`

```shell
> evans -r repl -p 8080
xiusl.pcbook.LaptopServices@127.0.0.1:8080>
```

- 常用命令

  - 查看所有的 `package`

    ```
    xiusl.pcbook.LaptopServices@127.0.0.1:8080>show package
    +-------------------------+
    |         PACKAGE         |
    +-------------------------+
    | grpc.reflection.v1alpha |
    | xiusl.pcbook            |
    +-------------------------+
    ```

  - 使用指定的package

    ```
    xiusl.pcbook.LaptopServices@127.0.0.1:8080> package xiusl.pcbook
    xiusl.pcbook@127.0.0.1:8080>
    ```

  - 查看当前package 下的服务

    ```
    xiusl.pcbook@127.0.0.1:8080> show service
    +----------------+--------------+---------------------+----------------------+
    |    SERVICE     |     RPC      |    REQUEST TYPE     |    RESPONSE TYPE     |
    +----------------+--------------+---------------------+----------------------+
    | LaptopServices | CreateLaptop | CreateLaptopRequest | CreateLaptopResponse |
    | LaptopServices | SearchLaptop | SearchLaptopRequest | SearchLaptopResponse |
    | LaptopServices | UploadImage  | UploadImageRequest  | UploadImageResponse  |
    | LaptopServices | RateLaptop   | RateLaptopRequest   | RateLaptopResponse   |
    +----------------+--------------+---------------------+----------------------+
    ```

  - 查看所有的消息

    ```
    xiusl.pcbook@127.0.0.1:8080> show message
    +----------------------+
    |       MESSAGE        |
    +----------------------+
    | CreateLaptopRequest  |
    | CreateLaptopResponse |
    | RateLaptopRequest    |
    | RateLaptopResponse   |
    | SearchLaptopRequest  |
    | SearchLaptopResponse |
    | UploadImageRequest   |
    | UploadImageResponse  |
    +----------------------+
    ```

  - 查看消息的结构描述

    ```
    xiusl.pcbook@127.0.0.1:8080> desc CreateLaptopRequest
    +--------+-----------------------+----------+
    | FIELD  |         TYPE          | REPEATED |
    +--------+-----------------------+----------+
    | laptop | TYPE_MESSAGE (Laptop) | false    |
    +--------+-----------------------+----------+
    ```

  - 调用创建便携电脑服务

    ```
    xiusl.pcbook.LaptopServices@127.0.0.1:8080> call CreateLaptop
    laptop::id (TYPE_STRING) =>
    laptop::brand (TYPE_STRING) => Apple
    laptop::name (TYPE_STRING) => Macbook pro
    laptop::cpu::brand (TYPE_STRING) => Intel
    laptop::cpu::name (TYPE_STRING) => i5
    laptop::cpu::number_cores (TYPE_UINT32) => 4
    laptop::cpu::number_threads (TYPE_UINT32) => 8
    laptop::cpu::min_ghz (TYPE_DOUBLE) => 2.5
    laptop::cpu::max_ghz (TYPE_DOUBLE) => 4.5
    ✔ dig down
    laptop::ram::value (TYPE_UINT64) => 8
    ✔ GIGABYTE
    <repeated> laptop::gpus::brand (TYPE_STRING) => AMD
    <repeated> laptop::gpus::name (TYPE_STRING) => Rx580
    <repeated> laptop::gpus::min_ghz (TYPE_DOUBLE) => 2
    <repeated> laptop::gpus::max_ghz (TYPE_DOUBLE) => 5
    <repeated> laptop::gpus::memory::value (TYPE_UINT64) => 8
    ✔ GIGABYTE
    <repeated> laptop::gpus::brand (TYPE_STRING) =>
    ✔ SDD
    <repeated> laptop::gpus::storages::memory::value (TYPE_UINT64) => 512
    ✔ GIGABYTE
    laptop::gpus::storages::screen::size_inch (TYPE_FLOAT) => 21
    laptop::gpus::storages::screen::resolution::width (TYPE_UINT32) => 3069
    laptop::gpus::storages::screen::resolution::height (TYPE_UINT32) => 2048
    ✔ OLED
    laptop::gpus::storages::screen::multitouch (TYPE_BOOL) => true
    ✔ QWERTY
    laptop::gpus::storages::keyboard::backlit (TYPE_BOOL) => true
    ✔ weight_kg
    laptop::gpus::storages::weight_kg (TYPE_DOUBLE) => 1.2
    laptop::gpus::storages::price_usd (TYPE_DOUBLE) => 200
    laptop::gpus::storages::release_year (TYPE_UINT32) => 2012
    laptop::gpus::storages::updated_year::seconds (TYPE_INT64) => 2020
    laptop::gpus::storages::updated_year::nanos (TYPE_INT32) => 12
    {
    "id": "e5d29114-3c64-4609-986d-ab0fae92a81e"
    }
    ```

  - 调用搜索便携电脑服务

    ```
    xiusl.pcbook.LaptopServices@127.0.0.1:8080> call SearchLaptop
    filter::max_price_usd (TYPE_DOUBLE) => 2000
    filter::min_cpu_cores (TYPE_UINT32) => 1
    filter::min_cpu_ghz (TYPE_DOUBLE) => 1
    filter::min_ram::value (TYPE_UINT64) => 1
    ✔ GIGABYTE
    {
        "laptop": {
        "id": "e5d29114-3c64-4609-986d-ab0fae92a81e",
        "brand": "Apple",
        "name": "Macbook pro",
        "cpu": {
            "brand": "Intel",
            "name": "i5",
            "numberCores": 4,
            "numberThreads": 8,
            "minGhz": 2.5,
            "maxGhz": 4.5
            },
        },
        // ...
    }
    ```

**--本节结束--**
