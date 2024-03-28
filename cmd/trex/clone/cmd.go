package clone

import (
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/golang/glog"
	"github.com/openshift-online/rh-trex/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type provisionCfgFlags struct {
	Name        string
	Destination string
}

func (c *provisionCfgFlags) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&c.Name, "name", c.Name, "Name of the new service being provisioned")
	fs.StringVar(&c.Destination, "destination", c.Destination, "Target directory for the newly provisioned instance")
}

var provisionCfg = &provisionCfgFlags{
	Name:        "clone-test",
	Destination: "/tmp",
}

// migrate sub-command handles running migrations
func NewCloneCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clone",
		Short: "Clone a new TRex instance",
		Long:  "Clone a new TRex instance",
		Run:   clone,
	}

	provisionCfg.AddFlags(cmd.PersistentFlags())
	cmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)
	return cmd
}

var rw os.FileMode = 0777

func clone(_ *cobra.Command, _ []string) {

	fullName := path.Join(provisionCfg.Destination, provisionCfg.Name)
	glog.Infof("creating new TRex instance as %s in directory %s", provisionCfg.Name, fullName)

	// walk the filesystem, starting at the root of the project
	err := filepath.Walk(".", func(currentPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// ignore git subdirectories
		if currentPath == ".git" || strings.Contains(currentPath, ".git/") {
			return nil
		}

		dest := path.Join(fullName, currentPath)
		dest = strings.Replace(dest, "trex", strings.ToLower(provisionCfg.Name), -1)

		if info.IsDir() {
			if _, err := os.Stat(dest); os.IsNotExist(err) {
				glog.Infof("Directory does not exist, creating: %s", dest)
			}
			return os.MkdirAll(dest, rw)

		}
		content, err := config.ReadFile(currentPath)
		if err != nil {
			return err
		}
		rhtrexRegx := regexp.MustCompile(`[Rr]?[Hh]?[-]?[Tt]rex`)
		content = rhtrexRegx.ReplaceAllString(content, provisionCfg.Name)
		os.WriteFile(dest, []byte(content), info.Mode().Perm())

		return nil
	})

	if err != nil {
		fmt.Println(err)
	}

}
