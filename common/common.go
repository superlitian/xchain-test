package common

import (
	"fmt"
	"github.com/xuperchain/xuper-sdk-go/v2/account"
	"github.com/xuperchain/xuper-sdk-go/v2/xuper"
	"io/ioutil"
	"math/big"
	"math/rand"
	"path/filepath"
	"strconv"
	"time"
)

const (
	ContractPath = "./build"

	AdminMnemonic = "豆 遭 牲 缺 译 尼 皆 缆 砖 柬 耳 造"
	AdminLanguage = 1

	TimeSpan = "timespan"
	Admin    = "admin"
	Increase = "increase"

	AdminAddress = "VVvz9eujCeKtCFuGKnEUeXFQa7uFmdncX"
	MinnerPath   = "./accounts/minner/"
)

var (
	// 此次测试的账户数组
	accounts []*account.Account
	// 是否需要补充账户余额
	notEnough = false
	// 矿工账号
	minner *account.Account
	// 管理员账号
	admin *account.Account
	// xuper client
	client *xuper.XClient
	err    error
	// 自增字段计数器数组
	increases map[string]int64
	//合约账号
	contractAccount string
	// 合约类型
	module string
	// 合约名称
	name string
	// 合约方法
	method string
)

func BeforeDeploy() {
	minner, err = account.GetAccountFromPlainFile(MinnerPath)
	if err != nil {
		Sugar.Fatalf("get minner account error: %v", err)
	}
	admin, err = account.RetrieveAccount(AdminMnemonic, AdminLanguage)
	if err != nil {
		Sugar.Fatalf("get admin account error: %v", err)
	}
	client, err = xuper.New(ComCfg.XchainCfg.Nodes[0])
	// 给管理员转账
	if ComCfg.XchainCfg.IsNeedTransferToAdmin {
		if err != nil {
			Sugar.Fatalf("new xuper client error: %v", client)

		}
		tx, err := client.Transfer(minner, AdminAddress, "10000000000000")
		if err != nil {
			Sugar.Fatalf("transfer to admin error: %v", err)
		}
		Sugar.Infow("transfer to admin success", "txid", fmt.Sprintf("%x", tx.Tx.Txid), "amount", "10000000000000")
	}
	contractAccount = ComCfg.XchainCfg.ContractAccount
	if ComCfg.XchainCfg.IsNeedCreateContractAccount {
		tx, err := client.CreateContractAccount(admin, contractAccount)
		if err != nil {
			Sugar.Fatalf("create contract account error: %v", err)
		}
		Sugar.Infow("create contract account success", "txid", fmt.Sprintf("%x", tx.Tx.Txid))
	}
	// 给合约账户转账
	tx, err := client.Transfer(minner, contractAccount, "100000")
	if err != nil {
		Sugar.Fatalf("transfer to contract account error: %v", err)
	}
	Sugar.Infow("transfer to contract account success", "txid", fmt.Sprintf("%x", tx.Tx.Txid), "amount", "100000")
	// 合约类型
	module = ComCfg.ContractCfg.Module
	// 合约名称
	name = ComCfg.ContractCfg.Name
	method = ComCfg.ContractCfg.Method
	// increases
	increases = make(map[string]int64)
	Sugar.Info("before deploy success")
}

// Deploy 部署合约
func Deploy() {
	// 初始化参数
	var codePath string
	var code []byte
	var bin []byte

	// node
	var node string
	node = ComCfg.XchainCfg.Nodes[0]
	client, err = xuper.New(node)

	// code
	if module == "evm" {
		codePath = filepath.Join(ContractPath, name+".abi")
		// abi
		code, err = ioutil.ReadFile(codePath)
		if err != nil {
			Sugar.Fatalf("get %v abi error: %v\n", module, err)
		}
		// bin
		codePath = filepath.Join(ContractPath, name+".bin")
		bin, err = ioutil.ReadFile(codePath)
		if err != nil {
			Sugar.Fatalf("get %v bin error: %v\n", module, err)
		}
	} else if module == "wasm" {
		codePath = filepath.Join(ContractPath, name+"."+module)
		code, err = ioutil.ReadFile(codePath)
		if err != nil {
			Sugar.Fatalf("get %v code error: %v\n", module, err)
		}
	} else if module == "native" {
		codePath = filepath.Join(ContractPath, name+"."+module)
		code, err = ioutil.ReadFile(codePath)
		if err != nil {
			Sugar.Fatalf("get %v code error: %v\n", module, err)
		}
	}
	err = admin.SetContractAccount(contractAccount)
	if err != nil {
		Sugar.Fatalf("set contract account error: %v\n", err)
	}

	args, err := loadContractParams(ComCfg.ContractCfg.Initialize)
	if err != nil {
		Sugar.Fatalf("load params error: %v", err)
	}

	var tx *xuper.Transaction

	switch module {
	case "evm":
		tx, err = client.DeployEVMContract(admin, name, code, bin, args)
		break
	case "wasm":
		tx, err = client.DeployWasmContract(admin, name, code, args)
		break
	case "native":
		tx, err = client.DeployNativeGoContract(admin, name, code, args)
	}
	if err != nil {
		Sugar.Fatalf("deploy %v contract error: %v", module, err)
	}
	Sugar.Infow("deploy contract success", "txid", fmt.Sprintf("%x", tx.Tx.Txid), "response", fmt.Sprintf("%s", tx.ContractResponse.Body))
}

