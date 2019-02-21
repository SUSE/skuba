package salt

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"path"
	"text/template"
)

const (
	MasterConfigTemplate = `log_level: info
root_dir: {{.WorkDir}}
cachedir: .salt/cache
ssh_log_file: .salt/logs/master
pki_dir: .salt/pki
file_roots:
  base:
    - {{.SaltPath}}
    - {{.WorkDir}}
`
)

type MasterConfig struct {
	SaltPath string
	WorkDir  string
	TempDir  string
}

func NewMasterConfig(saltPath string) MasterConfig {
	currDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	return MasterConfig{
		SaltPath: saltPath,
		WorkDir:  currDir,
	}
}

func (c *MasterConfig) GetTempDir() string {
	if len(c.TempDir) > 0 {
		return c.TempDir
	}

	template, err := template.New("masterConfig").Parse(MasterConfigTemplate)
	if err != nil {
		log.Fatal("could not parse internal master configuration")
	}

	var rendered bytes.Buffer
	if err := template.Execute(&rendered, c); err != nil {
		log.Fatal("could not render configuration")
	}

	c.TempDir, err = ioutil.TempDir("", "salt-master.conf.d")
	if err != nil {
		log.Fatal("could not create a temporary dir to save the master configuration")
	}

	masterConfig, err := os.Create(path.Join(c.TempDir, "master"))
	masterConfig.Write([]byte(rendered.String()))
	masterConfig.Close()

	return c.TempDir
}
