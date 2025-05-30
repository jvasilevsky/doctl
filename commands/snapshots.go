/*
Copyright 2018 The Doctl Authors All rights reserved.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package commands

import (
	"fmt"

	"github.com/digitalocean/doctl"
	"github.com/digitalocean/doctl/commands/displayers"
	"github.com/digitalocean/doctl/do"
	"github.com/gobwas/glob"
	"github.com/spf13/cobra"
)

// Snapshot creates the snapshot command
func Snapshot() *Command {
	cmd := &Command{
		Command: &cobra.Command{
			Use:     "snapshot",
			Aliases: []string{"s"},
			Short:   "Access and manage snapshots",
			Long:    "The subcommands of `doctl compute snapshot` allow you to manage and retrieve information about Droplet and block storage volume snapshots.",
		},
	}

	snapshotDetail := `

- The snapshot's ID
- The snapshot's name
- The date and time when the snapshot was created
- The slugs of the datacenter regions in which the snapshot is available
- The type of resource the snapshot was made from (either from Droplet or volume) and its ID
- The minimum size required for a Droplet or volume to use this snapshot, in GB
- The compressed, billable size of the snapshot
`

	cmdRunSnapshotList := CmdBuilder(cmd, RunSnapshotList, "list [glob]",
		"List Droplet and volume snapshots", "Retrieves a list of snapshots and their information, including:"+snapshotDetail,
		Writer, aliasOpt("ls"), displayerType(&displayers.Snapshot{}))
	AddStringFlag(cmdRunSnapshotList, doctl.ArgResourceType, "", "", "Filters by resource type (`droplet` or `volume`)")
	AddStringFlag(cmdRunSnapshotList, doctl.ArgRegionSlug, "", "", "Filters by regional availability")
	cmdRunSnapshotList.Example = `The following example lists all Droplet snapshots in the ` + "`" + `nyc1` + "`" + ` region and uses the ` + "`" + `--format` + "`" + ` flag to return only name, ID, and resource type for each snapshot: doctl compute snapshot list --resource droplet --region nyc1 --format Name,ID,ResourceType`

	cmdSnapshotGet := CmdBuilder(cmd, RunSnapshotGet, "get <snapshot-id>...",
		"Retrieve a Droplet or volume snapshot", "Retrieves information about a Droplet or block storage volume snapshot, including:"+snapshotDetail,
		Writer, aliasOpt("g"), displayerType(&displayers.Snapshot{}))
	cmdSnapshotGet.Example = `The following example retrieves information about a Droplet snapshot with ID ` + "`" + `386734086` + "`" + `: doctl compute snapshot get 386734086`

	cmdRunSnapshotDelete := CmdBuilder(cmd, RunSnapshotDelete, "delete <snapshot-id>...",
		"Delete a snapshot of a Droplet or volume", "Deletes the specified snapshot or volume. This is irreversible.",
		Writer, aliasOpt("d", "rm"), displayerType(&displayers.Snapshot{}))
	AddBoolFlag(cmdRunSnapshotDelete, doctl.ArgForce, doctl.ArgShortForce, false, "Delete the snapshot without confirmation")
	cmdRunSnapshotDelete.Example = `The following example deletes a Droplet snapshot with ID ` + "`" + `386734086` + "`" + `: doctl compute snapshot delete 386734086`

	return cmd
}

// RunSnapshotList returns a list of snapshots
func RunSnapshotList(c *CmdConfig) error {
	var err error
	ss := c.Snapshots()

	restype, err := c.Doit.GetString(c.NS, doctl.ArgResourceType)
	if err != nil {
		return err
	}

	region, err := c.Doit.GetString(c.NS, doctl.ArgRegionSlug)
	if err != nil {
		return err
	}

	matches := make([]glob.Glob, 0, len(c.Args))
	for _, globStr := range c.Args {
		g, err := glob.Compile(globStr)
		if err != nil {
			return fmt.Errorf("unknown glob %q", globStr)
		}

		matches = append(matches, g)
	}

	var matchedList []do.Snapshot
	var list []do.Snapshot

	switch restype {
	case "droplet":
		list, err = ss.ListDroplet()
		if err != nil {
			return err
		}
	case "volume":
		if region != "" {
			list, err = ss.ListVolumeSnapshotByRegion(region)
		} else {
			list, err = ss.ListVolume()
		}
		if err != nil {
			return err
		}

	default:
		list, err = ss.List()
		if err != nil {
			return err
		}
	}

	for _, snapshot := range list {
		var skip = true
		if len(matches) == 0 {
			skip = false
		} else {
			for _, m := range matches {
				if m.Match(snapshot.ID) {
					skip = false
				}
				if m.Match(snapshot.Name) {
					skip = false
				}
			}
		}

		if !skip && region != "" {
			for _, snapshotRegion := range snapshot.Regions {
				if region != snapshotRegion {
					skip = true
				} else {
					skip = false
					break
				}
			}

		}

		if !skip {
			matchedList = append(matchedList, snapshot)
		}
	}

	item := &displayers.Snapshot{Snapshots: matchedList}
	return c.Display(item)
}

// RunSnapshotGet returns a snapshot
func RunSnapshotGet(c *CmdConfig) error {
	if len(c.Args) == 0 {
		return doctl.NewMissingArgsErr(c.NS)
	}

	ss := c.Snapshots()
	ids := c.Args

	matchedList := make([]do.Snapshot, 0, len(ids))

	for _, id := range ids {
		s, err := ss.Get(id)
		if err != nil {
			return err
		}
		matchedList = append(matchedList, *s)
	}
	item := &displayers.Snapshot{Snapshots: matchedList}
	return c.Display(item)
}

// RunSnapshotDelete destroys snapshot(s) by id
func RunSnapshotDelete(c *CmdConfig) error {
	if len(c.Args) == 0 {
		return doctl.NewMissingArgsErr(c.NS)
	}

	force, err := c.Doit.GetBool(c.NS, doctl.ArgForce)
	if err != nil {
		return err
	}

	ss := c.Snapshots()
	ids := c.Args

	if force || AskForConfirmDelete("snapshot", len(ids)) == nil {
		for _, id := range ids {
			err := ss.Delete(id)
			if err != nil {
				return err
			}
		}
	} else {
		return errOperationAborted
	}
	return nil
}
