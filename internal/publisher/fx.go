package publisher

import "go.uber.org/fx"

const (
	FXPackagers   = `group:"publisher_packagers"`
	FXUploaders   = `group:"publisher_uploaders"`
	FXVCSFetchers = `group:"publisher_vcs_fetchers"`
)

var Module = fx.Module("publisher", fx.Provide(New))
