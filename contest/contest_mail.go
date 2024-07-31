package contest

import (
	"bytes"
	"cmp"
	_ "embed"
	"fmt"
	"github.com/SketchMaster2001/libwc24crypt"
	"github.com/WiiLink24/MiiContestChannel/common"
	"github.com/WiiLink24/nwc24"
	"github.com/wii-tools/libtpl"
	"image/jpeg"
	"math/rand"
	"net/mail"
	"os"
	"slices"
	"strconv"
	"time"
	"unicode/utf16"
)

var (
	ContestPosting = func(theme string) string {
		return fmt.Sprintf("*******************************\r\nA New Contest is Under Way\r\n*******************************\r\n\r\nCare to test your Mii-making skills\r\nby designing a Mii on a particular\r\ntheme?\r\n\r\n\u25c6Contest Theme:\r\n%s\r\n\r\n\u25c6How to Submit an Entry\r\n1. Design a Mii in the Mii\r\n   Channel.\r\n2. Go to the Check Mii Out\r\n   Channel and submit your Mii.\r\n\r\n\r\nP.S. Check out https://miicontest.wiilink.ca,\r\nit's the official companion website\r\nfor the Check Mii Out Channel!\r\n\r\n\r\n----------------------------------\r\nThis message is regarding the\r\nCheck Mii Out Channel.\r\n\r\nIf you do not wish to receive further\r\ncommercial messages from WiiLink,\r\nplease click the opt-out icon on the \r\nupper-right corner of the screen.\r\n\r\nYou can opt out of either (1) \r\nmessages for the Check Mii Out\r\nChannel only or (2) all messages for\r\nall channels and games.", theme)
	}

	ContestJudging = func(theme string) string {
		return fmt.Sprintf("*******************************\r\nCome and Judge a Contest\r\n*******************************\r\n\r\nCome over to the Check Mii Out\r\nChannel and judge a few Miis\r\nfor a contest.\r\n\r\n\u25c6Contest Theme:\r\n%s\r\n\r\n\r\nP.S. Check out https://miicontest.wiilink.ca,\r\nit's the official companion website\r\nfor the Check Mii Out Channel!\r\n\r\n\r\n----------------------------------\r\n\r\n\r\nThis message is regarding the\r\nCheck Mii Out Channel.\r\n\r\nIf you do not wish to receive further\r\ncommercial messages from WiiLink,\r\nplease click the opt-out icon on the \r\nupper-right corner of the screen.\r\n\r\nYou can opt out of either (1) \r\nmessages for the Check Mii Out\r\nChannel only or (2) all messages for\r\nall channels and games.", theme)
	}

	ContestResults = func(theme string) string {
		return fmt.Sprintf("*******************************\r\nContest Results\r\n*******************************\r\n\r\nWe've tallied up all the votes, and\r\nthe winners for this contest have\r\nbeen decided!\r\n\r\n\u25c6Contest Theme:\r\n%s\r\n\r\n\r\nP.S. Check out https://miicontest.wiilink.ca,\r\nit's the official companion website\r\nfor the Check Mii Out Channel!\r\n\r\n\r\n----------------------------------\r\nThis message is regarding the\r\nCheck Mii Out Channel.\r\n\r\nIf you do not wish to receive further\r\ncommercial messages from WiiLink,\r\nplease click the opt-out icon on the \r\nupper-right corner of the screen.\r\n\r\nYou can opt out of either (1) \r\nmessages for the Check Mii Out\r\nChannel only or (2) all messages for\r\nall channels and games.", theme)
	}

	key = []byte{0xBE, 0x37, 0x15, 0xC3, 0x08, 0xF3, 0x41, 0xA8, 0xF1, 0x6F, 0x0E, 0xF4, 0xFB, 0x14, 0x97, 0xAF}

	iv = []byte{70, 70, 20, 40, 143, 110, 36, 6, 184, 107, 135, 239, 96, 45, 80, 151}

	//go:embed cmoc_letter.arc
	cmocLetterScript []byte
)

