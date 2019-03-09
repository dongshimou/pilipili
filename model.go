package pilipili

type piliVidioPart struct {
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
	vidios      []piliVidioPart
}

type piliBangumiEpInfo struct {
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

type piliGetCidRes struct {
	piliGetResTemplate
	Data []piliGetCidResData `json:"data"`
}
type piliGetCidResData struct {
	Cid      int64  `json:"cid"`
	Page     int64  `json:"page"`
	Form     string `json:"form"`
	Part     string `json:"part"`
	Duration int64  `json:"duration"`
	Vid      string `json:"vid"`
	WebLink  string `json:"weblink"`
}
type piliGetResTemplate struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
	TTL     int64  `json:"ttl"`
}

type piliXmlVideoRes struct {
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
	Bp        string             `xml:"bp"`
	VipStatus string             `xml:"vip_status"`
	VipType   string             `xml:"vip_type"`
	HasPaid   string             `xml:"has_paid"`
	Status    string             `xml:"status"`
	Durl      []piliXmlVideoDurl `xml:"durl"`
}
type piliXmlVideoDurl struct {
	Order  int64 `xml:"order"`
	Length int64 `xml:"length"`
	Size   int64 `xml:"size"`
	piliXmlVideourl
	BackupUrl []piliXmlVideourl `xml:"backup_url"`
}
type piliXmlVideourl struct {
	Url string `xml:"url"`
}
