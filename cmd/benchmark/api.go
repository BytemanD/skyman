package benchmark

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/BytemanD/go-console/console"
	"github.com/BytemanD/skyman/benchmark"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/utility"

	"github.com/spf13/cobra"
)

var BenchmarkCmd = &cobra.Command{
	Use:  "benchmark <name>",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := openstack.DefaultClient()

		worker, _ := cmd.Flags().GetInt("worker")
		query := url.Values{
			"limit": []string{strconv.Itoa(worker)},
		}
		var results []benchmark.BenchmarkResult
		switch args[0] {
		case "server-show":
			console.Info("list servers")
			servers, err := client.NovaV2().Server().List(query)
			utility.LogIfError(err, true, "list server failed")
			cases := []benchmark.ServerShow{}
			for _, server := range servers {
				cases = append(cases, benchmark.ServerShow{
					Client: client,
					Server: server,
				})
			}
			console.Info("start test")
			results = benchmark.RunBenchmarkTest(cases)

		default:
			console.Fatal("invalid case: %s", args[0])
		}
		for _, result := range results {
			fmt.Println(
				result.Start.Local().Format("2006-01-02 15:04:05"),
				result.End.Local().Format("2006-01-02 15:04:05"),
				result.Spend())
		}
	},
}

func init() {
	BenchmarkCmd.Flags().Int("worker", 1, "Worker")

}
