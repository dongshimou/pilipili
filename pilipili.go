package main

import (
	"compress/flate"
	"compress/gzip"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/smallnest/goreq"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
	"./tool"
)

type Get_Res_Template struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
	TTL     int64  `json:"ttl"`
}

func B_build_url(base string, para map[string]string) string {
	res := base
	res += "?"
	for k, v := range para {
		res += k
		res += "="
		res += v
		res += "&"
	}
	res = strings.TrimRight(res, "&")
	return res
}
func B_tostring(v interface{}) string {
	return fmt.Sprintf("%v", v)
}

type Get_Cid_Res struct {
	Get_Res_Template
	Data []Get_Cid_ResData `json:"data"`
}
type Get_Cid_ResData struct {
	Cid      int64  `json:"cid"`
	Page     int64  `json:"page"`
	Form     string `json:"form"`
	Part     string `json:"part"`
	Duration int64  `json:"duration"`
	Vid      string `json:"vid"`
	WebLink  string `json:"weblink"`
}

func B_get_cid(aid string) (string, error) {

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

//deflate解压 Content-Encoding: deflate
func B_flate_decode(in io.ReadCloser) ([]byte, error) {
	reader := flate.NewReader(in)
	defer reader.Close()
	return ioutil.ReadAll(reader)
}

//gzip解压 Content-Encoding: Gzip
func B_Gzip_decode(in io.ReadCloser) ([]byte, error) {
	reader, err := gzip.NewReader(in)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return ioutil.ReadAll(reader)
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

func B_get_danmu(cid string) (string, error) {
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

//https://interface.bilibili.com/v2/playurl?cid=32253539&appkey=84956560bc028eb7&otype=json&type=&quality=0&qn=0&sign=0d1b3ad4dd857c060d28a31a05d13835
func B_httpBuildQuery(params map[string]string) string {
	list := make([]string, 0, len(params))
	buffer := make([]string, 0, len(params))
	for key := range params {
		list = append(list, key)
	}
	sort.Strings(list)
	for _, key := range list {
		value := params[key]
		buffer = append(buffer, key)
		buffer = append(buffer, "=")
		buffer = append(buffer, value)
		buffer = append(buffer, "&")
	}
	buffer = buffer[:len(buffer)-1]
	return strings.Join(buffer, "")
}
func B_EncodeSign(params map[string]string, secret string) (string, string) {
	queryString := B_httpBuildQuery(params)
	return queryString, B_Md5(queryString + secret)
}

func B_Md5(formal string) string {
	h := md5.New()
	h.Write([]byte(formal))
	return hex.EncodeToString(h.Sum(nil))
}

type Xml_video_Res struct {
	Result        string `xml:"result"`
	Timelength    int64  `xml:"timelength"`
	Format        string `xml:"format"`
	AcceptFormat  string `xml:"accept_format"`
	AcceptQuality string `xml:"accept_quality"`
	Quality       string `xml:"quality"`
	From          string `xml:"from"`
	SeekParam     string `xml:"seek_param"`
	SeekType      string `xml:"seek_type"`
	//番剧
	Bp        string           `xml:"bp"`
	VipStatus string           `xml:"vip_status"`
	VipType   string           `xml:"vip_type"`
	HasPaid   string           `xml:"has_paid"`
	Status    string           `xml:"status"`
	Durl      []Xml_video_durl `xml:"durl"`
}
type Xml_video_durl struct {
	Order  int64 `xml:"order"`
	Length int64 `xml:"length"`
	Size   int64 `xml:"size"`
	Xml_video_url
	BackupUrl []Xml_video_url `xml:"backup_url"`
}
type Xml_video_url struct {
	Url string `xml:"url"`
}

func B_get_flvurl(cid, av string, bangumi bool) (string, error) {
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

	referer_url := fmt.Sprintf("https://www.bilibili.com/video/%s/", av)
	user_agent := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.84 Safari/537.36"
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
		//goreq 库存在bug
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

func B_download_flv(av,title,flvxml string){
	user_agent := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.84 Safari/537.36"
	referer_url := fmt.Sprintf("https://www.bilibili.com/video/%s/", av)

	xml_res := Xml_video_Res{}
	err := xml.Unmarshal([]byte(flvxml), &xml_res)
	if err != nil {
		log.Println("视频 xml 解析错误")
		return
	}

	downflv := func() {
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

		go tool.B_send_flv_heart(xml_res.Timelength)

		//下载分段并合并
		for _, v := range order {
			f, err := os.Create(fmt.Sprintf("mv_%d.flv",v))
			if err != nil {
				log.Println("create error")
				return
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
				return
			}
			io.Copy(f, resp.Body)
			f.Close()
		}
	}

	downflv()
}

func B_get_cookie(cid,aid string){
	//获取cookies
	cookie_url:=fmt.Sprintf("http://interface.bilibili.com/player?id=cid:%s&aid=%s",cid,aid)
	log.Println(cookie_url)
}

type Bangumi_epinfo struct {
	Aid           int64  `json:"aid"`
	Cid           int64  `json:"cid"`
	Cover         string `json:"cover"`
	EpId          int64  `json:"ep_id"`
	EpisodeStatus int64  `json:"episode_status"`
	From          string `json:"from"`
	Index         string `json:"index"`
	IndexTitle    string `json:"index_title"`
	Mid           int64  `json:"mid"`
	Page          int64  `json:"page"`
	Vid           string `json:"vid"`
}

func main() {

	url := "https://www.bilibili.com/video/av16968840/?spm_id_from=333.334.bili_douga.9"
	//url = "https://www.bilibili.com/bangumi/play/ep183836"
	aid := ""
	cid := ""
	sid:=""
	av := ""
	mid:=""
	epid:=""
	title:=""
	bangumi:=false
	var err error
	get_av := func(url string) {

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

	get_av(url)
	//通过av号获取cid
	if aid == "" {
		log.Println("获取aid失败")
		return
	}
	if cid == "" {
		cid, err = B_get_cid(av)
		if err != nil {
			return
		}
	}
	log.Println("aid: ", aid, " cid: ", cid)
	//发送心跳包 防止限速


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


	//通过cid获取弹幕xml
	danmu, err := B_get_danmu(cid)
	if err != nil {
		return
	}
	log.Println(danmu)
	//获取视频源地址失败
	flvurl, err := B_get_flvurl(cid, av,bangumi )
	if err != nil {
		return
	}
	log.Println(flvurl)

	B_download_flv(av,title,flvurl)

}
