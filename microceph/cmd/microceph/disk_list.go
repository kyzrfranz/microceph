package main

import (
	"context"
	"fmt"
	"os"
	"sort"

	"github.com/canonical/microcluster/microcluster"
	"github.com/lxc/lxd/lxc/utils"
	"github.com/lxc/lxd/shared/units"
	"github.com/spf13/cobra"

	"github.com/canonical/microceph/microceph/client"
)

type cmdDiskList struct {
	common    *CmdControl
	disk      *cmdDisk
	apiClient client.ApiReader
}

func (c *cmdDiskList) Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List servers in the cluster",
		RunE:  c.Run,
	}

	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		m, err := microcluster.App(context.Background(), microcluster.Args{StateDir: c.common.FlagStateDir, Verbose: c.common.FlagLogVerbose, Debug: c.common.FlagLogDebug})
		if err != nil {
			return err
		}
		cli, err := m.LocalClient()
		c.apiClient = client.NewClient(cli)
		return err
	}

	return cmd
}

func (c *cmdDiskList) Run(cmd *cobra.Command, args []string) error {

	// List configured disks.
	disks, err := c.apiClient.GetDisks(context.Background())
	if err != nil {
		return err
	}

	data := make([][]string, len(disks))
	for i, disk := range disks {
		data[i] = []string{fmt.Sprintf("%d", disk.OSD), disk.Location, disk.Path}
	}

	header := []string{"OSD", "LOCATION", "PATH"}
	sort.Sort(utils.ByName(data))

	fmt.Println("Disks configured in MicroCeph:")
	err = utils.RenderTable(utils.TableFormatTable, header, data, disks)
	if err != nil {
		return err
	}

	// List local disks.
	err = listLocalDisks(c.apiClient)
	if err != nil {
		return err
	}

	return nil
}

func listLocalDisks(cli client.ApiReader) error {
	// List configured disks.
	disks, err := cli.GetDisks(context.Background())
	if err != nil {
		return err
	}

	// List physical disks.
	resources, err := cli.GetResources(context.Background())
	if err != nil {
		return err
	}

	// Get local hostname.
	hostname, err := os.Hostname()
	if err != nil {
		return err
	}

	// Prepare the table.
	data := [][]string{}
	for _, disk := range resources.Disks {
		if len(disk.Partitions) > 0 {
			continue
		}

		devicePath := fmt.Sprintf("/dev/disk/by-id/%s", disk.DeviceID)

		found := false
		for _, entry := range disks {
			if entry.Location != hostname {
				continue
			}

			if entry.Path == devicePath {
				found = true
				break
			}
		}

		if found {
			continue
		}

		data = append(data, []string{disk.Model, units.GetByteSizeStringIEC(int64(disk.Size), 2), disk.Type, devicePath})
	}

	fmt.Println("")
	fmt.Println("Available unpartitioned disks on this system:")

	header := []string{"MODEL", "CAPACITY", "TYPE", "PATH"}
	sort.Sort(utils.ByName(data))

	err = utils.RenderTable(utils.TableFormatTable, header, data, disks)
	if err != nil {
		return err
	}

	return nil
}
