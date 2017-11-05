package api

import (
	"github.com/arthurgustin/openbuzz/orm"
	"github.com/arthurgustin/openbuzz/shared"
	"net/http"
)

type ProspectHandler struct {
	Client interface {
		List() ([]orm.Prospect, error)
		GetEmails(prospectId string) ([]orm.Email, error)
		GetSocialMedia(prospectId string) (socialMedias []orm.SocialMedia, err error)
		GetAssets(prospectId string) (orm.Assets, error)
		GetTags(prospectId string) ([]orm.Tag, error)
		GetDescription(prospectId string) (string, error)
	} `inject:""`
	Logger shared.LoggerInterface `inject:""`
}

type JsonProspect struct {
	ProspectID  string              `json:"id"`
	Host        string              `json:"host"`
	Description string              `json:"description"`
	Emails      []JsonProspectEmail `json:"emails"`
	SocialMedia []JsonSocialMedia   `json:"socialMedia"`
	Assets      JsonAssets          `json:"assets"`
	Tags        []JsonTag           `json:"tags"`
}

type JsonProspectEmail struct {
	Email           string  `json:"email"`
	Confidence      float64 `json:"confidence"`
	ValidatedByUser bool    `json:"validatedByUser"`
}

type JsonSocialMedia struct {
	Name            string  `json:"name"`
	Link            string  `json:"link"`
	Confidence      float64 `json:"confidence"`
	ValidatedByUser bool    `json:"validatedByUser"`
}

type JsonAssets struct {
	Icons []JsonIcon `json:"icons"`
}

type JsonIcon struct {
	Link string `json:"link"`
}

type JsonTag string

type Response struct {
	Prospects []JsonProspect `json:"prospects"`
	Error     bool           `json:"error"`
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
			c.Logger.Warn("unable to get emails for "+p.ProspectId, "err", err.Error())
			continue
		}

		socialMedia, err := c.Client.GetSocialMedia(p.ProspectId)
		if err != nil {
			c.Logger.Warn("unable to get social media for "+p.ProspectId, "err", err.Error())
			continue
		}

		assets, err := c.Client.GetAssets(p.ProspectId)
		if err != nil {
			c.Logger.Warn("unable to get social media for "+p.ProspectId, "err", err.Error())
			continue
		}

		tags, err := c.Client.GetTags(p.ProspectId)
		if err != nil {
			c.Logger.Warn("unable to get tags for "+p.ProspectId, "err", err.Error())
			continue
		}

		description, err := c.Client.GetDescription(p.ProspectId)
		if err != nil {
			c.Logger.Warn("unable to get description for "+p.ProspectId, "err", err.Error())
			continue
		}

		result = append(result, JsonProspect{
			ProspectID:  p.ProspectId,
			Host:        p.GetUrl(),
			Description: description,
			Emails:      c.ormEmailsToJsonEmails(emails),
			SocialMedia: c.ormSocialMediaToJsonSocialMedia(socialMedia),
			Assets:      c.ormAssetsToJsonAssets(assets),
			Tags:        c.ormTagsToJsonTags(tags),
		})
	}

	resp := Response{
		Error:     false,
		Prospects: result,
	}

	writeSuccess(w, resp)

	return
}

func (c *ProspectHandler) ormEmailsToJsonEmails(emails []orm.Email) (jsonEmails []JsonProspectEmail) {
	for _, email := range emails {
		jsonEmails = append(jsonEmails, JsonProspectEmail{
			Email:           email.Email,
			ValidatedByUser: email.ValidatedByUser,
			Confidence:      email.Confidence,
		})
	}
	return
}

func (c *ProspectHandler) ormSocialMediaToJsonSocialMedia(socialMedia []orm.SocialMedia) (jsonSocialMedia []JsonSocialMedia) {
	for _, sm := range socialMedia {
		jsonSocialMedia = append(jsonSocialMedia, JsonSocialMedia{
			Name:            sm.Name,
			Link:            sm.Url,
			ValidatedByUser: sm.ValidatedByUser,
			Confidence:      sm.Confidence,
		})
	}
	return
}

func (c *ProspectHandler) ormAssetsToJsonAssets(assets orm.Assets) (jsonAssets JsonAssets) {
	for _, icon := range assets.Icons {
		jsonAssets.Icons = append(jsonAssets.Icons, JsonIcon{
			Link: icon.Link,
		})
	}
	return
}

func (c *ProspectHandler) ormTagsToJsonTags(tags []orm.Tag) (jsonTags []JsonTag) {
	for _, tag := range tags {
		jsonTags = append(jsonTags, JsonTag(tag))
	}
	return
}
