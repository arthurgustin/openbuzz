package orm

import (
	"github.com/jinzhu/gorm"
	"github.com/golang-plus/uuid"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"fmt"
	"strings"
	"github.com/bobesa/go-domain-util/domainutil"
	"github.com/davecgh/go-spew/spew"
	"open-buzz/shared"
)

type Client struct {
	Db *gorm.DB
	Logger shared.LoggerInterface `inject:""`
}

func NewClient() (*Client, error) {
	db, err := initDatabase()
	if err != nil {
		return nil, err
	}
	return &Client{
		Db: db,
	}, nil
}

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "postgres"
	dbname   = "openbuzz"
)
func initDatabase() (*gorm.DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := gorm.Open("postgres", psqlInfo)
	if err != nil {
		panic("failed to connect database")
	}
	db.LogMode(false)
	// Migrate the schema
	db.AutoMigrate(&dbProspectInfo{})
	db.AutoMigrate(&dbProspect{})
	return db, err
}

func NewProspect(url string) *Prospect {
	prospect := &Prospect{}
	return prospect.SetUrl(url)
}

type Prospect struct {
	ProspectId string
	prospect dbProspect
	infos []dbProspectInfo
}

type dbProspect struct {
	gorm.Model
	ProspectID string `gorm:"not null;unique"`
	Url string `gorm:"not null"`
	FirstName string
	MiddleName string
	LastName string
}

type dbProspectInfo struct {
	gorm.Model
	ProspectID string `gorm:"not null"`
	Key string
	Val string
	Confidence float64 // [0 - 1]
	ValidatedByUser bool
}

func (p *Prospect) SetIcon(targetUrl string) *Prospect {
	return p.addInfo("icon", targetUrl, 1)
}

func (p *Prospect) SetUrl(targetUrl string) *Prospect {
	p.prospect.Url = targetUrl
	return p.addInfo("domain", targetUrl, 1)
}

func (p *Prospect) GetUrl() string {
	return p.prospect.Url
}

func (p *Prospect) GetHost() string {
	domain := domainutil.Domain(p.GetUrl())

	// e.g for korben.info, returns korben
	return strings.Split(domain, ".")[0]
}

// Don't forget to update this slice when a new social media is added
var allSocialMedia = []string{"facebook", "twitter", "youtube", "google", "linkedin"}
func (p *Prospect) SetSocial(name, url string, confidence float64) *Prospect {
	switch name {
	case "facebook":
		return p.addInfo("facebook", url, confidence)
	case "twitter":
		return p.addInfo("twitter", url, confidence)
	case "google":
		return p.addInfo("google", url, confidence)
	case "linkedin":
		return p.addInfo("linkedin", url, confidence)
	case "youtube":
		return p.addInfo("youtube", url, confidence)
	}
	return p
}

func (p *Prospect) SetEmail(email string, confidence float64) *Prospect {
	return p.addInfo("email", email, confidence)
}

func (p *Prospect) SetFirstName(firstName string) *Prospect {
	p.prospect.FirstName = strings.ToLower(firstName)
	return p
}

func (p *Prospect) GetFirstName() string {
	return p.prospect.FirstName
}


func (p *Prospect) SetMiddleName(middleName string) *Prospect {
	p.prospect.MiddleName = strings.ToLower(middleName)
	return p
}

func (p *Prospect) GetMiddleName() string {
	return p.prospect.MiddleName
}

func (p *Prospect) SetLastName(lastName string) *Prospect {
	p.prospect.LastName = strings.ToLower(lastName)
	return p
}

func (p *Prospect) GetLastName() string {
	return p.prospect.LastName
}

func (p *Prospect) addInfo(key, val string, confidence float64) *Prospect {
	p.infos = append(p.infos, dbProspectInfo{
		ProspectID: p.prospect.ProspectID,
		Key: key,
		Val: val,
		ValidatedByUser: false,
		Confidence: confidence,
	})
	return p
}

func (c *Client) Save(p *Prospect) error {
	p.prospect.ProspectID = c.getOrCreateProspectId(p.GetUrl())

	if err := c.saveDbProspect(p.prospect); err != nil {
		return err
	}

	for _, info := range p.infos {
		transaction := c.Db.Begin()
		info.ProspectID = p.prospect.ProspectID
		if c.infoExist(info) {
			continue
		}
		if err := transaction.Create(&info).Error; err != nil {
			transaction.Rollback()
			return err
		}
		transaction.Commit()
	}

	return nil
}

