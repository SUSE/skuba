package deployments

type Target interface {
	Target() string
	Apply(states ...string) error
	UploadFile(sourcePath, targetPath string) error
	UploadFileContents(targetPath, contents string) error
	DownloadFileContents(sourcePath string) (string, error)
}
