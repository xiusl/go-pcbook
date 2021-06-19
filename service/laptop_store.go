package service

import (
    "context"
    "errors"
    "fmt"
    "log"
    "sync"

    "github.com/jinzhu/copier"
    "github.com/xiusl/pcbook/pb"
)

// ErrAlreadyExists 错误：对象已经存在
var ErrAlreadyExists = errors.New("record already exists")

// LaptopStore 存储 laptop 的接口
type LaptopStore interface {
    Save(laptop *pb.Laptop) error
    FindByID(id string) (*pb.Laptop, error)
    Search(ctx context.Context, filter *pb.Filter, found func(laptop *pb.Laptop) error) error
}

// InMemoryLaptopStore 内存存储
type InMemoryLaptopStore struct {
    mutex sync.Mutex
    data  map[string]*pb.Laptop
}

// NewInMemoryLaptopStore 创建一个内存存储
func NewInMemoryLaptopStore() *InMemoryLaptopStore {
    return &InMemoryLaptopStore{
        data: make(map[string]*pb.Laptop),
    }
}

// Save 内存存储，保存接口的具体实现
func (store *InMemoryLaptopStore) Save(laptop *pb.Laptop) error {
    store.mutex.Lock()
    defer store.mutex.Unlock()

    if store.data[laptop.Id] != nil {
        return ErrAlreadyExists
    }

    tmp := &pb.Laptop{}
    err := copier.Copy(tmp, laptop)
    if err != nil {
        return fmt.Errorf("connot copy laptop data: %w", err)
    }
    store.data[tmp.Id] = tmp

    log.Printf("store save success %s.\n", tmp.Id)
    return nil
}

// FindByID 根据 Id 获取 laptop
func (store *InMemoryLaptopStore) FindByID(id string) (*pb.Laptop, error) {
    store.mutex.Lock()
    defer store.mutex.Unlock()

    laptop := store.data[id]
    if laptop == nil {
        return nil, nil
    }

    tmp := &pb.Laptop{}
    err := copier.Copy(tmp, laptop)
    if err != nil {
        return nil, fmt.Errorf("connot copy laptop data: %w", err)
    }

    return tmp, nil
}

// Search 搜索指定的便携电脑
func (store *InMemoryLaptopStore) Search(ctx context.Context, filter *pb.Filter, found func(laptop *pb.Laptop) error) error {
    store.mutex.Lock()
    defer store.mutex.Unlock()

    for _, laptop := range store.data {

        if ctx.Err() == context.Canceled || ctx.Err() == context.DeadlineExceeded {
            log.Print("context is canceled")
            return errors.New("context is canceled")
        }

        log.Print("check laptop id ", laptop.Id)
        if isQualified(filter, laptop) {
            other, err := deepCopy(laptop)
            if err != nil {
                return err
            }
            err = found(other)
            if err != nil {
                return err
            }
        }
    }
    return nil
}

func isQualified(filter *pb.Filter, laptop *pb.Laptop) bool {
    if laptop.GetPriceUsd() > filter.MaxPriceUsd {
        return false
    }
    if laptop.GetCpu().GetNumberCores() < filter.MinCpuCores {
        return false
    }
    if laptop.GetCpu().GetMinGhz() < filter.MinCpuGhz {
        return false
    }
    if toBit(laptop.GetRam()) < toBit(filter.MinRam) {
        return false
    }
    return true
}

func toBit(memory *pb.Memory) uint64 {
    value := memory.GetValue()

    switch memory.GetUnit() {
    case pb.Memory_BIT:
        return value
    case pb.Memory_BYTE:
        return value << 3
    case pb.Memory_KILOBYTE:
        return value << 13
    case pb.Memory_MEGABYTE:
        return value << 23
    case pb.Memory_GIGABYTE:
        return value << 33
    case pb.Memory_TERABYTE:
        return value << 43
    default:
        return 0
    }
}

func deepCopy(laptop *pb.Laptop) (*pb.Laptop, error) {
    tmp := &pb.Laptop{}
    err := copier.Copy(tmp, laptop)
    if err != nil {
        return nil, fmt.Errorf("connot copy laptop data: %w", err)
    }
    return tmp, nil
}
