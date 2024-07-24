package plaza

import (
	"context"

	"github.com/WiiLink24/MiiContestChannel/common"
	"github.com/jackc/pgx/v4/pgxpool"
)

const (
	GetTop50Miis = `SELECT miis.entry_id, miis.initials, miis.perm_likes, miis.skill, miis.country_id, miis.mii_data, 
       			artisans.mii_data, artisans.artisan_id, artisans.is_master 
				FROM miis, artisans WHERE miis.artisan_id = artisans.artisan_id 
				ORDER BY miis.perm_likes DESC LIMIT 50`
)

func MakeTop50(pool *pgxpool.Pool, ctx context.Context) error {
	var miis []common.MiiWithArtisan
	rows, err := pool.Query(ctx, GetTop50Miis)
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

	err = MakeList(common.PopularList, miis, "pop_list.ces", nil)
	if err != nil {
		return err
	}

	return nil
}
