package cluster

import (
	"context"
	"log"
	"sync"
	"time"
)

// NodeStatus represents the state of an Aurix cluster node.
type NodeStatus string

const (
	NodeHealthy NodeStatus = "healthy"
	NodeSuspect NodeStatus = "suspect"
	NodeOffline NodeStatus = "offline"
)

// NodeInfo holds telemetry and status data for a cluster member.
type NodeInfo struct {
	ID        string     `json:"id"`
	Address   string     `json:"address"`
	Status    NodeStatus `json:"status"`
	Players   int        `json:"players"`
	LastPing  time.Time  `json:"lastPing"`
}

// ClusterManager manages multi-node health monitoring and automatic player failover.
type ClusterManager struct {
	nodeID   string
	nodes    map[string]*NodeInfo
	mu       sync.RWMutex
	stopChan chan struct{}
}

// NewClusterManager initializes a cluster manager for nodeID.
func NewClusterManager(nodeID string) *ClusterManager {
	return &ClusterManager{
		nodeID:   nodeID,
		nodes:    make(map[string]*NodeInfo),
		stopChan: make(chan struct{}),
	}
}

// RegisterNode adds or updates a node in the cluster view.
func (cm *ClusterManager) RegisterNode(node *NodeInfo) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	node.LastPing = time.Now()
	cm.nodes[node.ID] = node
}

// StartHeartbeatMonitor monitors node health and triggers failover when nodes go offline.
func (cm *ClusterManager) StartHeartbeatMonitor(ctx context.Context, onNodeFail func(failedNodeID string)) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-cm.stopChan:
			return
		case <-ticker.C:
			cm.mu.Lock()
			now := time.Now()
			for id, node := range cm.nodes {
				if id == cm.nodeID {
					continue
				}
				if now.Sub(node.LastPing) > 15*time.Second && node.Status != NodeOffline {
					node.Status = NodeOffline
					log.Printf("[Cluster] Node %s marked OFFLINE. Triggering failover...", id)
					if onNodeFail != nil {
						go onNodeFail(id)
					}
				}
			}
			cm.mu.Unlock()
		}
	}
}

// ListNodes returns a snapshot of all known cluster nodes.
func (cm *ClusterManager) ListNodes() []*NodeInfo {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var list []*NodeInfo
	for _, n := range cm.nodes {
		list = append(list, n)
	}
	return list
}

// Stop terminates the cluster monitor loop.
func (cm *ClusterManager) Stop() {
	close(cm.stopChan)
}
