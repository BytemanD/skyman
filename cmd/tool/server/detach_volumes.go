package server

import (
	"fmt"
	"runtime"

	"github.com/BytemanD/easygo/pkg/syncutils"
	"github.com/BytemanD/go-console/console"
	"github.com/BytemanD/skyman/common"
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

		client := common.DefaultClient()
		cinderClient := client.CinderV2()
		server, err := client.NovaV2().FindServer(args[0])
		utility.LogError(err, "show server failed:", true)
		if server.IsError() {
			utility.LogIfError(err, true, "server %s is Error", args[0])
		}
		attachedVolumes, err := client.NovaV2().ListServerVolumes(server.Id)
		utility.LogError(err, "list server interfaces failed:", true)

		console.Info("the server has %d volume(s)", len(attachedVolumes))

		detachVolumes := []string{}
		for i := len(attachedVolumes) - 1; i >= 0; i-- {
			console.Debug("found attached volumes: %s(%s)", attachedVolumes[i].VolumeId, attachedVolumes[i].Device)
			if len(detachVolumes) >= nums {
				break
			}
			if server.RootDeviceName != "" && attachedVolumes[i].Device == server.RootDeviceName {
				continue
			}
			detachVolumes = append(detachVolumes, attachedVolumes[i].VolumeId)
		}
		if len(detachVolumes) < nums {
			console.Fatal("the server only has %d volume(s) that can be detached", len(detachVolumes))
		}
		if len(detachVolumes) == 0 {
			console.Warn("nothing to do")
			return
		}
		taskGroup := syncutils.TaskGroup[string]{
			Items:        detachVolumes,
			MaxWorker:    parallel,
			ShowProgress: true,
			Func: func(p string) error {
				console.Info("[volume: %s] request to detach", p)
				err := client.NovaV2().DeleteServerVolumeAndWait(server.Id, p, 600)
				if err != nil {
					console.Error("[volume: %s] detach failed: %v", p, err)
					return err
				}
				console.Info("[volume: %s] detached", p)
				return nil
			},
		}
		console.Info("detaching %d volume(s) ...", len(detachVolumes))
		taskGroup.Start()
		if !clean {
			return
		}
		taskGroup2 := syncutils.TaskGroup[string]{
			Items:        detachVolumes,
			MaxWorker:    parallel,
			ShowProgress: true,
			Func: func(p string) error {
				console.Debug("[volume: %s] deleting", p)
				err := cinderClient.DeleteVolume(p, true, true)
				// TODO: wait deleted
				if err != nil {
					console.Error("[volume: %s] delete failed: %v", p, err)
					return err
				}
				console.Info("[volume: %s] deleted", p)
				return nil
			},
		}
		console.Info("cleaning ...")
		taskGroup2.Start()
	},
}

func init() {
	detachVolumes.Flags().Int("nums", 1, "nums of interfaces")
	detachVolumes.Flags().Int("parallel", runtime.NumCPU(), "nums of parallel")
	detachVolumes.Flags().Bool("clean", false, "delete volumes after detached")

	ServerCommand.AddCommand(detachVolumes)
}
