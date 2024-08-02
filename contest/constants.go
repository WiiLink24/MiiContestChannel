package contest

const (
	GetContests                 = `SELECT contest_id, open_time, close_time, english_name, status FROM contests`
	SetContestStatus            = `UPDATE contests SET status = $1 WHERE contest_id = $2`
	EditContestEndTimeAndStatus = `UPDATE contests SET status = $1, close_time = $2 WHERE contest_id = $3`
	GetContestMiis              = `SELECT artisan_id, mii_data FROM contest_miis WHERE contest_id = $1`
	GetContestTraits            = `SELECT has_thumbnail, has_special_award FROM contests WHERE contest_id = $1 AND contests.status != 'closed' AND contests.status != 'waiting'`
	GetNumberOfContestMiis      = `SELECT COUNT(*) FROM contest_miis WHERE contest_id = $1`
	GetContestThumbnailStatus   = `SELECT has_thumbnail, has_special_award FROM contests WHERE contest_id = $1`
	UpdateContestMiiLikes       = `UPDATE contest_miis SET likes = 
    			(SELECT COUNT(*) FROM contest_votes WHERE contest_votes.contest_id = $1 AND 
    			(vote_1 = contest_miis.artisan_id OR vote_2 = contest_miis.artisan_id OR vote_3 = contest_miis.artisan_id)) 
               	WHERE contest_id = $1`
	GetContestMiiLikes = `SELECT entry_id FROM contest_miis WHERE contest_id = $1 ORDER BY likes`
	UpdateMiiRank      = `UPDATE contest_miis SET rank = $1 WHERE entry_id = $2`
	GetBestContestMiis = `SELECT contest_miis.entry_id, contest_miis.artisan_id, contest_miis.mii_data, artisans.mii_data, 
       						artisans.country_id, artisans.is_master FROM contest_miis, artisans 
       						WHERE contest_miis.artisan_id = artisans.artisan_id AND contest_miis.contest_id = $1 
       						ORDER BY contest_miis.likes DESC LIMIT 50`
)

// DatabaseStatus are the states in which a contest can be in within the database.
type DatabaseStatus string

const (
	Waiting DatabaseStatus = "waiting"
	Open    DatabaseStatus = "open"
	Judging DatabaseStatus = "judging"
	Results DatabaseStatus = "results"
	Closed  DatabaseStatus = "closed"
)

// ContestStatus is the value of a contest state.
type ContestStatus uint8

const (
	COpen    ContestStatus = 2 << 0
	CJudging ContestStatus = 2 << 2
	CResults ContestStatus = 2 << 4
	// CCLosed is not used in the channel, rather as a state to not generate a contest file.
	CCLosed ContestStatus = 2 << 6
)

func DatabaseStatusToContestStatus(status DatabaseStatus) ContestStatus {
	switch status {
	case Open:
		return COpen
	case Judging:
		return CJudging
	case Results:
		return CResults
	}

	return COpen
}

// Option are the multiple options that a contest can have.
type Option uint8

const (
	Worldwide        Option = 1 << 0
	Thumbnail        Option = 1 << 1
	Souvenir         Option = 1 << 2
	NicknameChanging Option = 1 << 3
	SpecialAward     Option = 1 << 4
)
