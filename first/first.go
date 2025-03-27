package first

import (
	"fmt"
	"github.com/WiiLink24/MiiContestChannel/common"
	"math"
)

type First struct {
	Type             common.ListTag
	DiscontinuedFlag uint8
	_                [4]byte
	CountryCode      uint8
	_                [16]byte
	Padding          [8]byte
	Tag              common.ListTag
	TagSize          uint16
	Unk              uint32
	CountryGroup     uint8
	/*
		3rd lowest bits are unknown
		4th bit: Marquee Text
		3rd bit: Re-enable initials
		2nd bit: Disable initials
		1st bit: Show Mii Artisan
		0th bit: Show number of Miis in Posting Plaza
	*/
	FirstBitField uint8
	_             uint8
	/*
		The below toggle the languages available for the message board service.

		7th bit: Unknown
		6th bit: Dutch
		5th bit: Italian
		4th bit: Spanish
		3rd bit: French
		2nd bit: German
		1st bit: English
		0th bit: Japanese
	*/
	SecondBitField common.LanguageFlag
}

func (f First) ToBytes(data any) []byte {
	return common.ToBytes(data)
}

func GetSupportedLanguages(countryCode uint8) common.LanguageFlag {
	if 8 <= countryCode && countryCode <= 52 {
		return common.English | common.French | common.Spanish
	} else if 64 <= countryCode && countryCode <= 110 {
		return common.English | common.German | common.French | common.Spanish | common.Italian | common.Dutch
	}

	return common.Japanese | common.English | common.German | common.French | common.Spanish | common.Italian | common.Dutch
}

func MakeFirst() error {
	for _, code := range common.CountryCodes {
		first := First{
			Type:             common.FirstList,
			DiscontinuedFlag: 0,
			CountryCode:      code,
			Padding:          [8]byte{math.MaxUint8, math.MaxUint8, math.MaxUint8, math.MaxUint8, math.MaxUint8, math.MaxUint8, math.MaxUint8, math.MaxUint8},
			Tag:              common.FirstList,
			TagSize:          12,
			Unk:              1,
			CountryGroup:     151,
			FirstBitField:    19,
			SecondBitField:   GetSupportedLanguages(code),
		}

		err := common.Write(first, fmt.Sprintf("first/%d.ces", code))
		if err != nil {
			return err
		}
	}

	return nil
}
