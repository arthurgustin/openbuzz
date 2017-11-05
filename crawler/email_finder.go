package crawler

import (
	"github.com/arthurgustin/openbuzz/orm"
	"github.com/arthurgustin/openbuzz/shared"
	"github.com/badoux/checkmail"
	"github.com/bobesa/go-domain-util/domainutil"
	"github.com/pkg/errors"
	"strings"
)

type EmailFinder struct {
	Logger shared.LoggerInterface `inject:""`
	Config *shared.AppConfig      `inject:""`
}

var (
	ErrAllPolicyActivated = errors.New("All policy activated on this host")
)

func (f *EmailFinder) Find(prospect orm.Prospect) ([]Mail, error) {
	allPolicyActivated, err := f.isAllPolicyActivated(prospect)
	if err != nil {
		f.Logger.Info(err.Error())
	}
	if allPolicyActivated {
		f.Logger.Warn(ErrAllPolicyActivated.Error(), "domain", domainutil.Domain(prospect.GetUrl()))
		return []Mail{}, ErrAllPolicyActivated
	}

	mails := []Mail{}
	for _, mail := range f.generatePossibleMails(prospect) {
		isReachable, err := mail.isReachable()
		if err != nil {
			f.Logger.Info("not reachable", "email", mail.email, "err", err.Error())
		}
		if isReachable {
			mails = append(mails, mail)
		}
	}

	return mails, nil
}

func (f *EmailFinder) isAllPolicyActivated(prospect orm.Prospect) (bool, error) {
	m := Mail{
		email: "all_policy_activated@" + domainutil.Domain(prospect.GetUrl()),
	}
	return m.isReachable()
}

func (f *EmailFinder) generatePossibleMails(prospect orm.Prospect) []Mail {
	mails := []Mail{}

	for _, prefix := range f.getMailsPrefix(prospect) {
		for _, suffix := range f.getMailSuffix(prospect) {
			mails = append(mails, Mail{
				email: prefix + "@" + suffix,
			})
		}
	}

	commonWebsitePrefix := []string{"contact", "blog", "info", "infos", "admin", "support"}
	for _, commonPrefix := range commonWebsitePrefix {
		mails = append(mails, Mail{
			email: commonPrefix + "@" + domainutil.Domain(prospect.GetUrl()),
		})
	}

	return mails
}

func (f *EmailFinder) getMailSuffix(prospect orm.Prospect) []string {
	domain := domainutil.Domain(prospect.GetUrl())

	return []string{
		domain, "google.com",
	}
}

var (
	permutations = []string{
		"{fn}",
		"{ln}",
		"{fn}{ln}",
		"{fn}.{ln}",
		"{fi}{ln}",
		"{fi}.{ln}",
		"{fn}{li}",
		"{fn}.{li}",
		"{fi}{li}",
		"{fi}.{li}",
		"{ln}{fn}",
		"{ln}.{fn}",
		"{ln}{fi}",
		"{ln}.{fi}",
		"{li}{fn}",
		"{li}.{fn}",
		"{li}{fi}",
		"{li}.{fi}",
		"{fi}{mi}{ln}",
		"{fi}{mi}.{ln}",
		"{fn}{mi}{ln}",
		"{fn}.{mi}.{ln}",
		"{fn}{mn}{ln}",
		"{fn}.{mn}.{ln}",
		"{fn}-{ln}",
		"{fi}-{ln}",
		"{fn}-{li}",
		"{fi}-{li}",
		"{ln}-{fn}",
		"{ln}-{fi}",
		"{li}-{fn}",
		"{li}-{fi}",
		"{fi}{mi}-{ln}",
		"{fn}-{mi}-{ln}",
		"{fn}-{mn}-{ln}",
		"{fn}_{ln}",
		"{fi}_{ln}",
		"{fn}_{li}",
		"{fi}_{li}",
		"{ln}_{fn}",
		"{ln}_{fi}",
		"{li}_{fn}",
		"{li}_{fi}",
		"{fi}{mi}_{ln}",
		"{fn}_{mi}_{ln}",
		"{fn}_{mn}_{ln}",
	}
)

func (f *EmailFinder) getMailsPrefix(prospect orm.Prospect) []string {
	var fn, fi, mn, mi, ln, li string

	fn = prospect.GetFirstName()
	if len(fn) > 0 {
		fi = fn[:1]
	}
	mn = prospect.GetMiddleName()
	if len(mn) > 0 {
		mi = mn[:1]
	}
	ln = prospect.GetLastName()
	if len(ln) > 0 {
		li = ln[:1]
	}

	permutationsMails := []string{}

	replacer := strings.NewReplacer(
		"{li}", li,
		"{fn}", fn,
		"{ln}", ln,
		"{mn}", mn,
		"{fi}", fi,
		"{li}", li,
		"{mi}", mi,
	)

	for _, perm := range permutations {
		mail := replacer.Replace(perm)
		if f.getUniqueCharNumber(mail) > 1 {
			permutationsMails = append(permutationsMails, replacer.Replace(perm))
		}
	}

	return permutationsMails
}

func (f *EmailFinder) getUniqueCharNumber(s string) (n int) {
	lookup := map[rune]int{}

	for _, char := range s {
		lookup[char] += 1
	}
	return len(lookup)
}

type Mail struct {
	email string
}

func (m *Mail) isReachable() (bool, error) {
	err := checkmail.ValidateHost(m.email)
	if smtpErr, ok := err.(checkmail.SmtpError); ok && err != nil {
		return false, smtpErr.Err
	}
	return true, nil
}
