package sample

import (
    "math/rand"
    "time"

    "github.com/google/uuid"
    "github.com/xiusl/pcbook/pb"
)

func init() {
    rand.Seed(time.Now().UnixNano())
}

func randomKeyboardLayout() pb.Keyboard_Layout {
    switch rand.Intn(3) {
    case 1:
        return pb.Keyboard_AZERTY
    case 2:
        return pb.Keyboard_QWERTY
    default:
        return pb.Keyboard_UNKNOWN
    }
}

func randomCPUBrand() string {
    return randomStringFromSet("Intel", "AMD")
}

func randomCUPName(brand string) string {
    if brand == "Intel" {
        return randomStringFromSet(
            "Xeon W-3175X",
            "Core i9-10980XE",
            "Core i7-11700K",
            "Core i5-10600K",
            "Core i3-9300",
        )
    }
    return randomStringFromSet(
        "Ryzen 9 5950X",
        "Ryzen 7 5800X",
        "Ryzen 5 5600X",
        "Ryzen 3 3300X",
    )
}

func randomGPUBrand() string {
    return randomStringFromSet("AMD", "INVIDA")
}

func randomGPUName(brand string) string {
    if brand == "AMD" {
        return randomStringFromSet(
            "RX 6900 XT",
            "RX 5700 XT",
            "RX Vega 64",
            "RX 590",
            "R9 390X",
        )
    }
    return randomStringFromSet(
        "RTX 3090",
        "Titan RTX",
        "GTX 1080 Ti",
        "GTX 1070",
        "GTX 1650 Super",
    )
}

func randomScreenResolution() *pb.Screen_Resolution {
    height := randomInt(1080, 4030)
    width := height * 16 / 9
    return &pb.Screen_Resolution{
        Width:  uint32(width),
        Height: uint32(height),
    }
}

func randomScreenPanel() pb.Screen_Panel {
    if rand.Intn(2) == 1 {
        return pb.Screen_IPS
    }
    return pb.Screen_OLED
}

func randomID() string {
    return uuid.New().String()
}

func randomLaptopBrand() string {
    return randomStringFromSet(
        "Lenovo",
        "Dell",
        "Apple",
    )
}
func randomLaptopName(brand string) string {
    if brand == "Lenovo" {
        return randomStringFromSet("ThinkPad E14 2021", "YOGA 14s 2021", "Legion R9000P 2021H")
    }
    if brand == "Dell" {
        return randomStringFromSet("Inspiron 14-5000", "Vostro 14-3400", "Latitude 15 5000")
    }
    return randomStringFromSet("MacBook Pro M1", "MacBook Pro 2020", "MacBook Air")
}

func randomStringFromSet(a ...string) string {
    n := len(a)
    if n == 0 {
        return ""
    }
    return a[rand.Intn(n)]
}

func randomBool() bool {
    return rand.Intn(2) == 1
}

func randomInt(min, max int) int {
    return min + rand.Intn(max-min+1)
}

func randomInt32(min, max int32) int32 {
    return min + rand.Int31n(max-min+1)
}

func randomFloat(min, max float64) float64 {
    return min + rand.Float64()*(max-min)
}

func randomFloat32(min, max float32) float32 {
    return min + rand.Float32()*(max-min)
}
