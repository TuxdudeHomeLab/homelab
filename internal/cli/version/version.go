package version

type VersionInfo struct {
	PackageVersion   string
	PackageCommit    string
	PackageTimestamp string
}

func NewVersionInfo(pkgVersion, pkgCommit, pkgTimestamp string) *VersionInfo {
	return &VersionInfo{
		PackageVersion:   pkgVersion,
		PackageCommit:    pkgCommit,
		PackageTimestamp: pkgTimestamp,
	}
}
