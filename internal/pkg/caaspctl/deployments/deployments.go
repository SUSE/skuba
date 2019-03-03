package deployments

type Target interface {
	Target() string
	Apply(states ...string) error
	UploadFileContents(contents, target string) error
	DownloadFileContents(source string) (string, error)
}
