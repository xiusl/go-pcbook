package service

import (
    "fmt"

    "golang.org/x/crypto/bcrypt"
)

// User 用户基本信息
type User struct {
    Username       string
    HashedPassword string
    Role           string
}

// NewUser 创建一个新的用户
func NewUser(username, password, role string) (*User, error) {
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return nil, fmt.Errorf("cannot hashed password, %w", err)
    }

    user := &User{
        Username:       username,
        HashedPassword: string(hashedPassword),
        Role:           role,
    }

    return user, nil
}

// IsCorrentPassword 验证密码是否正确
func (user *User) IsCorrentPassword(password string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(password))
    return err == nil
}

// Clone 返回一个克隆的 User
func (user *User) Clone() *User {
    return &User{
        Username:       user.Username,
        HashedPassword: user.HashedPassword,
        Role:           user.Role,
    }
}
