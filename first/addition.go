package first

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"

	"github.com/WiiLink24/MiiContestChannel/common"
)

type Root struct {
	Regions []Languages `json:"regions"`
}

type Languages struct {
	Japanese *Entry `json:"0"`
	English  *Entry `json:"1"`
	German   *Entry `json:"2"`
	French   *Entry `json:"3"`
	Spanish  *Entry `json:"4"`
	Italian  *Entry `json:"5"`
	Dutch    *Entry `json:"6"`
}

type Entry struct {
	Countries []Child `json:"countries"`
	Skills    []Child `json:"skills"`
}

type Child struct {
	Code uint32 `json:"code"`
	Name string `json:"name"`
}

type Addition struct {
	Header       AdditionHeader
	Countries    []CountryField
	Skills       []SkillField
	MarqueeField MarqueeField
}

type AdditionHeader struct {
	Type         [2]byte
	_            [6]byte
	CountryGroup uint32
	_            [12]byte
	Padding      [8]byte
}

type CountryField struct {
	Type        [2]byte
	FieldSize   uint16
	CountryCode uint32
	Text        [192]byte
}

type SkillField struct {
	Type      [2]byte
	FieldSize uint16
	SkillId   uint32
	Text      [96]byte
}

type MarqueeField struct {
	Tag         [2]byte
	SectionSize uint16
	Unknown     uint32
	Text        [1536]byte
}

//go:embed addition.json
var AdditionJson []byte

func (a Addition) ToBytes(any) []byte {
	buffer := new(bytes.Buffer)
	common.WriteBinary(buffer, a.Header)

	for _, country := range a.Countries {
		common.WriteBinary(buffer, country)
	}

	for _, skill := range a.Skills {
		common.WriteBinary(buffer, skill)
	}

	common.WriteBinary(buffer, a.MarqueeField)

	return buffer.Bytes()
}

func GetSupportedLanguagesForRegion(region uint32) []uint32 {
	switch region {
	case 2:
		return []uint32{1, 3, 4}
	case 3:
		return []uint32{1, 2, 3, 4, 5, 6}
	}

	return []uint32{0}
}

func (l *Languages) GetLanguageFromJSON(language uint32) *Entry {
	switch language {
	case 0:
		return l.Japanese
	case 1:
		return l.English
	case 2:
		return l.German
	case 3:
		return l.French
	case 4:
		return l.Spanish
	case 5:
		return l.Italian
	case 6:
		return l.Dutch
	}

	// Should never reach here
	return nil
}

func MakeAddition() error {
	var marqueeText []byte
	var baseText = []byte("WiiLink Mii Contest Channel!!!!")

	resp, error := http.Get(common.GetConfig().ServerURL + "/assets/marquee/marquee.txt")
	if error != nil {
		marqueeText = baseText
	} else {
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			marqueeText = baseText
		} else {
			marqueeText = body
		}
	}

	var actual [1536]byte
	copy(actual[:], marqueeText)

	var root Root
	err := json.Unmarshal(AdditionJson, &root)
	if err != nil {
		return err
	}

	for i := uint32(1); i < 4; i++ {
		for _, u := range GetSupportedLanguagesForRegion(i) {
			addition := Addition{
				Header: AdditionHeader{
					Type:         [2]byte{'A', 'D'},
					CountryGroup: 100*i + u,
					Padding:      [8]byte{math.MaxUint8, math.MaxUint8, math.MaxUint8, math.MaxUint8, math.MaxUint8, math.MaxUint8, math.MaxUint8, math.MaxUint8},
				},
				Countries: []CountryField{},
				Skills:    []SkillField{},
				MarqueeField: MarqueeField{
					Tag:         [2]byte{'N', 'W'},
					SectionSize: 1544,
					Unknown:     1,
					Text:        actual,
				},
			}

			currentLanguage := root.Regions[i-1].GetLanguageFromJSON(u)
			for _, country := range currentLanguage.Countries {
				var text [192]byte
				copy(text[:], country.Name)

				addition.Countries = append(addition.Countries, CountryField{
					Type:        [2]byte{'N', 'H'},
					FieldSize:   200,
					CountryCode: country.Code,
					Text:        text,
				})
			}

			for _, skill := range currentLanguage.Skills {
				var text [96]byte
				copy(text[:], skill.Name)

				addition.Skills = append(addition.Skills, SkillField{
					Type:      [2]byte{'N', 'J'},
					FieldSize: 104,
					SkillId:   skill.Code,
					Text:      text,
				})
			}

			err = common.Write(addition, fmt.Sprintf("addition/%d.ces", 100*i+u))
			if err != nil {
				return err
			}
		}
	}

	return nil
}
