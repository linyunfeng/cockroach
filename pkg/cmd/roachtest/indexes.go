// Copyright 2018 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License included
// in the file licenses/BSL.txt and at www.mariadb.com/bsl11.
//
// Change Date: 2022-10-01
//
// On the date above, in accordance with the Business Source License, use
// of this software will be governed by the Apache License, Version 2.0,
// included in the file licenses/APL.txt and at
// https://www.apache.org/licenses/LICENSE-2.0

package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

func registerNIndexes(r *registry, secondaryIndexes int) {
	const nodes = 6
	geoZones := []string{"us-west1-b", "us-east1-b", "us-central1-a"}
	geoZonesStr := strings.Join(geoZones, ",")
	r.Add(testSpec{
		Name:    fmt.Sprintf("indexes/%d/nodes=%d/multi-region", secondaryIndexes, nodes),
		Cluster: makeClusterSpec(nodes+1, cpu(16), geo(), zones(geoZonesStr)),
		// Uses CONFIGURE ZONE USING ... COPY FROM PARENT syntax.
		MinVersion: `v19.1.0`,
		Run: func(ctx context.Context, t *test, c *cluster) {
			firstAZ := geoZones[0]
			roachNodes := c.Range(1, nodes)
			gatewayNodes := c.Range(1, nodes/3)
			loadNode := c.Node(nodes + 1)

			c.Put(ctx, cockroach, "./cockroach", roachNodes)
			c.Put(ctx, workload, "./workload", loadNode)
			c.Start(ctx, t, roachNodes)

			t.Status("running workload")
			m := newMonitor(ctx, c, roachNodes)
			m.Go(func(ctx context.Context) error {
				secondary := " --secondary-indexes=" + strconv.Itoa(secondaryIndexes)
				initCmd := "./workload init indexes" + secondary + " {pgurl:1}"
				c.Run(ctx, loadNode, initCmd)

				// Set lease preferences so that all leases for the table are
				// located in the availability zone with the load generator.
				if !local {
					leasePrefs := fmt.Sprintf(`ALTER TABLE indexes.indexes
						                       CONFIGURE ZONE USING
						                       constraints = COPY FROM PARENT,
						                       lease_preferences = '[[+zone=%s]]'`, firstAZ)
					c.Run(ctx, c.Node(1), `./cockroach sql --insecure -e "`+leasePrefs+`"`)
				}

				payload := " --payload=256"
				concurrency := ifLocal("", " --concurrency="+strconv.Itoa(nodes*32))
				duration := " --duration=" + ifLocal("10s", "30m")
				runCmd := fmt.Sprintf("./workload run indexes --histograms=logs/stats.json"+
					payload+concurrency+duration+" {pgurl%s}", gatewayNodes)
				c.Run(ctx, loadNode, runCmd)
				return nil
			})
			m.Wait()
		},
	})
}

func registerIndexes(r *registry) {
	registerNIndexes(r, 2)
}

func registerIndexesBench(r *registry) {
	for i := 0; i <= 10; i++ {
		registerNIndexes(r, i)
	}
}
