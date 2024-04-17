package main

import (
	"fmt"
	"github.com/LeBronQ/Mobility"
	"github.com/LeBronQ/RadioChannelModel"
	"sync"
)

const (
	NodeNum = 1000
)

type Node struct {
	ID      int64
	MobNode Mobility.Node
	WNode   RadioChannelModel.WirelessNode
	Range   float64
}

func GenerateNodes() []*Node {
	arr := make([]*Node, NodeNum)
	for i := 0; i < NodeNum; i++ {
		node := &Mobility.Node{
			Pos:  Mobility.Nbox.RandomPosition3D(),
			Time: 10,
			V: Mobility.Speed{
				X: 10., Y: 10., Z: 10.,
			},
			Model: "RandomWalk",
			Param: Mobility.RandomWalkParam{
				MinSpeed: 0,
				MaxSpeed: 20,
			},
		}
		wirelessNode := &RadioChannelModel.WirelessNode{
			Frequency:  2.4e+9,
			BitRate:    5.0e+7,
			Modulation: "BPSK",
			BandWidth:  2.0e+7,
			M:          0,
			PowerInDbm: 20,
		}
		n := &Node{
			ID:      int64(i),
			MobNode: *node,
			WNode:   *wirelessNode,
			Range:   2000.0,
		}
		arr[i] = n
	}
	return arr
}

type ChannelJob struct {
	TX Node
	RX Node
}

func ChannelWorker(id int, jobs <-chan ChannelJob, result chan<- float64) {
	for job := range jobs {
		//fmt.Printf("Worker %d processing job\n", id)
		PLR := RadioChannelModel.ChannelParameterCalculation(0, job.TX.WNode, job.RX.WNode, RadioChannelModel.Position(job.TX.MobNode.Pos), RadioChannelModel.Position(job.RX.MobNode.Pos))
		result <- PLR
	}
}
func IntfWorker(id int, jobs <-chan float64, wg *sync.WaitGroup) {
	for _ = range jobs {
		mutex.Lock() // 加锁
		counter++
		mutex.Unlock()
		/*s := strconv.FormatFloat(job*100, 'f', -1, 64)
		cmd := exec.Command("sudo", "tc", "qdisc", "change", "dev", "lo", "root", "netem", "loss", s)
		err := cmd.Run()
		if err != nil {
			fmt.Println("Error:", err)
		}*/

		//fmt.Println(job)
		wg.Done()
	}
}

var mutex sync.Mutex
var counter int64

func main() {
	counter = 0
	var wg sync.WaitGroup
	var graph [NodeNum][]*Node
	NodeArr := make([]*Node, NodeNum)
	NodeArr = GenerateNodes()
	/*interval := 100 * time.Millisecond
	ticker := time.Tick(interval)
	timer := time.After(10 * time.Second)

		for {
			select {
			case <-ticker:
				// 在每个时间间隔执行的函数
				Update(NodeArr, graph)
			case <-timer:
				// 计时器到期时停止循环
				fmt.Println("Timer expired. Stopping...")
			}
		}*/

	for _, node := range NodeArr {
		Mobility.UpdatePosition(&node.MobNode)
	}

	ChannelJobs := make(chan ChannelJob, 100)
	PLR := make(chan float64, 1000000)
	cnt := 0
	for i := 0; i < 10; i++ {
		go ChannelWorker(i, ChannelJobs, PLR)
	}
	for i := 0; i < 10; i++ {
		go IntfWorker(i, PLR, &wg)
	}
	for i, target := range NodeArr {
		distance := target.Range
		var neighbors []*Node
		for _, node := range NodeArr {
			if node.ID != target.ID {
				if Mobility.CalculateDistance3D(node.MobNode.Pos, target.MobNode.Pos) <= distance {
					cnt++
					neighbors = append(neighbors, node)
					link := ChannelJob{TX: *target, RX: *node}
					wg.Add(1)
					ChannelJobs <- link
				}
			}
		}
		graph[i] = neighbors
	}
	wg.Wait()
	fmt.Printf("cnt:%d\n", cnt)
	fmt.Printf("counter:%d\n", counter)

}

func UpdatePosition(NodeArr []*Node, start int64, end int64) {
	//fmt.Printf("%d-%d Done!\n", start, end)
	for i := start; i < end; i++ {
		node := NodeArr[i]
		Mobility.UpdatePosition(&node.MobNode)
	}
}

func UpdateNeighbors(graph *[NodeNum][]*Node, NodeArr []*Node, start int, end int, channel chan<- ChannelJob) {
	for i := start; i < end; i++ {
		target := NodeArr[i]
		distance := target.Range
		var neighbors []*Node
		for _, node := range NodeArr {
			if node.ID != target.ID {
				if Mobility.CalculateDistance3D(node.MobNode.Pos, target.MobNode.Pos) <= distance {
					neighbors = append(neighbors, node)
					link := ChannelJob{TX: *target, RX: *node}
					channel <- link
				}
			}
		}
		graph[i] = neighbors
	}
}

func CalculateChannel(graph [NodeNum][]*Node, NodeArr []*Node, start int64, end int64) {
	for i := start; i < end; i++ {
		n := graph[i]
		for _, node := range n {
			RadioChannelModel.ChannelParameterCalculation(0, NodeArr[i].WNode, node.WNode, RadioChannelModel.Position(NodeArr[i].MobNode.Pos), RadioChannelModel.Position(node.MobNode.Pos))
		}
	}
}
