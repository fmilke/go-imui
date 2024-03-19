package color

import (
    . "dyiui/internal/types"
)

type Color = uint32

func RGBA(r, g, b, a uint8) Color {
    return uint32(r) << 24 |
        uint32(g) << 16 |
        uint32(b) << 8 |
        uint32(a)
}

func ColorToGlVec4(c Color) [4]Float {
    r := float32(c >> 24) / 255.0
    g := float32((c >> 16) & 0xff) / 255.0
    b := float32((c >> 8) & 0xff) / 255.0
    a := float32(c & 0xff) / 255.0

    return [4]float32 { r, g, b, a }
}

