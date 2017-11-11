package orm

import (
	"errors"
	"fmt"
	"github.com/arthurgustin/openbuzz/shared"
	"github.com/golang-plus/uuid"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var ErrFailedToConnectToDabase = errors.New("failed to connect database")

type Client struct {
	Db     *gorm.DB
	Logger shared.LoggerInterface `inject:""`
	Config *shared.AppConfig      `inject:""`
}

func (c *Client) Init() error {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.Config.PgHost, c.Config.PgPort, c.Config.PgUser, c.Config.PgPassword, c.Config.PgDbName)
	c.Logger.Info("trying to connect to postgresql",
		"host", c.Config.PgHost,
		"port", fmt.Sprintf("%d", c.Config.PgPort),
		"user", c.Config.PgUser,
		"password", c.Config.PgPassword,
		"database", c.Config.PgDbName)
	db, err := gorm.Open("postgres", psqlInfo)
	if err != nil {
		return ErrFailedToConnectToDabase
	}
	db.LogMode(false)
	// Migrate the schema
	db.AutoMigrate(&dbProspectInfo{})
	db.AutoMigrate(&dbProspect{})
	c.Db = db
	return err
}

func (c *Client) getInfoToIgnore(p *Prospect) []int {
	toIgnore := make([]int, 0)
	for i := 0; i < len(p.infos)-1; i++ {
		for j := i + 1; j < len(p.infos); j++ {
			if p.infos[i].Equal(p.infos[j]) {
				toIgnore = append(toIgnore, i)
			}
		}
	}
	return toIgnore
}

func (c *Client) Save(p *Prospect) error {
	p.prospect.ProspectID = c.getOrCreateProspectId(p.GetUrl())

	if err := c.saveDbProspect(p.prospect); err != nil {
		return err
	}

	duplicatedInfos := c.getInfoToIgnore(p)

	for i, info := range p.infos {
		if contains(duplicatedInfos, i) {
			continue
		}
		info.ProspectID = p.prospect.ProspectID
		if c.infoExist(info) {
			c.Logger.Info("prospect information already exists", "key", info.Key, "val", info.Val)
			continue
		}
		if err := c.Db.Create(&info).Error; err != nil {
			return err
		}
	}

	return nil
}

func contains(l []int, v int) bool {
	for _, li := range l {
		if li == v {
			return true
		}
	}
	return false
}

func (c *Client) Delete(prospectId string) (err error) {
	transaction := c.Db.Begin()
	if err = transaction.Model(&dbProspect{}).Delete(&dbProspect{}, "prospect_id = ?", prospectId).Error; err != nil {
		c.Logger.Warn(err.Error())
		transaction.Rollback()
		return
	}

	if err = transaction.Model(&dbProspectInfo{}).Delete(&dbProspect{}, "prospect_id = ?", prospectId).Error; err != nil {
		c.Logger.Warn(err.Error())
		transaction.Rollback()
		return
	}
	transaction.Commit()
	return
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
			prospect:   prospect,
			infos:      prospectsInfo,
		})
	}
	return
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
			Email:           info.Val,
			Confidence:      info.Confidence,
			ValidatedByUser: info.ValidatedByUser,
		})
	}
	return
}

func (c *Client) GetSocialMedia(prospectId string) (socialMedias []SocialMedia, err error) {

	for _, socialMedia := range allSocialMedia {
		infos := []dbProspectInfo{}

		if err = c.Db.Model(&dbProspectInfo{}).
			Where("prospect_id = ? AND key = ?", prospectId, socialMedia).
			Order("confidence desc").
			Limit(1).
			Find(&infos).Error; err != nil {
			c.Logger.Warn(err.Error())
			return
		}

		for _, info := range infos {
			socialMedias = append(socialMedias, SocialMedia{
				Name:            info.Key,
				Url:             info.Val,
				Confidence:      info.Confidence,
				ValidatedByUser: info.ValidatedByUser,
			})
		}
	}

	return
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

func (c *Client) GetTags(prospectId string) (tags []Tag, err error) {
	infos := []dbProspectInfo{}

	if err = c.Db.Model(&dbProspectInfo{}).
		Where("prospect_id = ? AND key = ?", prospectId, "tag").
		Find(&infos).Error; err != nil {
		c.Logger.Warn(err.Error())
		return
	}

	for _, info := range infos {
		tags = append(tags, Tag(info.Val))
	}
	return
}

func (c *Client) GetDescription(prospectId string) (desc string, err error) {
	infos := []dbProspectInfo{}

	if err = c.Db.Model(&dbProspectInfo{}).
		Where("prospect_id = ? AND key = ?", prospectId, "description").
		Limit(1).
		Find(&infos).Error; err != nil {
		c.Logger.Warn(err.Error())
		return
	}

	for _, info := range infos {
		return info.Val, nil
	}
	return
}

func (c *Client) saveDbProspect(p dbProspect) error {
	pro := dbProspect{}
	if notFound := c.Db.Model(&dbProspect{}).
		Where("url = ?", p.Url).Scan(&pro).RecordNotFound(); !notFound {
		return nil
	}

	if err := c.Db.Model(&dbProspect{}).Create(&p).Error; err != nil {
		return err
	}
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
		return true
	}
	return false
}

func (c *Client) getOrCreateProspectId(url string) string {
	var prospect dbProspect
	c.Db.Model(&dbProspect{}).
		Where("url = ?", url).First(&prospect)

	if prospect.Url != "" {
		return prospect.ProspectID
	}

	id, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}
	return id.String()
}
