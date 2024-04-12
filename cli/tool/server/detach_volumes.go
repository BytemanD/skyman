package server

import (
	"fmt"
	"os"
	"runtime"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/easygo/pkg/syncutils"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/utility"
	"github.com/spf13/cobra"
)

var detachVolumes = &cobra.Command{
	Use:   "remove-volumes <server>",
	Short: "Remove volumes from server",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}
		nums, _ := cmd.Flags().GetInt("nums")
		if nums <= 0 {
			return fmt.Errorf("invalid flag --nums, it must >= 1")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		serverId := args[0]

		nums, _ := cmd.Flags().GetInt("nums")
		parallel, _ := cmd.Flags().GetInt("parallel")
		clean, _ := cmd.Flags().GetBool("clean")

		client := openstack.DefaultClient()
		cinderClient := client.CinderV2()
		server, err := client.NovaV2().Servers().Show(serverId)
		utility.LogError(err, "show server failed:", true)

		attachedVolumes, err := client.NovaV2().Servers().ListVolumes(server.Id)
		utility.LogError(err, "list server interfaces failed:", true)

		logging.Info("the server has %d volume(s)", len(attachedVolumes))

		detachVolumes := []string{}
		for i := len(attachedVolumes) - 1; i >= 0; i-- {
			if len(detachVolumes) >= nums {
				break
			}
			if attachedVolumes[i].Device == server.RootDeviceName {
				continue
			}
			detachVolumes = append(detachVolumes, attachedVolumes[i].VolumeId)
		}
		if len(detachVolumes) < nums {
			logging.Info("the server only has %d volume(s) that can be detached", len(detachVolumes))
			os.Exit(1)
		}
		if len(detachVolumes) == 0 {
			logging.Warning("nothing to do")
			return
		}
		taskGroup := syncutils.TaskGroup{
			Items:        detachVolumes,
			MaxWorker:    parallel,
			ShowProgress: true,
			Func: func(item interface{}) error {
				p := item.(string)
				logging.Debug("[volume: %s] detaching", p)
				err := client.NovaV2().Servers().DeleteVolumeAndWait(server.Id, p, 600)

				if err != nil {
					logging.Error("[volume: %s] detach failed: %v", p, err)
					return err
				}
				logging.Info("[volume: %s] detached", p)
				return nil
			},
		}
		logging.Info("detaching %d volume(s) ...", len(detachVolumes))
		taskGroup.Start()
		if !clean {
			return
		}
		taskGroup2 := syncutils.TaskGroup{
			Items:        detachVolumes,
			MaxWorker:    parallel,
			ShowProgress: true,
			Func: func(item interface{}) error {
				p := item.(string)
				logging.Debug("[volume: %s] deleting", p)
				err := cinderClient.Volumes().Delete(p)
				// TODO: wait deleted
				if err != nil {
					logging.Error("[volume: %s] delete failed: %v", p, err)
					return err
				}
				logging.Info("[volume: %s] deleted", p)
				return nil
			},
		}
		logging.Info("cleaning ...")
		taskGroup2.Start()
	},
}

func init() {
	detachVolumes.Flags().Int("nums", 1, "nums of interfaces")
	detachVolumes.Flags().Int("parallel", runtime.NumCPU(), "nums of parallel")
	detachVolumes.Flags().Bool("clean", false, "delete interface after detached")

	ServerCommand.AddCommand(detachVolumes)
}
