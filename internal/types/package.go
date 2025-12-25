package types

type PackageOptions struct {
	Dir     string // Where on disk to find files.
	Package string // The go module, cargo package, etc.
	Version string // The version to be published.
}