func (c *Client) List() (list []Prospect, err error) {
	allProspects := []dbProspect{}

	if err = c.Db.Model(&dbProspect{}).Find(&allProspects).Error; err != nil {
		c.Logger.Warn(err.Error())
		return
	}

	for _, prospect := range allProspects {
		prospectsInfo := []dbProspectInfo{}

		if err = c.Db.Model(&dbProspectInfo{}).
			Where("prospect_id = ?", prospect.ProspectID).
			Find(&prospectsInfo).Error; err != nil {
				c.Logger.Warn(err.Error())
				return
		}

		list = append(list, Prospect{
			ProspectId: prospect.ProspectID,
			prospect: prospect,
			infos: prospectsInfo,
		})
	}
	return
}

type Email struct {
	Email string
	Confidence float64
	ValidatedByUser bool
}

func (c *Client) GetEmails(prospectId string) (emails []Email, err error) {
	infos := []dbProspectInfo{}

	if err = c.Db.Model(&dbProspectInfo{}).
	Where("prospect_id = ? AND key = ?", prospectId, "email").
	Find(&infos).Error; err != nil {
		c.Logger.Warn(err.Error())
		return
	}

	for _, info := range infos {
		emails = append(emails, Email{
			Email: info.Val,
			Confidence: info.Confidence,
			ValidatedByUser: info.ValidatedByUser,
		})
	}
	return
}

type SocialMedia struct {
	Name string
	Url string
	Confidence float64
	ValidatedByUser bool
}

func (c *Client) GetSocialMedia(prospectId string) (socialMedias []SocialMedia, err error) {
	infos := []dbProspectInfo{}

	if err = c.Db.Model(&dbProspectInfo{}).
		Where("prospect_id = ? AND key IN (?)", prospectId, allSocialMedia).
		Find(&infos).Error; err != nil {
		c.Logger.Warn(err.Error())
		return
	}

	for _, info := range infos {
		socialMedias = append(socialMedias, SocialMedia{
			Name: info.Key,
			Url: info.Val,
			Confidence: info.Confidence,
			ValidatedByUser: info.ValidatedByUser,
		})
	}
	return
}

type Assets struct {
	Icons []Icon
}

type Icon struct {
	Link string
}

func (c *Client) GetAssets(prospectId string) (assets Assets, err error) {
	infos := []dbProspectInfo{}

	if err = c.Db.Model(&dbProspectInfo{}).
		Where("prospect_id = ? AND key = ?", prospectId, "icon").
		Find(&infos).Error; err != nil {
		c.Logger.Warn(err.Error())
		return
	}

	for _, info := range infos {
		assets.Icons = append(assets.Icons, Icon{
			Link: info.Val,
		})
	}
	return
}

func (c *Client) saveDbProspect(p dbProspect) error {
	transaction := c.Db.Begin()

	pro := dbProspect{}
	if notFound := transaction.Model(&dbProspect{}).
	Where("url = ?", p.Url).Scan(&pro).RecordNotFound(); !notFound {
		transaction.Rollback()
		fmt.Println("already exists")
		return nil
	}

	if err := transaction.Model(&dbProspect{}).Create(&p).Error; err != nil {
		transaction.Rollback()
		return err
	}
	transaction.Commit()
	return nil
}

func (c *Client) infoExist(info dbProspectInfo) bool {
	var count int
	if err := c.Db.Model(&dbProspectInfo{}).
	Where("key = ? AND val = ? AND prospect_id = ?", info.Key, info.Val, info.ProspectID).
	Count(&count).Error; err != nil {
		panic(err)
	}
	if count > 0 {
		fmt.Println(fmt.Sprintf("prospect information already exists: %s=%s", info.Key, info.Val))
		return true
	}
	return false
}

func (c *Client) getOrCreateProspectId(url string) string {
	var prospect dbProspect
	c.Db.Model(&dbProspect{}).
	Where("url = ?", url).First(&prospect)

	spew.Dump(prospect)

	if prospect.Url != "" {
		return prospect.ProspectID
	}

	id, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}
	return id.String()
}