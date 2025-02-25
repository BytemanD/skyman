package server

import (
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/BytemanD/go-console/console"
	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"github.com/BytemanD/easygo/pkg/syncutils"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/utility"
)

var attachVolume = &cobra.Command{
	Use:   "add-volumes <server>",
	Short: "Add volumes to server",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		nums, _ := cmd.Flags().GetInt("nums")
		parallel, _ := cmd.Flags().GetInt("parallel")
		size, _ := cmd.Flags().GetInt("size")
		volumeType, _ := cmd.Flags().GetString("type")

		client := common.DefaultClient()
		cinderClient := client.CinderV2()
		server, err := client.NovaV2().Server().Find(args[0])
		utility.LogError(err, "show server failed:", true)
		if server.IsError() {
			utility.LogIfError(err, true, "server %s is Error", args[0])
		}

		volumes := []Volume{}
		mu := sync.Mutex{}

		taskGroup := syncutils.TaskGroup{
			Items:        lo.Range(nums),
			MaxWorker:    parallel,
			Title:        fmt.Sprintf("create %d volume(s)", len(volumes)),
			ShowProgress: true,
			Func: func(item interface{}) error {
				p := item.(int)
				name := fmt.Sprintf("skyman-volume-%d", p+1)
				createOption := map[string]interface{}{
					"name": name, "size": size,
				}
				if volumeType != "" {
					createOption["volume_type"] = volumeType
				}
				console.Debug("creating volume %s", name)
				volume, err := cinderClient.Volume().CreateAndWait(createOption, 600)
				if err != nil {
					console.Error("create volume failed: %v", err)
					return err
				}
				console.Info("created volume: %v (%v)", volume.Name, volume.Id)
				mu.Lock()
				volumes = append(volumes, Volume{Id: volume.Id, Name: name})
				mu.Unlock()
				return nil
			},
		}
		console.Info("creating %d volume(s), waiting ...", nums)
		taskGroup.Start()

		if len(volumes) == 0 {
			return
		}
		taskGroup2 := syncutils.TaskGroup{
			Items:        volumes,
			Title:        fmt.Sprintf("attach %d volume(s)", len(volumes)),
			MaxWorker:    parallel,
			ShowProgress: true,
			Func: func(item interface{}) error {
				p := item.(Volume)
				console.Debug("[volume: %s] attaching", p)
				attachment, err := client.NovaV2().Server().AddVolume(server.Id, p.Id)
				if err != nil {
					console.Error("[volume: %s] attach failed: %v", p, err)
					return err
				}
				if attachment != nil && p.Id == "" {
					p.Id = attachment.VolumeId
				}
				startTime := time.Now()
				for {
					attachedVolume, err := client.NovaV2().Server().ListVolumes(server.Id)
					if err != nil {
						utility.LogError(err, "list server volumes failed:", false)
						return err
					}
					for _, vol := range attachedVolume {
						if vol.VolumeId != p.Id {
							continue
						}
						v, err := client.CinderV2().Volume().Show(vol.VolumeId)
						console.Info("[volume: %s] status is %s", vol.Id, v.Status)
						if err == nil && v.IsInuse() {
							console.Info("[volume: %s] attach success", p.Id)
							return nil
						}
					}
					if time.Since(startTime) >= time.Minute*10 {
						break
					}
					time.Sleep(time.Second * 5)

				}
				console.Error("[volume: %s] attach failed", p)
				return nil
			},
		}
		console.Info("attaching ...")
		taskGroup2.Start()
	},
}

func init() {
	attachVolume.Flags().Int("nums", 1, "nums of interfaces")
	attachVolume.Flags().Int("parallel", runtime.NumCPU(), "nums of parallel")
	attachVolume.Flags().Int("size", 10, "size of volume")
	attachVolume.Flags().String("type", "", "attach volume with specified type")

	ServerCommand.AddCommand(attachVolume)
}
