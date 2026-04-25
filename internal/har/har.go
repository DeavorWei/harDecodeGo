package har

// HAR 文件根结构
type HAR struct {
	Log Log `json:"log"`
}

type Log struct {
	Version string  `json:"version"`
	Creator Creator `json:"creator"`
	Browser Browser `json:"browser,omitempty"`
	Pages   []Page  `json:"pages,omitempty"`
	Entries []Entry `json:"entries"`
}

type Creator struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type Browser struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Page 页面信息
type Page struct {
	StartedDateTime string      `json:"startedDateTime"`
	ID              string      `json:"id"`
	Title           string      `json:"title"`
	PageTimings     PageTimings `json:"pageTimings"`
}

// PageTimings 页面加载时间
type PageTimings struct {
	OnContentLoad float64 `json:"onContentLoad,omitempty"`
	OnLoad        float64 `json:"onLoad,omitempty"`
}

type Entry struct {
	StartedDateTime string   `json:"startedDateTime"`
	Request         Request  `json:"request"`
	Response        Response `json:"response"`
	Cache           Cache    `json:"cache"`
	Timings         Timings  `json:"timings"`
	Time            float64  `json:"time"`
	Pageref         string   `json:"pageref,omitempty"`
}

type Request struct {
	Method      string       `json:"method"`
	URL         string       `json:"url"`
	HTTPVersion string       `json:"httpVersion"`
	Headers     []Header     `json:"headers"`
	Cookies     []Cookie     `json:"cookies"`
	QueryString []QueryParam `json:"queryString"`
	BodySize    int          `json:"bodySize"`
}

type Response struct {
	Status      int      `json:"status"`
	StatusText  string   `json:"statusText"`
	HTTPVersion string   `json:"httpVersion"`
	Headers     []Header `json:"headers"`
	Cookies     []Cookie `json:"cookies"`
	Content     Content  `json:"content"`
	RedirectURL string   `json:"redirectURL"`
	HeadersSize int      `json:"headersSize"`
	BodySize    int      `json:"bodySize"`
}

type Content struct {
	MimeType string `json:"mimeType"`
	Size     int    `json:"size"`
	Text     string `json:"text,omitempty"`
	Encoding string `json:"encoding,omitempty"`
}

type Header struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Cookie struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	Path     string `json:"path,omitempty"`
	Domain   string `json:"domain,omitempty"`
	Expires  string `json:"expires,omitempty"`
	HTTPOnly bool   `json:"httpOnly,omitempty"`
	Secure   bool   `json:"secure,omitempty"`
}

type QueryParam struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Cache 缓存信息
type Cache struct {
	BeforeCache *CacheEntry `json:"beforeRequest,omitempty"`
	AfterCache  *CacheEntry `json:"afterRequest,omitempty"`
}

// CacheEntry 缓存条目
type CacheEntry struct {
	Expires    string `json:"expires,omitempty"`
	LastAccess string `json:"lastAccess"`
	ETag       string `json:"eTag"`
	HitCount   int    `json:"hitCount"`
}

// Timings 请求各阶段耗时
type Timings struct {
	Blocked float64 `json:"blocked,omitempty"`
	DNS     float64 `json:"dns,omitempty"`
	Connect float64 `json:"connect,omitempty"`
	Send    float64 `json:"send,omitempty"`
	Wait    float64 `json:"wait,omitempty"`
	Receive float64 `json:"receive,omitempty"`
	SSL     float64 `json:"ssl,omitempty"`
}
