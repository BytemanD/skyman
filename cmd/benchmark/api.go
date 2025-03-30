package benchmark

import (
	"net/url"
	"strconv"

	"github.com/BytemanD/go-console/console"
	"github.com/BytemanD/skyman/benchmark"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack/model/nova"
	"github.com/BytemanD/skyman/utility"
	"github.com/samber/lo"

	"github.com/spf13/cobra"
)

var BenchmarkCmd = &cobra.Command{
	Use:  "benchmark <name>",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := common.DefaultClient()

		worker, _ := cmd.Flags().GetInt("worker")
		query := url.Values{
			"limit": []string{strconv.Itoa(worker)},
		}
		var results []benchmark.BenchmarkResult
		switch args[0] {
		case "server-show":
			console.Info("list servers")
			servers, err := client.NovaV2().ListServer(query)
			utility.LogIfError(err, true, "list server failed")
			cases := lo.Map(servers, func(item nova.Server, _ int) benchmark.ServerShow {
				return benchmark.ServerShow{
					Client: client,
					Server: item,
				}
			})
			console.Info("start test")
			results = benchmark.RunBenchmarkTest(cases)

		default:
			console.Fatal("invalid case: %s", args[0])
		}
		for _, result := range results {
			println(
				result.Start.Local().Format("2006-01-02 15:04:05"),
				result.End.Local().Format("2006-01-02 15:04:05"),
				result.Spend())
		}
	},
}

func init() {
	BenchmarkCmd.Flags().Int("worker", 1, "Worker")

}
