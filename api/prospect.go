package api

import (
	"net/http"
	"open-buzz/orm"
	"open-buzz/shared"
)

type ProspectHandler struct {
	Client interface {
		List() ([]orm.Prospect, error)
		GetEmails(prospectId string) ([]orm.Email, error)
		GetSocialMedia(prospectId string) (socialMedias []orm.SocialMedia, err error)
		GetAssets(prospectId string) (orm.Assets, error)
	} `inject:""`
	Logger shared.LoggerInterface `inject:""`
}

type JsonProspect struct {
	ProspectID string `json:"id"`
	Host string `json:"host"`
	Emails []JsonProspectEmail `json:"emails"`
	SocialMedia []JsonSocialMedia `json:"socialMedia"`
	Assets JsonAssets `json:"assets"`
}

type JsonProspectEmail struct {
	Email string `json:"email"`
	Confidence float64 `json:"confidence"`
	ValidatedByUser bool `json:"validatedByUser"`
}

type JsonSocialMedia struct {
	Name string `json:"name"`
	Link string `json:"link"`
	Confidence float64 `json:"confidence"`
	ValidatedByUser bool `json:"validatedByUser"`
}

type JsonAssets struct {
	Icons []JsonIcon `json:"icons"`
}

type JsonIcon struct {
	Link string `json:"link"`
}

func (c *ProspectHandler) List(w http.ResponseWriter, r *http.Request) {
	prospects, err := c.Client.List()
	if err != nil {
		writeError(w, err.Error())
		return
	}

	result := []JsonProspect{}

	for _, p := range prospects {
		emails, err := c.Client.GetEmails(p.ProspectId)
		if err != nil {
			c.Logger.Warn("unable to get emails for " + p.ProspectId, "err", err.Error())
			continue
		}

		socialMedia, err := c.Client.GetSocialMedia(p.ProspectId)
		if err != nil {
			c.Logger.Warn("unable to get social media for " + p.ProspectId, "err", err.Error())
			continue
		}

		assets, err := c.Client.GetAssets(p.ProspectId)
		if err != nil {
			c.Logger.Warn("unable to get social media for " + p.ProspectId, "err", err.Error())
			continue
		}

		result = append(result, JsonProspect{
			ProspectID: p.ProspectId,
			Host: p.GetUrl(),
			Emails: c.ormEmailsToJsonEmails(emails),
			SocialMedia: c.ormSocialMediaToJsonSocialMedia(socialMedia),
			Assets: c.ormAssetsToJsonAssets(assets),
		})
	}

	writeSuccess(w, result)

	return
}

func (c *ProspectHandler) ormEmailsToJsonEmails(emails []orm.Email) (jsonEmails []JsonProspectEmail){
	for _, email := range emails {
		jsonEmails = append(jsonEmails, JsonProspectEmail{
			Email: email.Email,
			ValidatedByUser: email.ValidatedByUser,
			Confidence: email.Confidence,
		})
	}
	return
}

func (c *ProspectHandler) ormSocialMediaToJsonSocialMedia(socialMedia []orm.SocialMedia) (jsonSocialMedia []JsonSocialMedia){
	for _, sm := range socialMedia {
		jsonSocialMedia = append(jsonSocialMedia, JsonSocialMedia{
			Name: sm.Name,
			Link: sm.Url,
			ValidatedByUser: sm.ValidatedByUser,
			Confidence: sm.Confidence,
		})
	}
	return
}

func (c *ProspectHandler) ormAssetsToJsonAssets(assets orm.Assets) (jsonAssets JsonAssets){
	for _, icon := range assets.Icons {
		jsonAssets.Icons = append(jsonAssets.Icons, JsonIcon{
			Link: icon.Link,
		})
	}
	return
}