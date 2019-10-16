package orcconfigdir

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"strings"

	"github.com/steinarvk/orc"
	"github.com/steinarvk/orclib/lib/versioninfo"

	homedir "github.com/mitchellh/go-homedir"
)

var (
	defaultConfigDirTemplates = []string{
		"~/.config/{{ .ProgramName }}",
		"~/.{{ .ProgramName }}",
		"/etc/{{ .ProgramName }}/",
	}
)

type Module struct {
	acceptablePaths []string
}

func (m *Module) GetConfigPath(name string) (string, bool) {
	for _, path := range m.acceptablePaths {
		info, err := os.Stat(path)
		if os.IsNotExist(err) {
			continue
		}

		if !info.IsDir() {
			continue
		}

		return path, true
	}

	return "", false
}

func (m *Module) ReadConfig(name string) ([]byte, error) {
	path, ok := m.GetConfigPath(name)
	if !ok {
		return nil, fmt.Errorf("no suitable config dir found")
	}

	return ioutil.ReadFile(path)
}

func (m *Module) GetWritableConfigDir() (string, error) {
	if len(m.acceptablePaths) == 0 {
		return "", fmt.Errorf("no config dir set")
	}

	best := m.acceptablePaths[0]

	if err := ensureDirExists(best); err != nil {
		return "", err
	}

	return best, nil
}

func (m *Module) ModuleName() string { return "ConfigDir" }

var M = &Module{}

func ensureDirExists(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			return fmt.Errorf("unable to create config dir %q: %v", path, err)
		}
		info, err = os.Stat(path)
	}

	if err != nil {
		return err
	}

	if !info.IsDir() {
		return fmt.Errorf("desired config dir %q exists but is not a directory", path)
	}

	return nil
}

func expandPath(path string) (string, error) {
	if strings.HasPrefix(path, "~") {
		expanded, err := homedir.Expand(path)
		if err != nil {
			return "", fmt.Errorf("failed to expand homedir in %q: %v", path, err)
		}
		return expanded, nil
	}
	return path, nil
}

func (m *Module) OnRegister(hooks orc.ModuleHooks) {
	var flagConfigDir string

	hooks.OnUse(func(ctx orc.UseContext) {
		ctx.Flags.StringVar(&flagConfigDir, "config_dir", "", "config file dir")
	})

	hooks.OnSetup(func() error {
		hasExplicitConfig := flagConfigDir != ""

		if hasExplicitConfig {
			expanded, err := expandPath(flagConfigDir)
			if err != nil {
				return err
			}

			if err := ensureDirExists(expanded); err != nil {
				return err
			}

			m.acceptablePaths = []string{expanded}

			return nil
		}

		programName := versioninfo.ProgramName

		params := struct {
			ProgramName string
		}{
			ProgramName: programName,
		}

		for _, tmpl := range defaultConfigDirTemplates {
			compiled, err := template.New("").Parse(tmpl)
			if err != nil {
				return fmt.Errorf("unable to compile config dir template %q: %v", tmpl, err)
			}

			var buf bytes.Buffer

			if err := compiled.Execute(&buf, params); err != nil {
				return fmt.Errorf("unable to execute config dir template %q: %v", tmpl, err)
			}

			x := buf.String()

			expanded, err := expandPath(x)
			if err != nil {
				return err
			}

			m.acceptablePaths = append(m.acceptablePaths, expanded)
		}

		return nil
	})
}
