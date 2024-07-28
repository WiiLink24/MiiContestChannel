package plaza

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"github.com/SketchMaster2001/libwc24crypt"
	"github.com/WiiLink24/MiiContestChannel/common"
	"github.com/jackc/pgx/v4/pgxpool"
	"os"
	"strconv"
	"time"
)

var (
	iv = []byte{70, 70, 20, 40, 143, 110, 36, 6, 184, 107, 135, 239, 96, 45, 80, 151}
)

type WiiSportsResort struct {
	OpenTimestamp  uint32
	CloseTimestamp uint32
	Header         [24]byte
	Miis           []ResortMii
}

type ResortMii struct {
	Index       uint8
	Initials    [2]byte
	CountryCode uint8
	_           uint32
	EntryNumber uint64
	MiiData     [76]byte
	ArtisanData [76]byte
}

func indexToEntryNumber(index int) uint64 {
	num := uint64(index)
	num ^= ((num << 0x1E) ^ (num << 0x12) ^ (num << 0x18)) & 0xFFFFFFFF
	num ^= (num & 0xF0F0F0F) << 4
	num ^= (num >> 0x1D) ^ (num >> 0x11) ^ (num >> 0x17) ^ 0x20070419

	crc := (num >> 8) ^ (num >> 24) ^ (num >> 16) ^ (num & 0xFF) ^ 0xFF

	i := uint64(0)
	if 0xD4A50FFF < num {
		i = 1
	}

	if 232 < (i + (crc & 0xFF)) {
		crc &= 0x7F
	}

	crc &= 0xFF

	binaryString := fmt.Sprintf("%08b%032b", crc, num)
	result, _ := strconv.ParseUint(binaryString, 2, 64)
	entryNumber := fmt.Sprintf("%012d", result)

	str, err := strconv.ParseUint(entryNumber, 10, 64)
	if err != nil {
		panic(err)
	}

	return str
}

func (w WiiSportsResort) ToBytes(_ any) []byte {
	buffer := new(bytes.Buffer)
	common.WriteBinary(buffer, w.OpenTimestamp)
	common.WriteBinary(buffer, w.CloseTimestamp)
	common.WriteBinary(buffer, w.Header)

	for _, mii := range w.Miis {
		common.WriteBinary(buffer, mii)
	}

	return buffer.Bytes()
}

func WriteWiiSportsResortMiis(pool *pgxpool.Pool, ctx context.Context) error {
	miis, err := GetPopularAndRandomMiis(pool, ctx, 100)
	if err != nil {
		return err
	}

	resort := WiiSportsResort{
		OpenTimestamp:  uint32(time.Now().Unix() - 946684800),
		CloseTimestamp: uint32(time.Now().Unix() - 946512000),
		Header:         [24]byte{0x00, 0x00, 0x41, 0xA0, 0x00, 0x00, 0x00, 0xA8, 0x64, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		Miis:           make([]ResortMii, len(miis)),
	}

	for i, mii := range miis {
		var tempMiiData [76]byte
		var tempArtisanMiiData [76]byte
		var tempInitials [2]byte

		copy(tempMiiData[:], mii.MiiData)
		copy(tempArtisanMiiData[:], mii.ArtisanMiiData)
		copy(tempInitials[:], mii.Initials)

		resort.Miis[i] = ResortMii{
			Index:       uint8(i + 1),
			Initials:    tempInitials,
			CountryCode: mii.CountryCode,
			EntryNumber: indexToEntryNumber(int(mii.EntryNumber)),
			MiiData:     tempMiiData,
			ArtisanData: tempArtisanMiiData,
		}
	}

	key, _ := hex.DecodeString("91D9A5DD10AAB467491A066EAD9FDD6F")
	rsa, err := os.ReadFile(common.GetConfig().AssetsPath + "/miidd.pem")
	if err != nil {
		return err
	}

	enc, err := libwc24crypt.EncryptWC24(resort.ToBytes(nil), key, iv, rsa)
	if err != nil {
		return err
	}

	return os.WriteFile(fmt.Sprintf("%s/dd/%s", common.GetConfig().AssetsPath, "miidd_018.enc"), enc, 0664)
}
