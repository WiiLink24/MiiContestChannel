package plaza

import (
	"bytes"
	_ "embed"
	"github.com/WiiLink24/MiiContestChannel/common"
	"math"
)

// MiiList is the structure that different types of inherit.
// It contains a list of Miis and their corresponding Artisans.
type MiiList struct {
	Header Header
	Miis   []MiiPair
}

// MiiPair is a Mii paired with it's Artisan
type MiiPair struct {
	Mii     Mii
	Artisan Artisan
}

type Header struct {
	Tag           common.ListTag
	_             uint16
	CountryRegion uint32
	ListNumber    uint32
	ErrorCode     uint32
	_             [12]byte
	Padding       [4]byte
	MiiTag        [2]byte
	MiiTagSize    uint16
	Unknown       uint32
	NumberOfMiis  uint32
}

type Mii struct {
	Tag         common.ListTag
	TagSize     uint16
	MiiIndex    uint32
	EntryNumber uint32
	MiiData     [76]byte
	_           uint16
	Popularity  uint8
	_           uint8
	Skill       uint16
	Initials    [2]byte
}

type Artisan struct {
	Tag                [2]byte
	TagSize            uint16
	ArtisanIndex       uint32
	ArtisanNumber      uint32
	MiiArtisanData     [76]byte
	_                  uint8
	IsMasterMiiArtisan uint8
	_                  [3]byte
	CountryCode        uint8
	_                  uint16
}

func (m MiiList) ToBytes(_ any) []byte {
	buffer := new(bytes.Buffer)
	common.WriteBinary(buffer, m.Header)

	for _, pair := range m.Miis {
		common.WriteBinary(buffer, pair.Mii)
		common.WriteBinary(buffer, pair.Artisan)
	}

	return buffer.Bytes()
}

func MakeList(listType common.ListTag, miis []common.MiiWithArtisan, filename string, number *uint32) error {
	var listNumber uint32
	if number != nil {
		listNumber = *number
	}

	list := MiiList{}

	header := Header{
		Tag:           listType,
		CountryRegion: 151,
		ListNumber:    listNumber,
		ErrorCode:     0,
		Padding:       [4]byte{math.MaxUint8, math.MaxUint8, math.MaxUint8, math.MaxUint8},
		MiiTag:        common.MiiPairNumber,
		MiiTagSize:    12,
		Unknown:       1,
		NumberOfMiis:  uint32(len(miis)),
	}

	var miiPair []MiiPair
	for i, mii := range miis {
		var tempMiiData [76]byte
		var tempArtisanMiiData [76]byte
		var tempInitials [2]byte

		copy(tempMiiData[:], mii.MiiData)
		copy(tempArtisanMiiData[:], mii.ArtisanMiiData)
		copy(tempInitials[:], mii.Initials)

		pair := MiiPair{
			Mii: Mii{
				Tag:         common.PairMii,
				TagSize:     96,
				MiiIndex:    uint32(i + 1),
				EntryNumber: mii.EntryNumber,
				MiiData:     tempMiiData,
				Popularity:  mii.Likes,
				Skill:       mii.Skill,
				Initials:    tempInitials,
			},
			Artisan: Artisan{
				Tag:                common.PairArtisan,
				TagSize:            96,
				ArtisanIndex:       uint32(i + 1),
				ArtisanNumber:      mii.ArtisanId,
				MiiArtisanData:     tempArtisanMiiData,
				IsMasterMiiArtisan: mii.IsMasterArtisan,
				CountryCode:        mii.CountryCode,
			},
		}

		miiPair = append(miiPair, pair)
	}

	list.Header = header
	list.Miis = miiPair

	if err := common.Write(list, "151/"+filename); err != nil {
		return err
	}

	list.Header.CountryRegion = 0

	return common.Write(list, "0/"+filename)
}
