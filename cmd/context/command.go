package context

import (
	"os"
	"path/filepath"

	"github.com/BytemanD/go-console/console"
	"github.com/spf13/cobra"
	"github.com/wxnacy/wgo/file"
	"gopkg.in/yaml.v3"
)

func LoadContextConf() (*ContextConf, error) {
	filePath := getContextFilePath()
	console.Debug("context file path: %s", filePath)
	conf := ContextConf{
		filePath: filePath,
	}
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &conf, nil
		} else {
			return nil, err
		}
	}
	if err := yaml.Unmarshal(data, &conf); err != nil {
		return nil, err
	}
	return &conf, nil
}

var ContextCmd = &cobra.Command{
	Use: "context",
}
var viewCmd = &cobra.Command{
	Use:   "view",
	Short: "Display context",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		cConf, err := LoadContextConf()
		if err != nil {
			console.Fatal("load context failed: %s", err)
		}
		data, _ := yaml.Marshal(cConf)
		println(string(data))
	},
}

var setCmd = &cobra.Command{
	Use:   "set <name> <conf file>",
	Short: "Set context",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		confPathAbs, err := filepath.Abs(args[1])
		if err != nil {
			console.Fatal("get '%s' abs path failed: %s", args[1], err)
		}
		if !file.IsFile(confPathAbs) {
			console.Fatal("%s is not a file or not exits", confPathAbs)
		}

		cConf, err := LoadContextConf()
		if err != nil {
			console.Fatal("load context failed: %s", err)
		}
		cConf.SetContext(args[0], confPathAbs)
		if err := cConf.Save(); err != nil {
			console.Fatal("save context failed: %s", err)
		}
	},
}

var removeCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove context",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cConf, err := LoadContextConf()
		if err != nil {
			console.Fatal("load context failed: %s", err)

		}
		cConf.RemoveContext(args[0])
		if err := cConf.Save(); err != nil {
			console.Fatal("remove context failed: %s", err)
		}
		cConf.Save()
	},
}
var useCmd = &cobra.Command{
	Use:   "use <name>",
	Short: "Use context",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		err := UseCluster(args[0])
		if err != nil {
			console.Fatal("use %s failed: %s", args[0], err)
		}
	},
}
var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset context",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		cConf, err := LoadContextConf()
		if err != nil {
			console.Fatal("load context failed: %s", err)

		}
		cConf.Reset()
		if err := cConf.Save(); err != nil {
			console.Fatal("reset context failed: %s", err)

		}
	},
}

func init() {
	ContextCmd.AddCommand(viewCmd, useCmd, resetCmd, setCmd, removeCmd)
}
