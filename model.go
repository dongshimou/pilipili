package main

type pili_vidio_part struct {
	cid       string
	page      int64
	part_name string
	dur       int64
}
type pilipili struct {
	title       string
	aid         string
	cid         string
	sid         string
	av          string
	mid         string
	epid        string
	bangumi     bool
	referer_url string
	pili_err    error
	file_name   string
	vidio_index int
	vidios      []pili_vidio_part
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
type Get_Res_Template struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
	TTL     int64  `json:"ttl"`
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
