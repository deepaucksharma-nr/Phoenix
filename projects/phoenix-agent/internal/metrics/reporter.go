package metrics

import (
	"context"
	"time"

	"github.com/phoenix/platform/projects/phoenix-agent/internal/config"
	"github.com/phoenix/platform/projects/phoenix-agent/internal/poller"
	"github.com/rs/zerolog/log"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

type Reporter struct {
	config *config.Config
	client *poller.Client
}

func NewReporter(cfg *config.Config, client *poller.Client) *Reporter {
	return &Reporter{
		config: cfg,
		client: client,
	}
}

// Start begins periodic metrics reporting
func (r *Reporter) Start(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Initial report
	r.reportMetrics(ctx)

	for {
		select {
		case <-ticker.C:
			r.reportMetrics(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (r *Reporter) reportMetrics(ctx context.Context) {
	metrics := r.collectSystemMetrics()
	
	if err := r.client.SendMetrics(ctx, metrics); err != nil {
		log.Error().Err(err).Msg("Failed to send metrics")
	}
}

func (r *Reporter) collectSystemMetrics() []map[string]interface{} {
	var metrics []map[string]interface{}
	timestamp := time.Now().Unix()

	// CPU metrics
	if cpuPercent, err := cpu.Percent(1*time.Second, false); err == nil && len(cpuPercent) > 0 {
		metrics = append(metrics, map[string]interface{}{
			"name":      "agent.cpu.percent",
			"value":     cpuPercent[0],
			"timestamp": timestamp,
			"labels": map[string]string{
				"host_id": r.config.HostID,
			},
		})
	}

	// Memory metrics
	if memInfo, err := mem.VirtualMemory(); err == nil {
		metrics = append(metrics, map[string]interface{}{
			"name":      "agent.memory.used_bytes",
			"value":     memInfo.Used,
			"timestamp": timestamp,
			"labels": map[string]string{
				"host_id": r.config.HostID,
			},
		})
		
		metrics = append(metrics, map[string]interface{}{
			"name":      "agent.memory.percent",
			"value":     memInfo.UsedPercent,
			"timestamp": timestamp,
			"labels": map[string]string{
				"host_id": r.config.HostID,
			},
		})
	}

	// Disk metrics
	if diskInfo, err := disk.Usage("/"); err == nil {
		metrics = append(metrics, map[string]interface{}{
			"name":      "agent.disk.used_bytes",
			"value":     diskInfo.Used,
			"timestamp": timestamp,
			"labels": map[string]string{
				"host_id": r.config.HostID,
				"path":    "/",
			},
		})
		
		metrics = append(metrics, map[string]interface{}{
			"name":      "agent.disk.percent",
			"value":     diskInfo.UsedPercent,
			"timestamp": timestamp,
			"labels": map[string]string{
				"host_id": r.config.HostID,
				"path":    "/",
			},
		})
	}

	// Network metrics
	if netStats, err := net.IOCounters(false); err == nil && len(netStats) > 0 {
		metrics = append(metrics, map[string]interface{}{
			"name":      "agent.network.bytes_sent",
			"value":     netStats[0].BytesSent,
			"timestamp": timestamp,
			"labels": map[string]string{
				"host_id": r.config.HostID,
			},
		})
		
		metrics = append(metrics, map[string]interface{}{
			"name":      "agent.network.bytes_recv",
			"value":     netStats[0].BytesRecv,
			"timestamp": timestamp,
			"labels": map[string]string{
				"host_id": r.config.HostID,
			},
		})
	}

	// Agent-specific metrics
	metrics = append(metrics, map[string]interface{}{
		"name":      "agent.uptime_seconds",
		"value":     time.Since(startTime).Seconds(),
		"timestamp": timestamp,
		"labels": map[string]string{
			"host_id": r.config.HostID,
			"version": "1.0.0",
		},
	})

	return metrics
}

var startTime = time.Now()