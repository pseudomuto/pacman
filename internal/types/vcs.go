package types

const (
	GitLab VCSType = iota
	GitHub VCSType = iota
)

type (
	VCSType int8

	VCSOptions struct {
		Ref string // SHA, branch, or tag
		Dir string // Directory within the repo to fetch.
	}
)

func (v VCSType) String() string {
	if v == GitHub {
		return "GitHub"
	}

	return "GitLab"
}
