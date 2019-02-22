package init

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"text/template"
)

type InitConfiguration struct {
	ProjectName          string
	ControlPlaneEndpoint string
}

func Init(initConfiguration InitConfiguration) {
	if err := os.MkdirAll(initConfiguration.ProjectName, 0700); err != nil {
		log.Fatalf("Could not create directory %s\n", initConfiguration.ProjectName)
	}
	if err := os.Chdir(initConfiguration.ProjectName); err != nil {
		log.Fatalf("Could not change to directory %s\n", initConfiguration.ProjectName)
	}
	for _, file := range scaffoldFiles {
		filePath, _ := filepath.Split(file.Location)
		if filePath != "" {
			if err := os.MkdirAll(filePath, 0700); err != nil {
				log.Fatalf("Could not create directory %s\n", filePath)
			}
		}
		f, err := os.Create(file.Location)
		if err != nil {
			log.Fatalf("Could not create file %s\n", file.Location)
		}
		f.WriteString(renderTemplate(file.Content, initConfiguration))
		f.Close()
	}
}

func renderTemplate(templateContents string, initConfiguration InitConfiguration) string {
	template, err := template.New("").Parse(templateContents)
	if err != nil {
		log.Fatal("could not parse template")
	}
	var rendered bytes.Buffer
	if err := template.Execute(&rendered, initConfiguration); err != nil {
		log.Fatal("could not render configuration")
	}
	return rendered.String()
}
