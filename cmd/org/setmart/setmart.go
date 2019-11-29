package setmart

import (
	"github.com/spf13/cobra"
	"github.com/srinandan/apigeecli/cmd/org/setprop"
	"github.com/srinandan/apigeecli/cmd/shared"
)

//Cmd to set mart endpoint
var Cmd = &cobra.Command{
	Use:   "setmart",
	Short: "Set MART endpoint for an Apigee Org",
	Long:  "Set MART endpoint for an Apigee Org",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		return setprop.SetOrgProperty("features.mart.server.endpoint", mart)
	},
}

var mart string

func init() {

	Cmd.Flags().StringVarP(&shared.RootArgs.Org, "org", "o",
		"", "Apigee organization name")
	Cmd.Flags().StringVarP(&mart, "mart", "m",
		"", "MART Endpoint")

	_ = Cmd.MarkFlagRequired("org")
	_ = Cmd.MarkFlagRequired("mart")
}
