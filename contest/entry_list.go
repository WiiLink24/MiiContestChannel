package contest

import (
	"bytes"
	"context"
	"fmt"
	"github.com/WiiLink24/MiiContestChannel/common"
	"github.com/jackc/pgx/v4/pgxpool"
	"math"
)

type EntryInfo struct {
	Header            EntryInfoHeader
	ContestMiiEntries []ContestMiiEntry
}

type EntryInfoHeader struct {
	Type       common.ListTag
	_          uint16
	ContestId  uint32
	ListNumber uint32
	ErrorCode  uint32
	_          [12]byte
	Padding    [4]byte
}

type ContestMiiEntry struct {
	Type      common.ListTag
	TagSize   uint16
	MiiIndex  uint32
	ArtisanId uint32
	MiiData   [76]byte
}

func (c EntryInfo) ToBytes(_ any) []byte {
	buffer := new(bytes.Buffer)
	common.WriteBinary(buffer, c.Header)

	for _, contest := range c.ContestMiiEntries {
		common.WriteBinary(buffer, contest)
	}

	return buffer.Bytes()
}

// readMiiList returns 10 entries from the front of the slice.
// Returns the remaining entries if it is less than 10
func readMiiList(slice []ContestMiiEntry) []ContestMiiEntry {
	if len(slice) < 10 {
		return slice
	}

	return slice[:10]
}

func removeSliceEntries(slice []ContestMiiEntry) []ContestMiiEntry {
	if 10 >= len(slice) {
		// If n is greater than or equal to the length of the slice, return an empty slice
		return []ContestMiiEntry{}
	}

	return slice[10:]
}

func MakeEntryLists(pool *pgxpool.Pool, ctx context.Context, contestId uint32) error {
	rows, err := pool.Query(ctx, GetContestMiis, contestId)
	if err != nil {
		return err
	}

	var index uint32 = 1
	var miis []ContestMiiEntry
	for rows.Next() {
		var artisanId int
		var _miiData []byte
		err = rows.Scan(&artisanId, &_miiData)
		if err != nil {
			return err
		}

		var miiData [76]byte
		copy(miiData[:], _miiData)

		miis = append(miis, ContestMiiEntry{
			Type:      common.ContestMii,
			TagSize:   88,
			MiiIndex:  index,
			ArtisanId: uint32(artisanId),
			MiiData:   miiData,
		})

		// Every list contains up to 10 miis.
		index++
	}

	val := math.Ceil(float64(len(miis)) / 10)
	for i := 0; i < int(val); i++ {
		header := EntryInfoHeader{
			Type:       common.EntryList,
			ContestId:  contestId,
			ListNumber: uint32(i + 1),
			ErrorCode:  0,
			Padding:    [4]byte{math.MaxUint8, math.MaxUint8, math.MaxUint8, math.MaxUint8},
		}

		entryInfo := EntryInfo{
			Header:            header,
			ContestMiiEntries: readMiiList(miis),
		}

		miis = removeSliceEntries(miis)
		err = common.Write(entryInfo, fmt.Sprintf("contest/%d/entry_list%d.ces", contestId, i+1))
		if err != nil {
			return err
		}
	}

	return nil
}
