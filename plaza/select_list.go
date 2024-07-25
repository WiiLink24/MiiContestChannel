package plaza

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/WiiLink24/MiiContestChannel/common"
	"github.com/WiiLink24/MiiContestChannel/first"
	"github.com/jackc/pgx/v4/pgxpool"
)

const GetMiisBySkillAndGender = `SELECT miis.entry_id, miis.initials, miis.perm_likes, miis.skill, miis.country_id, miis.mii_data, 
       			artisans.mii_data, artisans.artisan_id, artisans.is_master 
				FROM miis, artisans WHERE miis.artisan_id = artisans.artisan_id AND miis.skill = $1 AND miis.gender = $2
				ORDER BY miis.likes DESC LIMIT 50`

func MakeSelectList(pool *pgxpool.Pool, ctx context.Context) error {
	var root first.Root
	err := json.Unmarshal(first.AdditionJson, &root)
	if err != nil {
		return err
	}

	for i := 1; i <= 2; i++ {
		// Male = 1
		// Female = 2
		for _, skill := range root.Skills {
			var miis []common.MiiWithArtisan
			rows, err := pool.Query(ctx, GetMiisBySkillAndGender, skill.Code, i)
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

			rows.Close()
			listNumber := skill.Code*10 + uint32(i)
			err = MakeList(common.SelectList, miis, fmt.Sprintf("select_list%03d.ces", listNumber), &listNumber)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
