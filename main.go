package main

import (
	"errors"
	"sort"

	"github.com/docker/swarm/scheduler/strategy/plugin"
	"github.com/samalba/dockerclient"
)

type ReverseBinpackStrategy struct {
}

func (p *ReverseBinpackStrategy) Name() string {
	return "revbinpack"
}

func (p *ReverseBinpackStrategy) Initialize() error {
	return nil
}

func (p *ReverseBinpackStrategy) PlaceContainer(config *dockerclient.ContainerConfig, nodes []*plugin.Node) (*plugin.Node, error) {
	scores := scores{}

	for _, node := range nodes {
		nodeMemory := node.UsableMemory
		nodeCpus := node.UsableCpus

		// Skip nodes that are smaller than the requested resources.
		if nodeMemory < int64(config.Memory) || nodeCpus < config.CpuShares {
			continue
		}

		var (
			cpuScore    int64 = 100
			memoryScore int64 = 100
		)

		if config.CpuShares > 0 {
			cpuScore = (node.ReservedCpus + config.CpuShares) * 100 / nodeCpus
		}
		if config.Memory > 0 {
			memoryScore = (node.ReservedMemory + config.Memory) * 100 / nodeMemory
		}

		if cpuScore <= 100 && memoryScore <= 100 {
			scores = append(scores, &score{node: node, score: cpuScore + memoryScore})
		}
	}

	if len(scores) == 0 {
		return nil, errors.New("Resource not available")
	}

	sort.Sort(scores)

	return scores[0].node, nil
}

type score struct {
	node  *plugin.Node
	score int64
}

type scores []*score

func (s scores) Len() int {
	return len(s)
}

func (s scores) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s scores) Less(i, j int) bool {
	var (
		ip = s[i]
		jp = s[j]
	)

	// reverse comparison,
	// so the whole cluster will be distributedly filled
	return ip.score < jp.score
}

func main() {
	plugin.Run(&ReverseBinpackStrategy{})
}