func BeforeInvoke() {
	// 生成此次调用需要的账号
	nums := ComCfg.AccountCfg.Nums
	for i := 0; i < nums; i++ {
		acc, err := account.CreateAccount(1, 1)
		if err != nil {
			Sugar.Fatalf("create test account error: %v", err)
		}
		accounts = append(accounts, acc)
	}
	// 给此次调用需要的账号转账
	err := transferToAccounts()
	if err != nil {
		Sugar.Fatalf("before invoke faild: %v", err)
	}
	Sugar.Info("before invoke success")
}

func Invoke() {

	// node
	var node string
	if ComCfg.XchainCfg.IsSingle {
		node = ComCfg.XchainCfg.Nodes[0]
	} else {
		// 多节点随机选择
		node = getNode()
	}
	client, err = xuper.New(node)
	if err != nil {
		Sugar.Fatalf("new xuper client %v error: %v", node, err)
	}

	//account
	var account *account.Account
	account = getAccount()

	// params
	args, err := loadContractParams(ComCfg.ContractCfg.Params)
	if err != nil {
		Sugar.Fatalf("load params error: %v", err)
	}

	var tx *xuper.Transaction

	switch module {
	case "evm":
		tx, err = client.InvokeEVMContract(account, name, method, args)
		break
	case "wasm":
		tx, err = client.InvokeWasmContract(account, name, method, args)
		break
	case "native":
		tx, err = client.InvokeNativeContract(admin, name, method, args)
	}
	if err != nil {
		Sugar.Fatalf("invoke %v contract error: %v", module, err)
	}
	Sugar.Infow("invoke contract success", "txid", fmt.Sprintf("%x", tx.Tx.Txid), "response", fmt.Sprintf("%s", tx.ContractResponse.Body))
}

func transferToAccounts() error {
	// todo 不够时从矿工地址拿回一些
	if notEnough {

	}
	// 采用平均方式分配
	balance, err := client.QueryBalance(admin.Address)
	if err != nil {
		return err
	}
	amount := balance.Div(balance, big.NewInt(int64(len(accounts))))
	for i := 0; i < len(accounts); i++ {
		tx, err := client.Transfer(admin, accounts[i].Address, amount.String())
		if err != nil {
			Sugar.Fatalf("transfer to %s error", accounts[i].Address)
		}
		Sugar.Infow("transfer success", "to", accounts[i].Address, "amount", amount, "txid", fmt.Sprintf("%x", tx.Tx.Txid))
	}
	return nil
}

func getNode() string {
	index := rand.Intn(len(ComCfg.XchainCfg.Nodes))
	return ComCfg.XchainCfg.Nodes[index]
}

func getAccount() *account.Account {
	index := rand.Intn(len(accounts))
	return accounts[index]
}

// loadContractParams 解析 contract params
func loadContractParams(params map[string]string) (map[string]string, error) {
	// 关键字处理
	for k, v := range params {
		if v == TimeSpan {
			params[k] = getTimeSpan()
		}
		if v == Increase {
			params[k] = getIncrease(params[k])
		}
		if v == Admin {
			params[k] = getAdmin()
		}
	}

	return params, nil
}

func getAdmin() string {
	if module == "evm" {
		addr, _, _ := account.XchainToEVMAddress(admin.Address)
		return addr
	}
	return admin.Address
}

func getTimeSpan() string {
	return strconv.FormatInt(time.Now().Unix(), 10)
}

func getIncrease(key string) string {
	isNew := true
	for k, v := range increases {
		if k == key {
			increases[k] = v + 1
			isNew = false
		}
	}
	if isNew || increases == nil {
		increases[key] = 1
	}
	return strconv.FormatInt(increases[key], 10)
}

func GetSystemStatus() {
	var lastHeight int64
	client, err = xuper.New(ComCfg.XchainCfg.Nodes[0])
	tx, err := client.QuerySystemStatus()
	if err != nil {
		Sugar.Fatalf("query system status error: %v", err)
	} else {
		height := tx.SystemsStatus.BcsStatus[0].Block.Height
		if lastHeight == height {
			Sugar.Fatalw("the system status error", "height", height)
		}
		lastHeight = height
	}
}
