package sample

import (
    "github.com/golang/protobuf/ptypes"
    "github.com/xiusl/pcbook/pb"
)

// MewKeyboard 创建一个键盘
func MewKeyboard() *pb.Keyboard {
    return &pb.Keyboard{
        Layout:  randomKeyboardLayout(),
        Backlit: randomBool(),
    }
}

// NewCPU 创建一个CPU
func NewCPU() *pb.CPU {
    brand := randomCPUBrand()
    name := randomCUPName(brand)

    numberCores := randomInt(4, 8)
    numberThreads := randomInt(numberCores, 16)

    minGhz := randomFloat(2.0, 3.5)
    maxGhz := randomFloat(minGhz, 5.0)

    return &pb.CPU{
        Brand:         brand,
        Name:          name,
        NumberCores:   uint32(numberCores),
        NumberThreads: uint32(numberThreads),
        MinGhz:        minGhz,
        MaxGhz:        maxGhz,
    }
}

// NewGPU 创建一个GPU
func NewGPU() *pb.GPU {
    brand := randomGPUBrand()
    name := randomGPUName(brand)

    minGhz := randomFloat(1.0, 2.5)
    maxGhz := randomFloat(minGhz, 5.0)

    memory := &pb.Memory{
        Value: uint64(randomInt(4, 16)),
        Unit:  pb.Memory_GIGABYTE,
    }

    return &pb.GPU{
        Brand:  brand,
        Name:   name,
        MinGhz: minGhz,
        MaxGhz: maxGhz,
        Memory: memory,
    }
}

// NewRAM 创建一个内存 8-32G
func NewRAM() *pb.Memory {
    return &pb.Memory{
        Value: uint64(randomInt(8, 32)),
        Unit:  pb.Memory_GIGABYTE,
    }
}

// NewSSD 创建一个固态硬盘 128G-1T
func NewSSD() *pb.Storage {
    return &pb.Storage{
        Driver: pb.Storage_SDD,
        Memory: &pb.Memory{
            Value: uint64(randomInt(128, 1024)),
            Unit:  pb.Memory_GIGABYTE,
        },
    }
}

// NewSSD 创建一个机械硬盘 1T-2T
func NewHHD() *pb.Storage {
    return &pb.Storage{
        Driver: pb.Storage_SDD,
        Memory: &pb.Memory{
            Value: uint64(randomInt(1024, 2048)),
            Unit:  pb.Memory_GIGABYTE,
        },
    }
}

// NewScreen 创建一个屏幕
func NewScreen() *pb.Screen {
    return &pb.Screen{
        SizeInch:   randomFloat32(17, 25),
        Resolution: randomScreenResolution(),
        Panel:      randomScreenPanel(),
        Multitouch: randomBool(),
    }
}

// NewLaptop 创建一个
func NewLaptop() *pb.Laptop {
    brand := randomLaptopBrand()
    name := randomLaptopName(brand)
    return &pb.Laptop{
        Id:       randomID(),
        Brand:    brand,
        Name:     name,
        Cpu:      NewCPU(),
        Ram:      NewRAM(),
        Gpus:     []*pb.GPU{NewGPU()},
        Storages: []*pb.Storage{NewSSD(), NewHHD()},
        Screen:   NewScreen(),
        Keyboard: MewKeyboard(),
        Weight: &pb.Laptop_WeightKg{
            WeightKg: randomFloat(1, 3),
        },
        PriceUsd:    randomFloat(800, 2000),
        ReleaseYear: uint32(randomInt(2015, 2019)),
        UpdatedYear: ptypes.TimestampNow(),
    }
}

func RandomLaptopScore() float64 {
    return float64(randomInt(1, 10))
}
