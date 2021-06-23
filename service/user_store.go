package service

import "sync"

// UserStore 存储用户接口
type UserStore interface {
    Save(user *User) error
    Find(username string) (*User, error)
}

// InMemoryUserStore 在内存中存储用户
type InMemoryUserStore struct {
    mutex sync.RWMutex
    users map[string]*User
}

// NewInMemoryUserStore 新建一个用户内存存储实例
func NewInMemoryUserStore() *InMemoryUserStore {
    return &InMemoryUserStore{
        users: make(map[string]*User),
    }
}

// Save 存储用户到内存中
func (store *InMemoryUserStore) Save(user *User) error {
    store.mutex.Lock()
    defer store.mutex.Unlock()

    if store.users[user.Username] != nil {
        return ErrAlreadyExists
    }
    store.users[user.Username] = user.Clone()
    return nil
}

// Find 根据用户名在内存中查询用户
func (store *InMemoryUserStore) Find(username string) (*User, error) {
    store.mutex.Lock()
    defer store.mutex.Unlock()

    user := store.users[username]
    if user != nil {
        return user.Clone(), nil
    }
    return nil, nil
}