func MakeContestMail(contests []*ContestDetail, languageCode uint32) error {
	latestContest := slices.MaxFunc(contests, func(a, b *ContestDetail) int {
		return cmp.Compare(a.ContestID, b.ContestID)
	})

	// We have to set the language code.
	latestContest.Language = languageCode

	to, _ := mail.ParseAddress("allusers@rc24.xyz")
	from, _ := mail.ParseAddress("w9999999900000000@rc24.xyz")

	_min := 1000000
	_max := 9999999
	randomNum := rand.Intn(_max-_min+1) + _min

	boundary := "--BoundaryForDL" + time.Now().UTC().Format("200601021504") + "/" + strconv.Itoa(randomNum)
	message := nwc24.NewMessage(from, to)
	message.SetContentType(nwc24.MultipartMixed)
	message.SetBoundary("----=_CMOC_Contest_Details")
	message.SetTag("X-Wii-AppID", "3-48415041-3031")
	message.SetTag("X-Wii-Tag", "00000001")
	message.SetTag("X-Wii-Cmd", "00080001")

	contestDataPart := nwc24.NewMultipart()
	contestDataPart.AddFile(fmt.Sprintf("con_detail%d.bin", languageCode), latestContest.ToBytes(latestContest), nwc24.Binary)

	message.AddMultipart(contestDataPart)
	if ((latestContest.Options >> 1) & 1) == 1 {
		thumb, err := os.ReadFile(fmt.Sprintf("%s/contest/%d/thumbnail.jpg", common.GetConfig().AssetsPath, latestContest.ContestID))
		if err != nil {
			return err
		}

		img, err := jpeg.Decode(bytes.NewReader(thumb))
		if err != nil {
			return err
		}

		encoded, err := libtpl.ToRGB565(img)
		if err != nil {
			return err
		}

		thumbnailPart := nwc24.NewMultipart()
		thumbnailPart.AddFile(fmt.Sprintf("thumbnail_%d.tpl", latestContest.ContestID), encoded[64:], nwc24.Binary)
		message.AddMultipart(thumbnailPart)
	}

	var messages []*nwc24.Message
	messages = append(messages, message)
	for i, contest := range contests {
		currBoundary := fmt.Sprintf("----=_CMOC_Contest_%d", i+1)
		currMessageBody := GetMessageForStatus(contest.Status, string(contest.Description[:]))

		currMessage := nwc24.NewMessage(from, to)
		currMessage.SetContentType(nwc24.MultipartMixed)
		currMessage.SetBoundary(currBoundary)

		currMessage.SetTag("X-Wii-MB-OptOut", "1")
		currMessage.SetTag("X-Wii-MB-NoReply", "1")
		currMessage.SetTag("X-Wii-AppID", "3-48415041-3031")

		messagePart := nwc24.NewMultipart()
		messagePart.SetText(currMessageBody, nwc24.UTF16BE)

		letterPart := nwc24.NewMultipart()
		letterPart.AddFile("cmoc_letterform.arc", cmocLetterScript, nwc24.WiiMessageBoard)

		currMessage.AddMultipart(messagePart, letterPart)
		messages = append(messages, currMessage)
	}

	// Finally formulate the complete message.
	completeMessage, err := nwc24.CreateMessageToSend(boundary, messages...)
	if err != nil {
		return err
	}

	rsa, err := os.ReadFile(common.GetConfig().AssetsPath + "/cmoc.pem")
	if err != nil {
		return err
	}

	enc, err := libwc24crypt.EncryptWC24([]byte(completeMessage), key, iv, rsa)
	if err != nil {
		return err
	}

	return os.WriteFile(fmt.Sprintf("%s/151/con_task%d.bin", common.GetConfig().AssetsPath, languageCode), enc, 0664)
}

func GetMessageForStatus(status ContestStatus, theme string) string {
	switch status {
	case COpen:
		return nwc24.UTF16ToString(utf16.Encode([]rune(ContestPosting(theme))))
	case CJudging:
		return nwc24.UTF16ToString(utf16.Encode([]rune(ContestJudging(theme))))
	case CResults:
		return nwc24.UTF16ToString(utf16.Encode([]rune(ContestResults(theme))))
	}

	return ""
}
