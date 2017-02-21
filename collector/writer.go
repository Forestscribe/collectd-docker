package collector

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/influxdata/influxdb/client/v2"
)

// InfluxdbWriter is responsible for writing data to influxdb
type InfluxdbWriter struct {
	config client.HTTPConfig
	db     string
}

// NewInfluxdbWriter creates new InfluxdbWriter
// with specified hostname and writer
func NewInfluxdbWriter(host string, db string, username string, password string) InfluxdbWriter {
	return InfluxdbWriter{
		config: client.HTTPConfig{
			Addr:     host,
			Username: username,
			Password: password,
		},
		db: db,
	}
}

func (writer InfluxdbWriter) Write(s Stats) error {
	return writer.writeInts(s)
}

func (writer InfluxdbWriter) writeInts(s Stats) error {
	// Create a new HTTPClient
	c, err := client.NewHTTPClient(writer.config)
	if err != nil {
		log.Print(err)
		return err
	}

	// Create a new point batch
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  writer.db,
		Precision: "s",
	})
	if err != nil {
		log.Print(err)
		return err
	}
	tags := map[string]string{
		"app":  s.App,
		"task": s.Task,
	}
	// influxdb only supoprt int64 (not uint64), thus we need to convert everything to int64
	writer.addPoint(bp, "cpu", tags, map[string]interface{}{
		"user":   int64(s.Stats.CPUStats.CPUUsage.UsageInUsermode),
		"system": int64(s.Stats.CPUStats.CPUUsage.UsageInKernelmode),
		"total":  int64(s.Stats.CPUStats.CPUUsage.TotalUsage),
	}, s.Stats.Read)

	writer.addPoint(bp, "memory", tags, map[string]interface{}{
		"limit":         int64(s.Stats.MemoryStats.Limit),
		"max":           int64(s.Stats.MemoryStats.MaxUsage),
		"usage":         int64(s.Stats.MemoryStats.Usage),
		"active_anon":   int64(s.Stats.MemoryStats.Stats.TotalActiveAnon),
		"active_file":   int64(s.Stats.MemoryStats.Stats.TotalActiveFile),
		"cache":         int64(s.Stats.MemoryStats.Stats.TotalCache),
		"inactive_anon": int64(s.Stats.MemoryStats.Stats.TotalInactiveAnon),
		"inactive_file": int64(s.Stats.MemoryStats.Stats.TotalInactiveFile),
		"mapped_file":   int64(s.Stats.MemoryStats.Stats.TotalMappedFile),
		"pg_fault":      int64(s.Stats.MemoryStats.Stats.TotalPgfault),
		"pg_in":         int64(s.Stats.MemoryStats.Stats.TotalPgpgin),
		"pg_out":        int64(s.Stats.MemoryStats.Stats.TotalPgpgout),
		"rss":           int64(s.Stats.MemoryStats.Stats.TotalRss),
		"rss_huge":      int64(s.Stats.MemoryStats.Stats.TotalRssHuge),
		"unevictable":   int64(s.Stats.MemoryStats.Stats.TotalUnevictable),
		"writeback":     int64(s.Stats.MemoryStats.Stats.TotalWriteback),
	}, s.Stats.Read)
	metrics := map[string]int64{}

	for _, network := range s.Stats.Networks {
		metrics["rx_bytes"] += int64(network.RxBytes)
		metrics["rx_dropped"] += int64(network.RxDropped)
		metrics["rx_errors"] += int64(network.RxErrors)
		metrics["rx_packets"] += int64(network.RxPackets)

		metrics["tx_bytes"] += int64(network.TxBytes)
		metrics["tx_dropped"] += int64(network.TxDropped)
		metrics["tx_errors"] += int64(network.TxErrors)
		metrics["tx_packets"] += int64(network.TxPackets)
	}
	fields := map[string]interface{}{}
	for k, v := range metrics {
		fields[k] = v
	}
	writer.addPoint(bp, "net", tags, fields, s.Stats.Read)

	// Write the batch
	if err := c.Write(bp); err != nil {
		log.Print(err)
		/* automatically create the database if necessary */
		if strings.Contains(err.Error(), "database not found") {
			q := client.NewQuery(fmt.Sprintf("CREATE DATABASE %s", writer.db), "", "")
			if response, err2 := c.Query(q); err2 == nil && response.Error() == nil {
				fmt.Println(response.Results)
			}
		}
		return err
	}

	return nil
}

func (writer InfluxdbWriter) addPoint(bp client.BatchPoints, name string, tags map[string]string, fields map[string]interface{}, t time.Time) {
	pt, err := client.NewPoint(name, tags, fields, t)
	if err != nil {
		log.Print(err)
		return
	}
	bp.AddPoint(pt)
}
