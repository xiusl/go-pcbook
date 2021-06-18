package service

import (
	"errors"
	"fmt"
	"sync"

	"github.com/jinzhu/copier"
	"github.com/xiusl/pcbook/pb"
)

// ErrAlreadyExists 错误：对象已经存在
var ErrAlreadyExists = errors.New("record already exists")

// LaptopStore 存储 laptop 的接口
type LaptopStore interface {
	Save(laptop *pb.Laptop) error
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
	return nil
}
