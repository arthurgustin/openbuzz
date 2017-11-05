package orm

import (
	"github.com/bobesa/go-domain-util/domainutil"
	"github.com/jinzhu/gorm"
	"strings"
)

func NewProspect(url string) *Prospect {
	prospect := &Prospect{}
	return prospect.SetUrl(url)
}

type Prospect struct {
	ProspectId string
	prospect   dbProspect
	infos      []dbProspectInfo
}

type dbProspect struct {
	gorm.Model
	ProspectID string `gorm:"not null;unique"`
	Url        string `gorm:"not null"`
	FirstName  string
	MiddleName string
	LastName   string
}

type dbProspectInfo struct {
	gorm.Model
	ProspectID      string `gorm:"not null"`
	Key             string
	Val             string
	Confidence      float64 // [0 - 1]
	ValidatedByUser bool
}

func (i dbProspectInfo) Equal(j dbProspectInfo) bool {
	return i.Key == j.Key && i.ProspectID == j.ProspectID && i.Val == j.Val && i.Confidence == j.Confidence && i.ValidatedByUser == j.ValidatedByUser
}

func (p *Prospect) SetIcon(targetUrl string) *Prospect {
	if strings.HasPrefix(targetUrl, "/") {
		targetUrl = p.GetBaseUrl() + targetUrl
	}
	return p.addInfo("icon", targetUrl, 1)
}

func (p *Prospect) SetTag(tag string) *Prospect {
	return p.addInfo("tag", tag, 1)
}

func (p *Prospect) SetDescription(description string) *Prospect {
	return p.addInfo("description", description, 1)
}

func (p *Prospect) SetUrl(targetUrl string) *Prospect {
	p.prospect.Url = targetUrl
	return p.addInfo("domain", targetUrl, 1)
}

func (p *Prospect) GetUrl() string {
	return p.prospect.Url
}

func (p *Prospect) GetUrlPrefix() string {
	return strings.Split(p.prospect.Url, "://")[0]
}

func (p *Prospect) GetDomainNameWithoutExtension() string {
	// e.g for korben.info, returns korben
	return strings.Split(p.GetDomain(), ".")[0]
}

func (p *Prospect) GetBaseUrl() string {
	// http or https
	urlPrefix := strings.Split(p.GetUrl(), "://")[0]

	return urlPrefix + "://" + p.GetDomain()
}

func (p *Prospect) GetDomain() string {
	domain := domainutil.Domain(p.GetUrl())
	return domain
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
		ProspectID:      p.prospect.ProspectID,
		Key:             key,
		Val:             val,
		ValidatedByUser: false,
		Confidence:      confidence,
	})
	return p
}

type Email struct {
	Email           string
	Confidence      float64
	ValidatedByUser bool
}

type Assets struct {
	Icons []Icon
}

type Icon struct {
	Link string
}

type Tag string

type SocialMedia struct {
	Name            string
	Url             string
	Confidence      float64
	ValidatedByUser bool
}
