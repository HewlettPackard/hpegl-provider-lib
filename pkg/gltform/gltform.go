// (C) Copyright 2021-2023 Hewlett Packard Enterprise Development LP

package gltform

import (
	"io"
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"
)

const fileExtension = ".gltform"

// Gljwt - the contents of the .gltform file
type Gljwt struct {
	// SpaceName is optional, and is only required for metal if we want to create a project
	SpaceName string `yaml:"space_name,omitempty"`
	// ProjectID - the metal/Quake project ID
	ProjectID string `yaml:"project_id"`
	// RestURL - the URL to be used for metal, at present it refers to a Quake portal URL
	RestURL string `yaml:"rest_url"`
	// TODO remove this entry once we've switched quake-client over to using this package
	// Token - the GL IAM token
	Token string `yaml:"access_token,omitempty"`
}

// GetGLConfig - reads the .gltform file, note that the .gltform can be in the home directory of the
// user running terraform, or in the directory from which terraform is run
func GetGLConfig() (gljwt *Gljwt, err error) {
	homeDir, _ := os.UserHomeDir()
	workingDir, _ := os.Getwd()
	for _, p := range []string{workingDir, homeDir} {
		gljwt, err = loadGLConfig(p)
		if err == nil {
			break
		}
	}

	return gljwt, err
}

// WriteGLConfig takes a map[string]interface{} which will normally come from a
// service block in the provider stanza and writes out a .gltform file in the directory
// from which terraform is being run.  See the use of this function
// for metal in terraform-provider-hpegl.
func WriteGLConfig(d map[string]interface{}) error {
	config := &Gljwt{
		// If space_name isn't present, we'll just write out ""
		SpaceName: d["space_name"].(string),
		ProjectID: d["project_id"].(string),
		RestURL:   d["rest_url"].(string),
	}

	// Marshal config
	b, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	// Write out marshalled config into .gltform
	workingDir, _ := os.Getwd()

	for _, p := range []string{workingDir} {
		err = writeGLConfigToFile(b, p)
		if err != nil {
			break
		}
	}

	return err
}

func writeGLConfigToFile(b []byte, dir string) error {
	f, err := os.Create(filepath.Clean(filepath.Join(dir, fileExtension)))
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.Write(b)
	if err != nil {
		return err
	}

	return f.Sync()
}

func loadGLConfig(dir string) (*Gljwt, error) {
	f, err := os.Open(filepath.Clean(filepath.Join(dir, fileExtension)))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return parseGLStream(f)
}

func parseGLStream(s io.Reader) (*Gljwt, error) {
	contents, err := io.ReadAll(s)
	if err != nil {
		return nil, err
	}

	q := &Gljwt{}
	err = yaml.Unmarshal(contents, q)

	return q, err
}
