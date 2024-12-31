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
		nums, _ := cmd.Flags().GetInt("nums")
		parallel, _ := cmd.Flags().GetInt("parallel")
		clean, _ := cmd.Flags().GetBool("clean")

		client := openstack.DefaultClient()
		cinderClient := client.CinderV2()
		server, err := client.NovaV2().Server().Find(args[0])
		utility.LogError(err, "show server failed:", true)
		if server.IsError() {
			utility.LogIfError(err, true, "server %s is Error", args[0])
		}
		attachedVolumes, err := client.NovaV2().Server().ListVolumes(server.Id)
		utility.LogError(err, "list server interfaces failed:", true)

		logging.Info("the server has %d volume(s)", len(attachedVolumes))

		detachVolumes := []string{}
		for i := len(attachedVolumes) - 1; i >= 0; i-- {
			logging.Debug("found attached volumes: %s(%s)", attachedVolumes[i].VolumeId, attachedVolumes[i].Device)
			if len(detachVolumes) >= nums {
				break
			}
			if server.RootDeviceName != "" && attachedVolumes[i].Device == server.RootDeviceName {
				continue
			}
			detachVolumes = append(detachVolumes, attachedVolumes[i].VolumeId)
		}
		if len(detachVolumes) < nums {
			logging.Error("the server only has %d volume(s) that can be detached", len(detachVolumes))
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
				err := client.NovaV2().Server().DeleteVolumeAndWait(server.Id, p, 600)

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
				err := cinderClient.Volume().Delete(p, true, true)
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
	detachVolumes.Flags().Bool("clean", false, "delete volumes after detached")

	ServerCommand.AddCommand(detachVolumes)
}
