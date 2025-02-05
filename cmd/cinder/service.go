package cinder

import (
	"net/url"

	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack/model/cinder"
	"github.com/BytemanD/skyman/utility"
	"github.com/spf13/cobra"
)

var service = &cobra.Command{Use: "service", Short: "Volume service command"}

var list = &cobra.Command{
	Use:   "list",
	Short: "List volume services",
	Run: func(cmd *cobra.Command, _ []string) {
		client := common.DefaultClient()

		query := url.Values{}
		binary, _ := cmd.Flags().GetString("binary")
		host, _ := cmd.Flags().GetString("host")
		zone, _ := cmd.Flags().GetString("zone")
		long, _ := cmd.Flags().GetBool("long")

		if binary != "" {
			query.Set("binary", binary)
		}
		if host != "" {
			query.Set("host", host)
		}

		services, err := client.CinderV2().Service().List(query)
		utility.LogIfError(err, true, "get services failed")
		if zone != "" {
			filterServices := []cinder.Service{}
			for _, srv := range services {
				if srv.Zone != zone {
					continue
				}
				filterServices = append(filterServices, srv)
			}
			common.PrintVolumeServices(filterServices, long)
		} else {
			common.PrintVolumeServices(services, long)
		}
	},
}

func init() {
	list.Flags().BoolP("long", "l", false, "List additional fields in output")

	service.AddCommand(list)

	Volume.AddCommand(service)
}
