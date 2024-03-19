package text

import (
	"bytes"
	"fmt"
	"os"

	"dyiui/internal/globals"
	. "dyiui/internal/units"

	"github.com/benoitkugler/textlayout/fonts/truetype"
	"github.com/benoitkugler/textlayout/harfbuzz"
	"github.com/danielgatis/go-findfont/findfont"
	"github.com/danielgatis/go-freetype/freetype"
)

type FontRepoEntry struct {
    FontFace *freetype.Face
    HbFont *harfbuzz.Font
    Ttf *truetype.Font
}

type FontRepo struct {
    entries []FontRepoEntry
}

func NewFontRepo() FontRepo {
    return FontRepo{}
}

func (r FontRepo) Get() *FontRepoEntry {
    if r.entries == nil || len(r.entries) == 0 {
        return nil
    }

    return &r.entries[0]
}

func LoadTTF(path string) (*truetype.Font, error) {
	f := GetSomeFont()

	file, err := os.Open(f)
	
	if err != nil {
		return nil, err
	}

	font, err := truetype.Parse(file)

	return font, err
}

func InitFace(px float32, path string) *freetype.Face {
	data, err := os.ReadFile(path)

	if err != nil {
		panic(err)
	}

	lib, err := freetype.NewLibrary()
	if err != nil {
		panic(err)
	}

	face, err := freetype.NewFace(lib, data, 0)
	if err != nil {
		panic(err)
	}

	pt := int(PxToPt(px))

	if globals.LOG_FONT {
		fmt.Printf("Retrieving font face of size %vpt\n", pt)
	}
	
	err = face.Pt(pt, int(PIXELS_PER_LOGICAL_INCH))
	if err != nil {
		panic(err)
	}

	return face
}

func (r *FontRepo) Load(path string) {
    face := InitFace(32, path)
	ttf, err := LoadTTF(path)
	if err != nil {
		panic(err)
	}

    hbFont := HBFont(ttf)

    r.Add(
        face,
        hbFont,
        ttf,
    )
}

func (r *FontRepo) Add(
    fontFace *freetype.Face,
    hbFont *harfbuzz.Font,
    ttf *truetype.Font,
) {
    r.entries = append(r.entries, FontRepoEntry {
        FontFace: fontFace,
        HbFont: hbFont,
        Ttf: ttf,
    })
}

type Fonts struct {
	font *[]*truetype.Font
}

func LoadFont(path string) *Fonts {
	f := Fonts {
		font: &[]*truetype.Font{},
	}

	data, err := os.ReadFile(path)

	if err != nil {
		panic(err)
	}

	reader := bytes.NewReader(data)
	ttf, err := truetype.Parse(reader)

	*f.font = append(*f.font, ttf)

	return &f
}

func HBFont(font *truetype.Font) *harfbuzz.Font {
	return harfbuzz.NewFont(font)
}

func GetSomeFont() string {
	fs, err := findfont.Find("Noto Sans", findfont.FontRegular)

	if err != nil {
		panic(err)
	}

	path := fs[0][2]
	return path
}
