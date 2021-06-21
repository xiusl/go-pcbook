package service

import "sync"

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

    rating := store.rating[laptopID]
    if rating == nil {
        rating = &Rating{
            Count: 1,
            Sum:   score,
        }
    } else {
        rating.Count += 1
        rating.Sum += score
    }
    store.rating[laptopID] = rating
    return rating, nil
}
