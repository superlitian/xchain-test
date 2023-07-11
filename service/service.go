package service

import (
	"math/rand"
	"strconv"
	"sync"
	"time"
	"xchain_test/common"

	"github.com/xuperchain/xuper-sdk-go/v2/account"
	"github.com/xuperchain/xuper-sdk-go/v2/xuper"
)

// Start 程序入口
func Start() {

	// 加载配置文件
	common.LoadConfig()
	// 加载日志系统
	common.InitLogger()

	//部署合约
	common.BeforeDeploy()
	time.Sleep(time.Second * 15)
	common.Deploy()
	time.Sleep(time.Second * 15)
	// 合约调用前置工作
	//common.BeforeInvoke()
	// 监控程序 链状态异常，结束
	Crontab()
}

// func Stop() {
// 	os.Exit(1)
// }

func Crontab() {
	times := common.ComCfg.ServiceCfg.Times
	ticker := time.NewTicker(time.Second * 30) // 高度监控，30s 一次。
	// invokeTicker := time.NewTicker(time.Second * 1)             // 合约调用 1s 执行一次
	stoper := time.NewTimer(time.Minute * time.Duration(times)) //程序运行时间，到时退出

	go invokeContract()

	for {
		select {
		case <-stoper.C:
			return
		case <-ticker.C:
			common.GetSystemStatus() // 监控程序 链状态异常退出
		// case <-invokeTicker.C:
		// 	invokeContract() // 调用合约

		default:
			time.Sleep(time.Second * 3)
		}
	}
	// Stop:
	// 	Stop()
}

func invokeContract() {
	gos := common.ComCfg.ServiceCfg.Routines

	for i := 0; i < gos-1; i++ {
		node := common.ComCfg.XchainCfg.Nodes[i] // 并发大约节点数量会panic，暂时使用两个或者三个并发，大于三个节点。
		go doInvoke(node)
	}
}

func doInvoke(node string) {
	xclient, err := xuper.New(node)
	if err != nil {
		panic(err)
	}
	defer xclient.Close()

	// 确保 admin 有钱。
	admin, err := account.RetrieveAccount(common.AdminMnemonic, common.AdminLanguage)
	if err != nil {
		panic(err)
	}

	for { // 一直进行调用合约，参数随机生成。
		tx, err := xclient.InvokeEVMContract(admin, common.ComCfg.ContractCfg.Name, common.ComCfg.ContractCfg.Method, genRandomArgs())
		if err != nil {
			common.Sugar.Error("invoke failed:", err.Error())
		} else {
			common.Sugar.Infof("invoke ok: %x\n", tx.Tx.Txid)
		}
	}
}

var (
	_id uint64 = 1
	m   sync.Mutex
)

func genRandomArgs() map[string]string {
	m.Lock()
	defer m.Unlock()

	_id++
	_data := RandBytes(100)

	args := map[string]string{
		"_id":            strconv.FormatUint(_id, 10),
		"_initialSupply": "1",
		"_data":          string(_data),
		"tokenTime":      "100",
	}
	return args
}

const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandBytes(n int) []byte {
	b := make([]byte, n)
	for i, cache, remain := n-1, rand.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = rand.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return b
}
