package contest

import (
	"bytes"
	"context"
	"github.com/WiiLink24/MiiContestChannel/common"
	"github.com/jackc/pgx/v4/pgxpool"
	"math"
	"time"
)

// ContestInfo is the complete structure of con_info.ces.
type ContestInfo struct {
	Header   ContestInfoHeader
	Contests []Contest
}

type ContestInfoHeader struct {
	Type         common.ListTag
	_            uint16
	CountryGroup uint32
	ListNumber   uint32
	ErrorCode    uint32
	_            [12]byte
	Padding      [4]byte
}

type Contest struct {
	Type         common.ListTag
	Size         uint16
	ContestIndex uint32
	ContestID    uint32
	Status       ContestStatus
	Options      uint8
	_            [18]byte
}

func (c ContestInfo) ToBytes(_ any) []byte {
	buffer := new(bytes.Buffer)
	common.WriteBinary(buffer, c.Header)

	for _, contest := range c.Contests {
		common.WriteBinary(buffer, contest)
	}

	return buffer.Bytes()
}

func setRankings(pool *pgxpool.Pool, ctx context.Context, contestId uint32) error {
	_, err := pool.Exec(ctx, UpdateContestMiiLikes, contestId)
	if err != nil {
		return err
	}

	rows, err := pool.Query(ctx, GetContestMiiLikes, contestId)
	if err != nil {
		return err
	}

	var miis []int
	for rows.Next() {
		var entryId int
		err = rows.Scan(&entryId)
		if err != nil {
			return err
		}

		miis = append(miis, entryId)
	}

	for i, mii := range miis {
		// Some formula used in the old RC24 code.
		percentile := math.Round((float64(i)/10)/float64(len(miis))*90) + 1

		_, err = pool.Exec(ctx, UpdateMiiRank, percentile, mii)
		if err != nil {
			return err
		}
	}

	return nil
}

func MakeContestInfos(pool *pgxpool.Pool, ctx context.Context) error {
	rows, err := pool.Query(ctx, GetContests)
	if err != nil {
		return err
	}

	var contests []Contest
	var contestDetails []*ContestDetail
	var index uint32 = 1
	for rows.Next() {
		contest := Contest{
			Type:         common.ContestInfo,
			Size:         32,
			ContestIndex: index,
			ContestID:    0,
			Status:       0,
			Options:      0,
		}

		var openTime *time.Time
		var closeTime *time.Time
		var description string
		var status DatabaseStatus
		err = rows.Scan(&contest.ContestID, &openTime, &closeTime, &description, &status)
		if err != nil {
			return err
		}

		// Statuses are:
		// - Waiting
		// - Open
		// - Judging
		// - Results
		// - Closed
		if status == Waiting && openTime.Before(time.Now().UTC()) && closeTime.After(time.Now().UTC()) {
			// Contest is ready to be opened.
			// TODO: Send the announcement to the message board
			_, err = pool.Exec(ctx, SetContestStatus, Open, contest.ContestID)
			if err != nil {
				return err
			}

			contest.Status = COpen
			status = Open
		} else if openTime.Before(time.Now()) && closeTime.Before(time.Now()) {
			if status == Open {
				// Set to judging. Contest will be open to judging for 7 days.
				*closeTime = closeTime.AddDate(0, 0, 7)
				_, err = pool.Exec(ctx, EditContestEndTimeAndStatus, Judging, closeTime, contest.ContestID)
				if err != nil {
					return err
				}

				contest.Status = CJudging

				// Create the entry lists
				err = MakeEntryLists(pool, ctx, contest.ContestID)
				if err != nil {
					return err
				}
			} else if status == Judging {
				// Set to results. Contest will be open to viewing for 1 month.
				*closeTime = closeTime.AddDate(0, 1, 0)
				_, err = pool.Exec(ctx, EditContestEndTimeAndStatus, Results, closeTime, contest.ContestID)
				if err != nil {
					return err
				}

				// Set rankings
				err = setRankings(pool, ctx, contest.ContestID)
				if err != nil {
					return err
				}

				// Finally create the list of best miis.
				err = MakeBestList(pool, ctx, contest.ContestID)
				if err != nil {
					return err
				}

				contest.Status = CResults
			} else if status == Results {
				_, err = pool.Exec(ctx, SetContestStatus, Closed, contest.ContestID)
				if err != nil {
					return err
				}

				contest.Status = CCLosed
			}
		} else {
			err = setRankings(pool, ctx, contest.ContestID)
			if err != nil {
				return err
			}

			contest.Status = DatabaseStatusToContestStatus(status)
		}

		if contest.Status != CCLosed && status != Waiting {
			detail, err := MakeContestDetailData(pool, ctx, contest.ContestID, openTime, closeTime, description, contest.Status)
			if err != nil {
				return err
			}

			for i := uint32(0); i < 7; i++ {
				detail.Language = i
				err = MakeContestDetailFile(detail)
				if err != nil {
					return err
				}
			}

			contestDetails = append(contestDetails, detail)
			contests = append(contests, contest)
			index++
		}
	}

	// TODO: Localization for contests + mail
	// For now we will just generate for all languages using English.
	for i := uint32(0); i < 7; i++ {
		err = MakeContestMail(contestDetails, i)
		if err != nil {
			return err
		}
	}

	info := ContestInfo{
		Header: ContestInfoHeader{
			Type:         common.ContestInfo,
			CountryGroup: 151,
			ListNumber:   0,
			ErrorCode:    0,
			Padding:      [4]byte{math.MaxUint8, math.MaxUint8, math.MaxUint8, math.MaxUint8},
		},
		Contests: contests,
	}

	return common.Write(info, "151/con_info.ces")
}
