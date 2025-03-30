package nova

import (
	"net/url"

	"github.com/duke-git/lancet/v2/fileutil"
	"github.com/spf13/cobra"

	"github.com/BytemanD/go-console/console"
	"github.com/BytemanD/skyman/cmd/flags"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack/model/nova"
	"github.com/BytemanD/skyman/utility"
)

var (
	keypairListFlags   flags.KeypairListFlags
	keypairCreateFlags flags.KeypairCreateFlags
)

var Keypair = &cobra.Command{Use: "keypair"}

var keypairList = &cobra.Command{
	Use:   "list",
	Short: "List keypairs",
	Args:  cobra.ExactArgs(0),
	Run: func(_ *cobra.Command, _ []string) {
		client := common.DefaultClient()
		query := url.Values{}
		if *keypairListFlags.UserId != "" {
			query.Set("user_id", *keypairListFlags.UserId)
		}
		keypairs, err := client.NovaV2().ListKeypair(query)
		if err != nil {
			console.Fatal("%s", err)
		}
		pt := common.PrettyTable{
			ShortColumns: []common.Column{
				{Name: "Name", Slot: func(item interface{}) interface{} {
					p, _ := item.(nova.Keypair)
					return p.Keypair.Name
				}},
				{Name: "Type", Slot: func(item interface{}) interface{} {
					p, _ := item.(nova.Keypair)
					return p.Keypair.Type
				}},
				{Name: "Fingerprint", Slot: func(item interface{}) interface{} {
					p, _ := item.(nova.Keypair)
					return p.Keypair.Fingerprint
				}},
			},
		}
		pt.AddItems(keypairs)
		common.PrintPrettyTable(pt, false)
	},
}
var keypairShow = &cobra.Command{
	Use:   "show <name>",
	Short: "show keypair",
	Args:  cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		client := common.DefaultClient()
		keypair, err := client.NovaV2().GetKeypair(args[0])

		utility.LogIfError(err, true, "get keypair failed")
		common.PrintKeypair(*keypair)
	},
}
var keypairCreate = &cobra.Command{
	Use:   "create <name>",
	Short: "create keypair",
	Args:  cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		client := common.DefaultClient()
		opt := nova.KeypairOpt{
			UserId: *keypairCreateFlags.UserId,
		}
		if *keypairCreateFlags.PubKey != "" {
			if !fileutil.IsExist(*keypairCreateFlags.PubKey) {
				console.Fatal("file '%s' not exists", *keypairCreateFlags.PubKey)
			}
			if fileutil.IsDir(*keypairCreateFlags.PubKey) {
				console.Fatal("'%s' is not a file", *keypairCreateFlags.PubKey)
			}
			if content, err := fileutil.ReadFileToString(*keypairCreateFlags.PubKey); err == nil {
				opt.PublicKey = content
			} else {
				utility.LogIfError(err, true, "read public key failed")
			}
		}

		keypair, err := client.NovaV2().CreateKeypair(
			args[0], *keypairCreateFlags.Type, opt)

		utility.LogIfError(err, true, "create keypair failed")
		if err != nil {
			console.Fatal("create keypair failed: %s", err)
		}
		common.PrintKeypair(*keypair)
	},
}

var keypairDelete = &cobra.Command{
	Use:   "delete <keypair> [<keyapir> ...]",
	Short: "delete keypair(s)",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := common.DefaultClient()
		for _, keypair := range args {
			err := client.NovaV2().DeleteKeypair(keypair)
			utility.LogIfError(err, false, "delete keypair %s failed", keypair)
		}
	},
}

func init() {
	keypairListFlags = flags.KeypairListFlags{
		UserId: keypairList.Flags().String("user-id", "", "List by user id"),
	}
	keypairCreateFlags = flags.KeypairCreateFlags{
		UserId: keypairCreate.Flags().String("user-id", "", "ID of user to whom to add key-pair (Admin only)."),
		Type:   keypairCreate.Flags().String("type", "ssh", " Keypair type. Can be ssh, ecdsa or x509."),
		PubKey: keypairCreate.Flags().String("pub-key", "", "Path to a public ssh key."),
	}
	Keypair.AddCommand(keypairList, keypairShow, keypairCreate, keypairDelete)
}
