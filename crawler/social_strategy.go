package crawler

type SocialStrategy interface {
	GetUrlPrefix() string
	GetName() string
}

func GetAllSocialStrategies() []SocialStrategy {
	return []SocialStrategy{
		&TwitterStrategy{},
		&FacebookStrategy{},
		&YoutubeStrategy{},
		&LinkedinStrategy{},
		&LinkedinCompanyStrategy{},
	}
}

type TwitterStrategy struct{}

func (s *TwitterStrategy) GetUrlPrefix() string {
	return "twitter.com/"
}

func (s *TwitterStrategy) GetName() string {
	return "twitter"
}

type FacebookStrategy struct{}

func (s *FacebookStrategy) GetUrlPrefix() string {
	return "facebook.com/"
}

func (s *FacebookStrategy) GetName() string {
	return "facebook"
}

type YoutubeStrategy struct{}

func (s *YoutubeStrategy) GetUrlPrefix() string {
	return "youtube.com/"
}

func (s *YoutubeStrategy) GetName() string {
	return "youtube"
}

type LinkedinStrategy struct{}

func (s *LinkedinStrategy) GetUrlPrefix() string {
	return "linkedin.com/in/"
}

func (s *LinkedinStrategy) GetName() string {
	return "linkedin"
}

type LinkedinCompanyStrategy struct{}

func (s *LinkedinCompanyStrategy) GetUrlPrefix() string {
	return "linkedin.com/company/"
}

func (s *LinkedinCompanyStrategy) GetName() string {
	return "linkedin"
}
