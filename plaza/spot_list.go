package plaza

import (
	"context"
	"github.com/WiiLink24/MiiContestChannel/common"
	"github.com/jackc/pgx/v4/pgxpool"
)

const (
	GetPopularMiis = `SELECT miis.entry_id, miis.initials, miis.perm_likes, miis.skill, miis.country_id, miis.mii_data, 
       			artisans.mii_data, artisans.artisan_id, artisans.is_master 
				FROM miis, artisans WHERE miis.artisan_id = artisans.artisan_id 
				ORDER BY miis.likes LIMIT $1`

	GetPermPopularMiis = `SELECT miis.entry_id, miis.initials, miis.perm_likes, miis.skill, miis.country_id, miis.mii_data, 
       			artisans.mii_data, artisans.artisan_id, artisans.is_master 
				FROM miis, artisans WHERE miis.artisan_id = artisans.artisan_id AND miis.perm_likes > 21
				ORDER BY miis.entry_id LIMIT $1`

	GetNumberOfMiisWithLikes = `SELECT COUNT(*) FROM miis WHERE likes > 0`

	ResetMiiLikes = `UPDATE miis SET likes = 0`
)

func GetPopularAndRandomMiis(pool *pgxpool.Pool, ctx context.Context, max int) ([]common.MiiWithArtisan, error) {
	var numberOfMiisWithLikes int
	err := pool.QueryRow(ctx, GetNumberOfMiisWithLikes).Scan(&numberOfMiisWithLikes)
	if err != nil {
		return nil, err
	}

	// Based on original footage, Nintendo had 10 popular Mii pages. There are 50 Miis per page.
	var extraCount int
	if numberOfMiisWithLikes > max {
		numberOfMiisWithLikes = max
	} else {
		extraCount = max - numberOfMiisWithLikes
	}

	// First handle popular
	var miis []common.MiiWithArtisan
	rows, err := pool.Query(ctx, GetPopularMiis, numberOfMiisWithLikes)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var isMaster bool
		var likes int
		mii := common.MiiWithArtisan{}
		err = rows.Scan(&mii.EntryNumber, &mii.Initials, &likes, &mii.Skill, &mii.CountryCode, &mii.MiiData,
			&mii.ArtisanMiiData, &mii.ArtisanId, &isMaster)
		if err != nil {
			return nil, err
		}

		// Downcast to u8 as database contains numbers larger.
		mii.Likes = uint8(likes)

		if isMaster {
			mii.IsMasterArtisan = 1
		}

		miis = append(miis, mii)
	}

	// Now handle the random miis.
	rows, err = pool.Query(ctx, GetPermPopularMiis, extraCount)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var isMaster bool
		var likes int
		mii := common.MiiWithArtisan{}
		err = rows.Scan(&mii.EntryNumber, &mii.Initials, &likes, &mii.Skill, &mii.CountryCode, &mii.MiiData,
			&mii.ArtisanMiiData, &mii.ArtisanId, &isMaster)
		if err != nil {
			return nil, err
		}

		// Downcast to u8 as database contains numbers larger.
		mii.Likes = uint8(likes)

		if isMaster {
			mii.IsMasterArtisan = 1
		}

		miis = append(miis, mii)
	}

	return miis, nil
}

func MakeSpotList(pool *pgxpool.Pool, ctx context.Context) error {
	miis, err := GetPopularAndRandomMiis(pool, ctx, 500)
	if err != nil {
		return err
	}

	err = MakeList(common.SpotList, miis, "spot_list.ces", nil)
	if err != nil {
		return err
	}

	// Finally reset the temp likes.
	_, err = pool.Exec(ctx, ResetMiiLikes)
	if err != nil {
		return err
	}

	return nil
}
