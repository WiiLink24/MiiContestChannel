package common

type ListTag [2]byte

var (
	FirstList       = ListTag{'F', 'D'}
	SpotList        = ListTag{'S', 'L'}
	BargainList     = ListTag{'R', 'L'}
	NewList         = ListTag{'N', 'L'}
	ContestDetail   = ListTag{'C', 'D'}
	ContestInfo     = ListTag{'C', 'I'}
	PopcraftsList   = ListTag{'C', 'L'}
	PopularList     = ListTag{'P', 'L'}
	ExtendedArtisan = ListTag{'R', 'C'}
	RK              = ListTag{'R', 'K'}
	PairMii         = ListTag{'P', 'M'}
	PairArtisan     = ListTag{'P', 'C'}
	EntryList       = ListTag{'E', 'L'}
	ContestMii      = ListTag{'C', 'M'}
	BestList        = ListTag{'B', 'L'}
	ContestTag      = ListTag{'C', 'N'}
	ContestArtisan  = ListTag{'C', 'C'}
	Thumbnail       = ListTag{'T', 'H'}
	MiiPairNumber   = ListTag{'P', 'N'}
	SpecialList     = ListTag{'I', 'L'}
	NumberInfo      = ListTag{'N', 'I'}
	Souvenir        = ListTag{'P', 'H'}
	ArtisanInfo     = ListTag{'I', 'N'}
	MiiInfo         = ListTag{'I', 'M'}
	OwnSearch       = ListTag{'O', 'S'}
	NameSearch      = ListTag{'N', 'S'}
	Search          = ListTag{'S', 'R'}
	SelectList      = ListTag{'L', 'L'}
)

// LanguageFlag is a bitmask that allows for toggling supported languages for a country.
type LanguageFlag uint8

const (
	Japanese LanguageFlag = 1 << 0
	English  LanguageFlag = 1 << 1
	German   LanguageFlag = 1 << 2
	French   LanguageFlag = 1 << 3
	Spanish  LanguageFlag = 1 << 4
	Italian  LanguageFlag = 1 << 5
	Dutch    LanguageFlag = 1 << 6
)

var CountryCodes = []uint8{1, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52, 64, 65, 66, 67, 68, 69, 70, 71, 72, 73, 74, 75, 76, 77, 78, 79, 80, 81, 82, 83, 84, 85, 86, 87, 88, 89, 90, 91, 92, 93, 94, 95, 96, 97, 98, 99, 100, 101, 102, 103, 104, 105, 106, 107, 108, 109, 110, 136}
