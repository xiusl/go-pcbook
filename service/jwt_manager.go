package service

import (
    "fmt"
    "time"

    "github.com/dgrijalva/jwt-go"
)

// JWTManager jwt 管理类
type JWTManager struct {
    secretKey     string
    tokenDuration time.Duration
}

// UserClaims 包含一些用户信息的 jwt claims
type UserClaims struct {
    jwt.StandardClaims
    Username string `json:"username"`
    Role     string `json:"role"`
}

// NewJWTManager 新建一个 JWT 管理对象
func NewJWTManager(secretKey string, duration time.Duration) *JWTManager {
    return &JWTManager{
        secretKey:     secretKey,
        tokenDuration: duration,
    }
}

// Generate 根据用户信息生成 jwt token
func (manager *JWTManager) Generate(user *User) (string, error) {
    claims := UserClaims{
        StandardClaims: jwt.StandardClaims{
            ExpiresAt: time.Now().Add(manager.tokenDuration).Unix(),
        },
        Username: user.Username,
        Role:     user.Role,
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(manager.secretKey))
}

// Verify 验证 token 字符串，如果有效将返回用户 claims
func (manager *JWTManager) Verify(tokenString string) (*UserClaims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(t *jwt.Token) (interface{}, error) {
        _, ok := t.Method.(*jwt.SigningMethodHMAC)
        if !ok {
            return nil, fmt.Errorf("unexpected token signing method")
        }

        return []byte(manager.secretKey), nil
    })

    if err != nil {
        return nil, fmt.Errorf("invalid token: %w", err)
    }

    claims, ok := token.Claims.(*UserClaims)
    if !ok {
        return nil, fmt.Errorf("invalid token claims")
    }

    return claims, nil
}
