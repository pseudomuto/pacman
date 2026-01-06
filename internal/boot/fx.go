package boot

import "go.uber.org/fx"

var Module = fx.Module("boot", fx.Provide(
	InitSumDBs,
))
