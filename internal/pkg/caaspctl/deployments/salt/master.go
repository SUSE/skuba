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

	RosterTemplate = `{{.Target.Node}}:
  host: {{.Target.Node}}
  user: {{.Target.User}}
  sudo: {{.Target.Sudo}}
  tty: True
`
)

type MasterConfig struct {
	SaltPath string
	WorkDir  string
	TempDir  string
	Target   Target
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

func (c *MasterConfig) GetTempDir(target Target) string {
	if len(c.TempDir) > 0 && c.Target == target {
		return c.TempDir
	}

	c.Target = target

	masterConfigTemplate, err := template.New("masterConfig").Parse(MasterConfigTemplate)
	if err != nil {
		log.Fatal("could not parse internal master configuration")
	}

	var masterConfigRendered bytes.Buffer
	if err := masterConfigTemplate.Execute(&masterConfigRendered, c); err != nil {
		log.Fatal("could not render configuration")
	}

	rosterTemplate, err := template.New("roster").Parse(RosterTemplate)
	if err != nil {
		log.Fatal("could not parse internal roster")
	}

	var rosterRendered bytes.Buffer
	if err := rosterTemplate.Execute(&rosterRendered, c); err != nil {
		log.Fatal("could not render roster")
	}

	c.TempDir, err = ioutil.TempDir("", "salt-master.conf.d")
	if err != nil {
		log.Fatal("could not create a temporary dir to save the master configuration")
	}

	masterConfig, err := os.Create(path.Join(c.TempDir, "master"))
	masterConfig.Write([]byte(masterConfigRendered.String()))
	masterConfig.Close()

	roster, err := os.Create(path.Join(c.TempDir, "roster"))
	roster.Write([]byte(rosterRendered.String()))
	roster.Close()

	return c.TempDir
}
