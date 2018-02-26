package tool

import (
	"strings"
	"net/http"
	"io/ioutil"
	"regexp"
	"log"
	"encoding/json"
	"github.com/smallnest/goreq"
	"errors"
	"os"
	"sort"
	"fmt"
	"time"
	"io"
	"encoding/xml"
)

var (
	title=""
	aid = ""
	cid = ""
	sid=""
	av = ""
	mid=""
	epid=""
	bangumi=false
	referer_url=""
	user_agent="Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.84 Safari/537.36"
)

const(
	heart_sleep=15
)
func get_some_id(url string) {

	resp, err := http.Get(url)
	if err != nil {
		log.Println(err.Error)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	// log.Println(string(body))

	if strings.Index(url, "bangumi") >= 0 {
		bangumi=true
		//正则匹配吧
		reg := regexp.MustCompile(`"epInfo":{.*?}`)
		result := reg.Find(body)
		// log.Println(string(result))
		//去掉 "epInfo": 9个
		result = result[9:]
		epinfo := Bangumi_epinfo{}
		err = json.Unmarshal(result, &epinfo)
		// log.Println(epinfo)
		if err != nil {
			log.Println(err.Error)
			return
		}
		aid = B_tostring(epinfo.Aid)
		av = "av" + aid
		cid = B_tostring(epinfo.Cid)
		mid=B_tostring(epinfo.Mid)
		epid=B_tostring(epinfo.EpId)

		//"ssId":21603
		reg=regexp.MustCompile(`"ssId":[0-9]*`)
		result=reg.Find(body)
		result=result[7:]
		sid=string(result)
		log.Println(sid)
	} else {
		reg := regexp.MustCompile(`av([0-9]+)`)
		av = string(reg.Find([]byte(url)))
		aid = strings.TrimLeft(av, "av")
	}

	title_reg:=regexp.MustCompile(`<title>.*?</title>`)
	title=string(title_reg.Find(body))
	title=strings.TrimLeft(title,`<title>`)
	title=strings.TrimRight(title,`</title>`)

}
func B_get_cid() (string, error) {
	getcid_url := "https://api.bilibili.com/x/player/pagelist"
	last_url := B_build_url(getcid_url, map[string]string{
		"aid": strings.Replace(aid, "av", "", -1),
	})
	res_body := Get_Cid_Res{}
	resp, _, errs := goreq.New().Get(last_url).BindBody(&res_body).End()
	if errs != nil {
		return "", errors.New("https error")
	}
	if resp.StatusCode != http.StatusOK {
		return "", errors.New("https status code is not equal 200")
	}
	if res_body.Code != 0 {
		return "", errors.New(res_body.Message)
	}
	log.Println(res_body.Data)
	return B_tostring(res_body.Data[0].Cid), nil
}

//<d p="3.15700,1,25,16777215,1519036793,0,3c613191,4318544039">第一无误</d>
//p=视频中的相对时间戳,方向(滚动1,顶部5,底部4),字体大小,颜色,发送弹幕的时间戳,弹幕池,用户标识(用户ID的CRC32b加密),弹幕唯一id
type Xml_danmaku_Res struct {
	Chatserver string          `xml:"chatserver"`
	Chatid     int64           `xml:"chatid"`
	Mission    int64           `xml:"mission"`
	Maxlimit   int64           `xml:"maxlimit"`
	State      int64           `xml:"state"`
	Realname   string          `xml:"realname"`
	Source     string          `xml:"source"`
	D          []Xml_danmaku_d `xml:"d"`
}
type Xml_danmaku_d struct {
	P    string `xml:"p,attr"`
	Text string `xml:",chardata"`
}

func B_get_danmu() (string, error) {
	get_danmu_url := "https://comment.bilibili.com/%s.xml"
	last_url := fmt.Sprintf(get_danmu_url, cid)

	resp, err := http.Get(last_url)
	if err != nil {
		return "", errors.New("https error")
	}
	if resp.StatusCode != http.StatusOK {
		return "", errors.New("https status code is not equal 200")
	}
	// Content-Encoding: deflate
	res, err := B_flate_decode(resp.Body)
	if err != nil {
		return "", err
	}
	xml_res := Xml_danmaku_Res{}
	err = xml.Unmarshal(res, &xml_res)
	if err != nil {
		log.Println("弹幕 xml 解析错误")
		return "", errors.New("弹幕 xml 解析错误")
	}
	return string(res), nil
}

func B_send_flv_heart(length int64){
	/*
	//番剧的POST心跳包 https://api.bilibili.com/x/report/web/heartbeat
	type:4 //未知 (默认4)
	sub_type:1 //未知,暂时可以固定为1 (默认1)
	epid:183836 //番剧id (必填)
	aid:18221694 //视频id (必填)
	cid:29749244 //弹幕id (必填)
	sid:21603 //某个id (ssId必填)
	played_time:636 //当前视频的相对播放时间 (0开始,每次+15)
	start_ts:1519443810 //第一次请求心跳包的时间戳 (必填)
	realTime:29 //真实播放的时间,两个包相差也是15s (0开始,每次+15)
	csrf:b5d2b419cb964512e9ce423acabef2c0 //cookie里的bili_jct (默认为空)
	play_type:0 //未知 (默认0)
	mid:768525 //用户id... (默认为0)

	//普通视频的POST心跳包 https://api.bilibili.com/x/report/web/heartbeat
	play_type:0 //未知 0 1 都有 (默认0)
	type:3 //未知 (默认3)
	mid:768525 //用户id...(默认0)
	cid:32253539 //同上 (必填)
	aid:19780254 //同上 (必填)
	start_ts:1519443863 //同上 (必填)
	csrf:b5d2b419cb964512e9ce423acabef2c0 //同上 (默认空)
	played_time:240 //同上 (0开始,每次+15)
	realTime:42 //同上 (0开始,每次+15)
	*/
	last_url:="https://api.bilibili.com/x/report/web/heartbeat"
	form:=map[string]string{}
	if bangumi{
		form=map[string]string{
			"type":"4",
			"sub_type":"1",
			"epid":epid,
			"aid":aid,
			"cid":cid,
			"sid":sid,
			"played_time":"0",
			"realTime":"0",
			"start_ts": B_tostring(time.Now().Unix()),
			"csrf":"",
			"play_type":"0",
			"mid":"0",
		}
	}else{
		form=map[string]string{
		"play_type":"0",
			"type":"3",
			"aid":aid,
			"cid":cid,
			"played_time":"0",
			"realTime":"0",
			"start_ts": B_tostring(time.Now().Unix()),
			"csrf":"",
			"mid":"0",
		}
	}
	var play_time int64 =0
	var real_time int64 =0
	for;;{
		resp,_,_:=goreq.New().Post(last_url).ContentType("form").SendMapString(B_httpBuildQuery(form)).End()

		log.Println("heart status code : ",resp.StatusCode)

		time.Sleep(time.Second*heart_sleep)
		play_time+=heart_sleep;
		real_time+=heart_sleep;
		form["played_time"]=B_tostring(play_time)
		form["realTime"]=B_tostring(real_time)

		if play_time>=length{
			break
		}
	}
}

func B_get_flvurl() (string, error) {
	/*  //V2版本的方法 sign错误
	//base_url:="https://interface.bilibili.com/v2/playurl"
	//app_key:="f3bb208b3d081dc8"
	//app_secret:="1c15888dc316e05a15fdd0a02ed6584f"

	param:=map[string]string{
	"cid":cid,
	"appkey":app_key,
	"otype":"json",
	"type":"",
	"quality":"0",
	"qn":"0",
	}
	//抓取的正确的请求(html5的源)
	//resp,body,errs=goreq.New().Get("https://interface.bilibili.com/v2/playurl?cid=32253539&appkey=84956560bc028eb7&otype=json&type=&quality=0&qn=0&sign=0d1b3ad4dd857c060d28a31a05d13835").End()
	//低画质 无验证
	http://api.bilibili.com/playurl?&aid=19796564&page=1&platform=html5
	//目前使用 flash的源,似乎已经被限速了,参数rate=240000
	*/

	param := map[string]string{
		"cid":     cid,
		"player":  "1",
		"quality": "0",
		"ts":      B_tostring(time.Now().Unix()),
	}

	referer_url = fmt.Sprintf("https://www.bilibili.com/video/%s/", av)

	app_bangumi_secret := "9b288147e5474dd2aa67085f716c560d"
	app_normal_secret := "1c15888dc316e05a15fdd0a02ed6584f"

	last_url := ""
	if bangumi {
		base_url := "http://bangumi.bilibili.com/player/web_api/playurl"
		param["module"] = "bangumi"
		query, sign := B_EncodeSign(param, app_bangumi_secret)
		last_url = base_url + "?" + query + "&sign=" + sign
	} else {
		base_url := "http://interface.bilibili.com/playurl"
		query, sign := B_EncodeSign(param, app_normal_secret)
		last_url = base_url + "?" + query + "&sign=" + sign
	}
	log.Println("last url: ", last_url)

	/*
		//goreq 库存在bug(提issues后,作者已修复)
		resp,body,errs:=goreq.New().
		SetHeader("Accept-Encoding","identity").
		SetHeader("Host","interface.bilibili.com").
		SetHeader("Referer",fmt.Sprintf("https://www.bilibili.com/video/%s/",av)).
		SetHeader("User-Agent",user_agent).
		SetHeader("Connection","close").
		Get(last_url).End()
		if errs!=nil{
		}
	*/

	req, _ := http.NewRequest("GET", last_url, nil)
	req.Header.Add("Accept-Encoding", "identity")
	req.Header.Add("Referer", referer_url)
	req.Header.Add("User-Agent", user_agent)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {

	}
	if resp.StatusCode != http.StatusOK {

	}
	// 依旧是xml格式 不需要解压
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	// log.Println(string(body))


	return string(body), nil
}

func down_flv(rawxml string) error{

	xml_res := Xml_video_Res{}
	err := xml.Unmarshal([]byte(rawxml), &xml_res)
	if err != nil {
		log.Println("视频 xml 解析错误")
		return errors.New("视频 xml 解析错误")
	}

	/*
		//flash的源 rate=240000
		download:="http://upos-hz-mirrorkodo.acgvideo.com/upgcxcode/39/35/32253539/32253539-1-64.flv?e=ig8g5X10ugNcXBlqNxHxNEVE5XREto8KqJZHUa6m5J0SqE85tZvEuENvNC8xNEVE9EKE9IMvXBvE2ENvNCImNEVEK9GVqJIwqa80WXIekXRE9IMvXBvEuENvNCImNEVEua6m2jIxux0CkF6s2JZv5x0DQJZY2F8SkXKE9IB5to8euxZM2rNcNbUVhwdVhoM1hwdVhwdVNCM%3D&platform=pc&uipk=5&uipv=5&deadline=1519378528&gen=playurl&um_deadline=1519378528&rate=240000&um_sign=2c45de4e3d83baa30b56da7349634451&dynamic=1&os=kodo&oi=3030949926&upsig=66fd2d9e147c21fbbc0e9fedc2ea4eb3"
		//rate 应该是速率 大于0应该是被限速了
		//html的源 rete=0
		download="http://upos-hz-mirrorkodo.acgvideo.com/upgcxcode/39/35/32253539/32253539-1-64.flv?e=ig8g5X10ugNcXBlqNxHxNEVE5XREto8KqJZHUa6m5J0SqE85tZvEuENvNC8xNEVE9EKE9IMvXBvE2ENvNCImNEVEK9GVqJIwqa80WXIekXRE9IMvXBvEuENvNCImNEVEua6m2jIxux0CkF6s2JZv5x0DQJZY2F8SkXKE9IB5to8euxZM2rNcNbUVhwdVhoM1hwdVhwdVNCM%3D&platform=pc&uipk=5&uipv=5&deadline=1519383359&gen=playurl&um_deadline=1519383359&rate=0&um_sign=30d0e864f9f0072c029831a475d46529&dynamic=1&os=kodo&oi=3030949926&upsig=d4d0701dcb7704638aec8f7edfcd13e5"
	*/

	download_order := map[int64]*Xml_video_durl{}
	order := []int{}
	for i, _ := range xml_res.Durl {
		order = append(order, int(xml_res.Durl[i].Order))
		download_order[xml_res.Durl[i].Order] = &xml_res.Durl[i]
	}
	sort.Ints(order)

	//统计下载速度
	//go func() {
	//	var last int64 = 0
	//	for {
	//		fi, err := f.Stat()
	//		if err == nil {
	//			speed := (fi.Size() - last) / 1024
	//			log.Println(fmt.Sprintf("download : %d byte . speed : %d KB/s ", fi.Size(), speed))
	//			last = fi.Size()
	//		}
	//		time.Sleep(time.Second)
	//	}
	//}()

	//下载分段
	for i, v := range order {
		f, err := os.Create(fmt.Sprintf("%s_%d.flv",title,v))
		if err != nil {
			log.Println("create error")
			return errors.New("创建下载文件失败")
		}
		download := download_order[int64(v)].Url
		log.Println("download url: ", download)
		req, _ := http.NewRequest("GET", download, nil)
		req.Header.Add("Referer", referer_url)
		req.Header.Add("User-Agent", user_agent)
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Println("download error")
			return errors.New("下载错误")
		}
		io.Copy(f, resp.Body)
		f.Close()
		continue

		//todo 合并flv文件.
		func(){
			if i==0{
				io.Copy(f, resp.Body)
			}else{
				defer resp.Body.Close()


				flv_header:=make([]byte,9)
				//(前三个为FLV)(01表示版本)(05表示音频,视频都有)(09表示header长度)
				//46 4C 56 01 05 00 00 00 09
				//去掉 flv header
				resp.Body.Read(flv_header)

				//上一个tag的大小,4字节(包括tag_header)
				flv_body_tag_size:=make([]byte,4)
				//当前tag的类型,1字节 0x08音频tag 0x09视频tag 0x12脚本tag
				flv_body_tag_type:=make([]byte,1)
				//当前tag的长度,3字节
				flv_body_tag_len:=make([]byte,3)
				//当前tag的时间戳:3+1字节
				flv_body_tag_ts:=make([]byte,4)
				//streamid,3字节 目前总是为0
				flv_body_tag_stream:=make([]byte,3)

				resp.Body.Read(flv_body_tag_size)
				resp.Body.Read(flv_body_tag_type)
				resp.Body.Read(flv_body_tag_len)
				resp.Body.Read(flv_body_tag_ts)
				resp.Body.Read(flv_body_tag_stream)

				//当前tag的数据
				flv_body_tag_data:=make([]byte,Byte32Uint32(flv_body_tag_len,true))
				resp.Body.Read(flv_body_tag_data)

				//metadata里面通常为2个AMF包
				//第一个AMF包：
				//第1个字节表示AMF包类型，一般总是0x02，表示字符串，其他值表示意义请查阅文档
				//第2-3个字节为UI16类型值，表示字符串的长度，一般总是0x000A 即 'onMetaData'长度
				//后面字节为字符串数据，一般总为 'onMetaData'
				//第二个AMF包：
				//第1个字节表示AMF包类型，一般总是0x08，表示数组
				//第2-5个字节为UI32类型值，表示数组元素的个数
				//后面即为键值对
				//第1-2个字节表示元素名称的长度，假设为L
				//后面跟着为长度为L的字符串
				//L后面3个字节表示值类型(
				//后面为值的字节

				//忽略掉metadata的tag size
				resp.Body.Read(flv_body_tag_size)

			}
		}()
	}
	return nil
}

func Init(url string)error{
	var err error
	get_some_id(url)
	//通过av号获取cid
	if aid == "" {
		log.Println("获取aid失败")
		return errors.New("获取aid失败")
	}
	if cid == "" {
		cid, err = B_get_cid()
		if err != nil {
			return err
		}
	}
	log.Println("aid: ", aid, " cid: ", cid)
	return nil
}