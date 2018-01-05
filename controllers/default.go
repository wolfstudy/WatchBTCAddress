package controllers

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/astaxie/beego"
	"github.com/tidwall/gjson"
	"go4.org/sort"
)

type BitcoinSite struct {
	Site string
	T    string
}

type TxInfo struct {
	//是否为输入的钱数
	IsIN     bool
	//交易是否被更新过
	Updated  bool
	//交易类型（BTC && BCH）
	Type     string
	//拼接url前缀，调用API文档
	TxPrefix string
	//地址
	AdPrefix string
	//tx_hash的值
	TxID     string
	//交易日期
	Date     string
	//交易钱数
	Amount   float64
}

//需要监控的BTC && BCH的地址，通过地址获取btc.com的数据。
var siteInfo []BitcoinSite = []BitcoinSite{
	{"1Nh7uHdvY6fNwtQtM1G5EZAFPLC33B59rB", "BTC"},
	{"1HaPVJr5pf2UAxKvn3LRHhQqNPF8Pf28L3", "BTC"},
	{"3CnzuFFbtgVyHNiDH8BknGo3PQ3dpdThgJ", "BCH"},
	{"18XtqyB9iehZ5ApdA98C4L72mvQACpvHDm", "BCH"},
	{"1Lj6KHuyGHXuGHPbPkWzsnTXXnzHLesGAo", "BCH"},
}

type MainController struct {
	beego.Controller
}

func (c *MainController) Get() {
	var info = make(map[string][]TxInfo)

	length := len(siteInfo)
	for i := 0; i < length; i++ {
		value := siteInfo[i]
		var prefix string
		//内部端口（BTC && BCH）
		if value.T == "BTC" {
			prefix = "https://chain.api.btc.com/v3/address/"
		} else {
			prefix = "https://bch-chain.api.btc.com/v3/address/"
		}

		//接收获取到的请求，是一个string的字符串
		content := getContet(prefix + value.Site)

		//gjson解析获取到的json串
		hash := gjson.Get(content, "data.list.#.hash").Array()
		amount := gjson.Get(content, "data.list.#.inputs_value").Array()
		created := gjson.Get(content, "data.list.#.created_at").Array()
		inputs := gjson.Get(content, "data.list.#.inputs").Array()

		//循环遍历所有的hash值
		for index, iterm := range hash {
			txinfo := TxInfo{}
			if value.T == "BTC" {
				txinfo.TxPrefix = "https://btc.com/"
				txinfo.AdPrefix = "https://btc.com/"
				txinfo.Type = "BTC"
			} else {
				txinfo.TxPrefix = "https://bch.btc.com/"
				txinfo.AdPrefix = "https://bch.btc.com/"
				txinfo.Type = "BCH"
			}
			//money数量
			txinfo.Amount = amount[index].Float() / float64(100000000)
			//时间戳转时间
			tm := time.Unix(created[index].Int(), 0)
			txinfo.Date = tm.Format("2006-01-02 15:04:05")
			txinfo.TxID = iterm.String()
			//判断是否为最新交易
			if created[index].Int() > 151489440 {
				http.Get("http://v.juhe.cn/sms/send?mobile=18410205081&tpl_id=嫁给我！&tpl_value=%23code%23%3D654654&key=381489926eb25231b5c3468683d32338")
				txinfo.Updated = true
			}
			//input adjust
			if(!strings.Contains(inputs[index].String(), value.Site)){
				txinfo.IsIN = true
			}

			info[value.Site] = append(info[value.Site], txinfo)
		}
	}

	//用时间+地址拼接为一个字符串，利用时间的有序性，排序。
	address := make([]string, 0)
	for key, iterm := range info {
		address = append(address, iterm[0].Date+key)
	}

	sort.Strings(address)
	l := len(address)
	str := make([]string, 0)

	//排好序之后，截取后面的地址，这样做的目的是为了规避map结构的无序性。
	for i := l - 1; i >= 0; i-- {
		str = append(str, address[i][19:])
	}

	c.Data["str"] = str
	c.Data["lists"] = info
	c.TplName = "index.html"
}

//获取btc.com网站的数据。
func getContet(url string) string {
	get, err := http.Get(url + "/tx")
	if err != nil {
		fmt.Println(err)
		return ""
	}
	defer get.Body.Close()
	content, _ := ioutil.ReadAll(get.Body)
	return string(content)
}
