package first

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"github.com/WiiLink24/MiiContestChannel/common"
	"math"
)

type Root struct {
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
var additionJson []byte

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

func MakeAddition() error {
	marqueeText := []byte("WiiLink Mii Contest Channel!!!!")
	var actual [1536]byte
	copy(actual[:], marqueeText)

	var root Root
	err := json.Unmarshal(additionJson, &root)
	if err != nil {
		return err
	}

	addition := Addition{
		Header: AdditionHeader{
			Type:         [2]byte{'A', 'D'},
			CountryGroup: 201,
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

	for _, country := range root.Countries {
		var text [192]byte
		copy(text[:], country.Name)

		addition.Countries = append(addition.Countries, CountryField{
			Type:        [2]byte{'N', 'H'},
			FieldSize:   200,
			CountryCode: country.Code,
			Text:        text,
		})
	}

	for _, skill := range root.Skills {
		var text [96]byte
		copy(text[:], skill.Name)

		addition.Skills = append(addition.Skills, SkillField{
			Type:      [2]byte{'N', 'J'},
			FieldSize: 104,
			SkillId:   skill.Code,
			Text:      text,
		})
	}

	return common.Write(addition, "addition/201.ces")
}
