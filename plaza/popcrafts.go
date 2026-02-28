package plaza

import (
	"bytes"
	"context"
	"github.com/WiiLink24/MiiContestChannel/common"
	"github.com/jackc/pgx/v4/pgxpool"
	"math"
	"time"
)

const (
	GetPopularArtisans = `SELECT artisan_id FROM artisans WHERE artisan_id != 100000993 ORDER BY total_likes DESC LIMIT 100`
	GetArtisan         = `SELECT mii_data, total_likes, country_id, is_master, last_post FROM artisans WHERE artisan_id = $1`
)

type Popcrafts struct {
	Header   PopcraftsHeader
	Artisans []ExtendedArtisan
}

type PopcraftsHeader struct {
	Tag           [2]byte
	_             uint16
	CountryRegion uint32
	ListNumber    uint32
	ErrorCode     uint32
	_             [12]byte
	Padding       [4]byte
}

type ExtendedArtisan struct {
	Tag                [2]byte
	TagSize            uint16
	ArtisanIndex       uint32
	ArtisanNumber      uint32
	MiiArtisanData     [76]byte
	_                  uint8
	IsMasterMiiArtisan uint8
	Popularity         uint8
	Arrow              uint8
	CountryCode        uint16
	_                  uint16
	RK
}

type RK struct {
	Tag          [2]byte
	TagSize      uint16
	Unknown      uint32
	ArtisanIndex uint32
	Day          uint16
	Month        uint8
	_            uint8
}

func (p Popcrafts) ToBytes(_ any) []byte {
	buffer := new(bytes.Buffer)
	common.WriteBinary(buffer, p.Header)

	for _, artisan := range p.Artisans {
		common.WriteBinary(buffer, artisan)
	}

	return buffer.Bytes()
}

func MakePopCraftsList(pool *pgxpool.Pool, ctx context.Context) error {
	var popcrafts Popcrafts

	// Formulate popcrafts header
	popcrafts.Header = PopcraftsHeader{
		Tag:           common.PopcraftsList,
		CountryRegion: 151,
		ListNumber:    0,
		ErrorCode:     0,
		Padding:       [4]byte{math.MaxUint8, math.MaxUint8, math.MaxUint8, math.MaxUint8},
	}

	rows, err := pool.Query(ctx, GetPopularArtisans)
	if err != nil {
		return err
	}

	var bareArtisans []common.MiiWithArtisan
	for rows.Next() {
		artisan := common.MiiWithArtisan{}
		err = rows.Scan(&artisan.ArtisanId)
		if err != nil {
			return err
		}

		bareArtisans = append(bareArtisans, artisan)
	}

	for i, artisan := range bareArtisans {
		var lastPost *time.Time
		var isMaster bool
		var likes int
		err = pool.QueryRow(ctx, GetArtisan, artisan.ArtisanId).Scan(&artisan.ArtisanMiiData, &likes, &artisan.CountryCode, &isMaster, &lastPost)
		if err != nil {
			return err
		}

		artisan.Likes = uint8(likes)

		if isMaster {
			artisan.IsMasterArtisan = 1
		}

		// If the Mii was posted in the last 24 hours, it is considered new.
		if lastPost.AddDate(0, 0, 1).After(time.Now()) {
			// TODO: The flag thing
			if isMaster {
				artisan.IsMasterArtisan = 3
			} else {
				artisan.IsMasterArtisan = 2
			}
		}

		var tempArtisanMiiData [76]byte
		copy(tempArtisanMiiData[:], artisan.ArtisanMiiData)

		popcrafts.Artisans = append(popcrafts.Artisans, ExtendedArtisan{
			Tag:                common.ExtendedArtisan,
			TagSize:            96,
			ArtisanIndex:       uint32(i + 1),
			ArtisanNumber:      artisan.ArtisanId,
			MiiArtisanData:     tempArtisanMiiData,
			IsMasterMiiArtisan: artisan.IsMasterArtisan,
			Popularity:         artisan.Likes,
			CountryCode:        uint16(artisan.CountryCode),
			RK: RK{
				Tag:          common.RK,
				TagSize:      16,
				Unknown:      1,
				ArtisanIndex: uint32(i + 1),
				Day:          uint16(time.Now().Day()),
				Month:        uint8(time.Now().Month()),
			},
		})
	}

	return common.Write(popcrafts, "151/popcrafts_list.ces")
}
