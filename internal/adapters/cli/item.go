package cli

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/RockLikeAmadeus/go-application-starter/internal/bootstrap"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
)

func newItemCmd(open appOpener) *cobra.Command {
	var asJSON bool

	cmd := &cobra.Command{Use: "item", Short: "Manage items"}
	cmd.PersistentFlags().BoolVar(&asJSON, "json", false, "print machine-readable JSON")

	printItems := func(cmd *cobra.Command, items []domain.Item) error {
		if asJSON {
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			out := make([]itemJSON, 0, len(items))
			for _, it := range items {
				out = append(out, toItemJSON(it))
			}
			return enc.Encode(out)
		}
		tw := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 4, 2, ' ', 0)
		fmt.Fprintln(tw, "ID\tNAME\tUPDATED")
		for _, it := range items {
			fmt.Fprintf(tw, "%s\t%s\t%s\n", it.ID, it.Name, it.UpdatedAt.Format(time.RFC3339))
		}
		return tw.Flush()
	}
	printItem := func(cmd *cobra.Command, item domain.Item) error {
		if asJSON {
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(toItemJSON(item))
		}
		return printItems(cmd, []domain.Item{item})
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "add NAME",
			Short: "Create an item",
			Args:  cobra.ExactArgs(1),
			RunE: withApp(open, func(cmd *cobra.Command, args []string, app *bootstrap.App) error {
				item, err := app.Items.CreateItem(cmd.Context(), args[0])
				if err != nil {
					return err
				}
				return printItem(cmd, item)
			}),
		},
		&cobra.Command{
			Use:     "list",
			Aliases: []string{"ls"},
			Short:   "List items",
			Args:    cobra.NoArgs,
			RunE: withApp(open, func(cmd *cobra.Command, _ []string, app *bootstrap.App) error {
				items, err := app.Items.ListItems(cmd.Context())
				if err != nil {
					return err
				}
				return printItems(cmd, items)
			}),
		},
		&cobra.Command{
			Use:   "get ID",
			Short: "Show one item",
			Args:  cobra.ExactArgs(1),
			RunE: withApp(open, func(cmd *cobra.Command, args []string, app *bootstrap.App) error {
				item, err := app.Items.GetItem(cmd.Context(), args[0])
				if err != nil {
					return err
				}
				return printItem(cmd, item)
			}),
		},
		&cobra.Command{
			Use:   "rename ID NAME",
			Short: "Rename an item",
			Args:  cobra.ExactArgs(2),
			RunE: withApp(open, func(cmd *cobra.Command, args []string, app *bootstrap.App) error {
				item, err := app.Items.RenameItem(cmd.Context(), args[0], args[1])
				if err != nil {
					return err
				}
				return printItem(cmd, item)
			}),
		},
		&cobra.Command{
			Use:     "delete ID",
			Aliases: []string{"rm"},
			Short:   "Delete an item",
			Args:    cobra.ExactArgs(1),
			RunE: withApp(open, func(cmd *cobra.Command, args []string, app *bootstrap.App) error {
				if err := app.Items.DeleteItem(cmd.Context(), args[0]); err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "deleted %s\n", args[0])
				return nil
			}),
		},
		Gated("item-export", &cobra.Command{
			Use:   "export",
			Short: "Export the items as CSV",
			Args:  cobra.NoArgs,
			RunE: withApp(open, func(cmd *cobra.Command, _ []string, app *bootstrap.App) error {
				csv, err := app.Items.ExportItems(cmd.Context())
				if err != nil {
					return err
				}
				_, err = cmd.OutOrStdout().Write(csv)
				return err
			}),
		}),
	)
	return cmd
}

type itemJSON struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func toItemJSON(i domain.Item) itemJSON {
	return itemJSON{ID: i.ID, Name: i.Name, CreatedAt: i.CreatedAt, UpdatedAt: i.UpdatedAt}
}
