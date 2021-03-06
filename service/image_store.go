package service

import (
    "bytes"
    "fmt"
    "os"
    "sync"

    "github.com/google/uuid"
)

// ImageStore 图像存储接口
type ImageStore interface {
    // Save 将便携计算机的图片存储下来
    Save(laptopID string, imageType string, imageData bytes.Buffer) (string, error)
}

// DiskImageStore 存储图像到硬盘，并在内存中保存图像的信息
type DiskImageStore struct {
    mutex       sync.RWMutex
    imageFolder string
    images      map[string]*ImageInfo
}

// ImageInfo 包含了便携计算机图像的一些信息
type ImageInfo struct {
    LaptopID string
    Type     string
    Path     string
}

// NewDiskImageStore 创建一个新的 DiskImageStore
func NewDiskImageStore(imageFolder string) *DiskImageStore {
    return &DiskImageStore{
        imageFolder: imageFolder,
        images:      make(map[string]*ImageInfo),
    }
}

// Save 存储图像
func (store *DiskImageStore) Save(laptopID string, imageType string, imageData bytes.Buffer) (string, error) {
    imageID, err := uuid.NewRandom()
    if err != nil {
        return "", fmt.Errorf("cannot generate image id %w", err)
    }

    // 拼接图片路径
    imagePath := fmt.Sprintf("%s/%s%s", store.imageFolder, imageID, imageType)

    // 创建图片的文件
    file, err := os.Create(imagePath)
    if err != nil {
        return "", fmt.Errorf("cannot create image file %w", err)
    }

    // 将图片的二进制数据写入文件中
    _, err = imageData.WriteTo(file)
    if err != nil {
        return "", fmt.Errorf("cannot write image date to file %w", err)
    }

    // 读写锁
    store.mutex.Lock()
    defer store.mutex.Unlock()

    // 更新内存中的图像信息
    store.images[imageID.String()] = &ImageInfo{
        LaptopID: laptopID,
        Type:     imageType,
        Path:     imagePath,
    }

    return imageID.String(), nil
}
