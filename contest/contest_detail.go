package contest

import (
	"context"
	"fmt"
	"github.com/WiiLink24/MiiContestChannel/common"
	"github.com/jackc/pgx/v4/pgxpool"
	"math"
	"time"
)

// ContestDetail contains the metadata for contest.
type ContestDetail struct {
	Tag            common.ListTag
	_              uint16
	ContestID      uint32
	Language       uint32
	_              [8]byte
	StartTimestamp uint32
	EndTimestamp   uint32
	Padding        [4]byte
	CDTag          common.ListTag
	TagSize        uint16
	Unknown2       uint32
	ContestID2     uint32
	Status         ContestStatus
	Options        Option
	_              uint16
	EntryCount     uint32
	Padding2       [20]byte
	Topic          [32]byte
	Description    [64]byte
}

func (c ContestDetail) ToBytes(data any) []byte {
	return common.ToBytes(data)
}

func MakeContestDetailData(pool *pgxpool.Pool, ctx context.Context, contestId uint32, startTime, endTime *time.Time, description string, status ContestStatus) (*ContestDetail, error) {
	// Get entry numbers
	var entryCount uint32
	var hasThumbnail bool
	var hasSpecialAward bool
	if status != COpen {
		err := pool.QueryRow(ctx, GetContestTraits, contestId).Scan(&hasThumbnail, &hasSpecialAward)
		if err != nil {
			return nil, err
		}

		err = pool.QueryRow(ctx, GetNumberOfContestMiis, contestId).Scan(&entryCount)
		if err != nil {
			return nil, err
		}
	} else {
		err := pool.QueryRow(ctx, GetContestThumbnailStatus, contestId).Scan(&hasThumbnail, &hasSpecialAward)
		if err != nil {
			return nil, err
		}
	}

	var topic [32]byte
	var tempDescription [64]byte
	copy(topic[:], description)
	copy(tempDescription[:], description)

	var options Option
	options |= Worldwide
	options |= NicknameChanging
	if hasThumbnail {
		options |= Thumbnail
	}

	if hasSpecialAward {
		options |= SpecialAward
	}

	detail := ContestDetail{
		Tag:       common.ContestDetail,
		ContestID: contestId,
		// Gets changed by contest_info.go
		Language:       0,
		StartTimestamp: uint32(startTime.Unix() - 946684800),
		EndTimestamp:   uint32(endTime.Unix() - 946684800),
		Padding:        [4]byte{math.MaxUint8, math.MaxUint8, math.MaxUint8, math.MaxUint8},
		CDTag:          common.ContestDetail,
		TagSize:        136,
		Unknown2:       1,
		ContestID2:     contestId,
		Status:         status,
		Options:        options,
		EntryCount:     entryCount,
		Padding2:       [20]byte{},
		Topic:          topic,
		Description:    tempDescription,
	}

	return &detail, nil
}

func MakeContestDetailFile(detail *ContestDetail) error {
	return common.Write(detail, fmt.Sprintf("contest/%d/con_detail%d.ces", detail.ContestID, detail.Language))
}
