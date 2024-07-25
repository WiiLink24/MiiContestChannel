package plaza

import (
	"context"
	"fmt"
	"github.com/WiiLink24/MiiContestChannel/common"
	"github.com/jackc/pgx/v4/pgxpool"
)

const GetBargainMiis = `SELECT miis.entry_id, miis.initials, miis.perm_likes, miis.skill, miis.country_id, miis.mii_data, 
       			artisans.mii_data, artisans.artisan_id, artisans.is_master 
				FROM miis, artisans WHERE miis.artisan_id = artisans.artisan_id 
				ORDER BY random() DESC LIMIT 450`

func MakeBargainList(pool *pgxpool.Pool, ctx context.Context) error {
	var miis []common.MiiWithArtisan
	rows, err := pool.Query(ctx, GetBargainMiis)
	if err != nil {
		return err
	}

	for rows.Next() {
		var isMaster bool
		var likes int
		mii := common.MiiWithArtisan{}
		err = rows.Scan(&mii.EntryNumber, &mii.Initials, &likes, &mii.Skill, &mii.CountryCode, &mii.MiiData,
			&mii.ArtisanMiiData, &mii.ArtisanId, &isMaster)
		if err != nil {
			return err
		}

		mii.Likes = uint8(likes)

		if isMaster {
			mii.IsMasterArtisan = 1
		}

		miis = append(miis, mii)
	}

	for i := 1; i < 11; i++ {
		var currMiis []common.MiiWithArtisan
		if i < 4 {
			currMiis = miis[:150]
			miis = miis[150:]
		}

		// Limit to 5 pages
		listNumber := uint32(i)
		err = MakeList(common.BargainList, currMiis, fmt.Sprintf("bargain_list%s.ces", common.ZFill(i, 2)), &listNumber)
		if err != nil {
			return err
		}
	}

	return nil
}
