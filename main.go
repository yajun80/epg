package main

import (
    "encoding/xml"
    "fmt"
    "io"
    "log"
    "net/http"
    "strings"
    "sync"
    "time"

    "github.com/tidwall/gjson"
)

// XMLTV 结构定义 - 修正为正确的XML标签
type TV struct {
    XMLName   xml.Name `xml:"tv"`                       // 根标签是 tv
    Generator string   `xml:"generator-info-name,attr"` // 属性名 generator-info-name
    URL       string   `xml:"generator-info-url,attr"`  // 属性名 generator-info-url
    Channels  []Channel  `xml:"channel"`                 // channel标签
    Programmes []Programme `xml:"programme"`             // programme标签
}

type Channel struct {
    ID      string      `xml:"id,attr"`          // id属性
    Display DisplayName `xml:"display-name"`      // display-name子标签
}

type DisplayName struct {
    Lang  string `xml:"lang,attr"`   // lang属性
    Value string `xml:",chardata"`    // 标签内容
}

type Programme struct {
    Channel string `xml:"channel,attr"` // channel属性
    Start   string `xml:"start,attr"`   // start属性
    Stop    string `xml:"stop,attr"`    // stop属性
    Title   Title  `xml:"title"`        // title子标签
}

type Title struct {
    Lang  string `xml:"lang,attr"`   // lang属性
    Value string `xml:",chardata"`    // 标签内容
}

// 频道配置结构
type ChannelConfig struct {
    Name   string // 频道显示名称
    Source string // 数据源: cntv 或 migu
    ID     string // 源ID
}

// ==================== 频道配置（共187个）====================
var channels = []ChannelConfig{
    // ---------- 1. CNTV源 - 央视全套（23个）----------
    {"CCTV1综合", "cntv", "cctv1"},
    {"CCTV2财经", "cntv", "cctv2"},
    {"CCTV3综艺", "cntv", "cctv3"},
    {"CCTV4中文国际", "cntv", "cctv4"},
    {"CCTV5体育", "cntv", "cctv5"},
    {"CCTV5+体育赛事", "cntv", "cctv5plus"},
    {"CCTV6电影", "cntv", "cctv6"},
    {"CCTV7国防军事", "cntv", "cctv7"},
    {"CCTV8电视剧", "cntv", "cctv8"},
    {"CCTV9纪录", "cntv", "cctvjilu"},
    {"CCTV10科教", "cntv", "cctv10"},
    {"CCTV11戏曲", "cntv", "cctv11"},
    {"CCTV12社会与法", "cntv", "cctv12"},
    {"CCTV13新闻", "cntv", "cctv13"},
    {"CCTV14少儿", "cntv", "cctvchild"},
    {"CCTV15音乐", "cntv", "cctv15"},
    {"CCTV17农业农村", "cntv", "cctv17"},
    {"CCTV4欧洲", "cntv", "cctveurope"},
    {"CCTV4美洲", "cntv", "cctvamerica"},
    {"CCTV16奥林匹克", "cntv", "cctv16"},
    {"CCTV8K超高清", "cntv", "cctv8k"},
    {"CGTN", "cntv", "cgtn"},
    {"CGTN纪录", "cntv", "cgtndocumentary"},

    // ---------- 2. 咪咕源 - 全国卫视（39个）----------
    {"湖南卫视", "migu", "HUNAN"},
    {"浙江卫视", "migu", "ZJTV"},
    {"江苏卫视", "migu", "JSTV"},
    {"东方卫视", "migu", "DFTV"},
    {"北京卫视", "migu", "BTV1"},
    {"广东卫视", "migu", "GDTV"},
    {"深圳卫视", "migu", "SZTV"},
    {"天津卫视", "migu", "TJTV"},
    {"山东卫视", "migu", "SDTV"},
    {"安徽卫视", "migu", "AHTV"},
    {"湖北卫视", "migu", "HBTV"},
    {"四川卫视", "migu", "SCTV"},
    {"重庆卫视", "migu", "CQTV"},
    {"黑龙江卫视", "migu", "HLJTV"},
    {"辽宁卫视", "migu", "LNTV"},
    {"江西卫视", "migu", "JXTV"},
    {"东南卫视", "migu", "FJTV"},
    {"广西卫视", "migu", "GXTV"},
    {"海南卫视", "migu", "HAINANTV"},
    {"贵州卫视", "migu", "GZTV"},
    {"云南卫视", "migu", "YNTV"},
    {"河北卫视", "migu", "HEBTV"},
    {"河南卫视", "migu", "HENANTV"},
    {"山西卫视", "migu", "SXTV"},
    {"陕西卫视", "migu", "SXTVS"},
    {"甘肃卫视", "migu", "GSTV"},
    {"宁夏卫视", "migu", "NXTV"},
    {"新疆卫视", "migu", "XJTV"},
    {"内蒙古卫视", "migu", "NMGTV"},
    {"西藏卫视", "migu", "XZTV"},
    {"青海卫视", "migu", "QHTV"},
    {"吉林卫视", "migu", "JILINTV"},
    {"厦门卫视", "migu", "XIAMENTV"},
    {"兵团卫视", "migu", "BTTV"},
    {"三沙卫视", "migu", "SSTV"},
    {"延边卫视", "migu", "YBTY"},
    {"陕西农林卫视", "migu", "SNTV"},
    {"大湾区卫视", "migu", "GREATBAY"},
    {"澳亚卫视", "migu", "MAS"},

    // ---------- 3. 数字付费频道（66个）----------
    {"金鹰卡通", "migu", "JYKT"},
    {"卡酷少儿", "migu", "KAKU"},
    {"嘉佳卡通", "migu", "JIAJIA"},
    {"优漫卡通", "migu", "YOUMAN"},
    {"炫动卡通", "migu", "XUANDONG"},
    {"动画放映厅", "migu", "DONGMAN"},
    {"CCTV第一剧场", "migu", "CCTVDYJC"},
    {"CCTV风云剧场", "migu", "CCTVFYJC"},
    {"CCTV怀旧剧场", "migu", "CCTVHJJC"},
    {"CCTV风云音乐", "migu", "CCTVFYYY"},
    {"CCTV风云足球", "migu", "CCTVFYZQ"},
    {"CCTV高尔夫网球", "migu", "CCTVGEF"},
    {"CCTV世界地理", "migu", "CCTVSJDL"},
    {"CCTV女性时尚", "migu", "CCTVNXSS"},
    {"CCTV央视文化精品", "migu", "CCTVWHJP"},
    {"CCTV兵器科技", "migu", "CCTVBQKJ"},
    {"CCTV卫生健康", "migu", "CCTVWSJK"},
    {"CCTV梨园", "migu", "CCTVLiyuan"},
    {"CCTV台球", "migu", "CCTVTQ"},
    {"CCTV发现之旅", "migu", "CCTVFXL"},
    {"CCTV中学生", "migu", "CCTVZXS"},
    {"CCTV老故事", "migu", "CCTVLAOGUSHI"},
    {"CCTV新科动漫", "migu", "CCTVXKDM"},
    {"CCTV国防军事", "migu", "CCTVGFJS"},
    {"CHC家庭影院", "migu", "CHC"},
    {"CHC动作电影", "migu", "CHCDZ"},
    {"CHC高清电影", "migu", "CHCGQ"},
    {"第一剧场", "migu", "DYJC"},
    {"动作电影", "migu", "DZDY"},
    {"家庭剧场", "migu", "JTJC"},
    {"经典电影", "migu", "JDDY"},
    {"新片大片", "migu", "XPDP"},
    {"探索发现", "migu", "TSFX"},
    {"地理中国", "migu", "DLZG"},
    {"人文历史", "migu", "RWLS"},
    {"军事纪实", "migu", "JSJS"},
    {"科技之光", "migu", "KJZG"},
    {"法治天地", "migu", "FZTD"},
    {"劲爆体育", "migu", "JBTY"},
    {"天元围棋", "migu", "TYWQ"},
    {"快乐垂钓", "migu", "KLCD"},
    {"四海钓鱼", "migu", "SHDY"},
    {"汽摩频道", "migu", "QMPD"},
    {"冰雪体育", "migu", "BXTY"},
    {"赛事直播", "migu", "SSZB"},
    {"茶频道", "migu", "CHAPD"},
    {"美食频道", "migu", "MSPD"},
    {"靓妆频道", "migu", "LZPD"},
    {"魅力时尚", "migu", "MLSS"},
    {"居家生活", "migu", "JUSH"},
    {"收藏天下", "migu", "SCTX"},
    {"书画频道", "migu", "SHPDA"},
    {"摄影频道", "migu", "SYPD"},
    {"音像世界", "migu", "YXSJ"},
    {"早期教育", "migu", "ZQJY"},
    {"家政频道", "migu", "JZPD"},
    {"环球旅游", "migu", "HQTV"},
    {"世界地理", "migu", "SJDL"},
    {"全纪实", "migu", "QJS"},
    {"文物宝库", "migu", "WWBK"},
    {"音乐之声", "migu", "YYZS"},
    {"经典音乐", "migu", "JDYY"},
    {"广场舞", "migu", "GCW"},
    {"梨园频道", "migu", "LYPD"},
    {"戏曲频道", "migu", "XQPD"},

    // ---------- 4. 港澳台频道（32个）----------
    {"凤凰中文", "migu", "FENGHUANG"},
    {"凤凰资讯", "migu", "FHZX"},
    {"凤凰香港", "migu", "FHHK"},
    {"凤凰电影", "migu", "FHDY"},
    {"凤凰卫视", "migu", "FHTV"},
    {"TVBS", "migu", "TVBS"},
    {"TVBS新闻", "migu", "TVBSNEWS"},
    {"TVBS欢乐", "migu", "TVBSHL"},
    {"TVBS精彩", "migu", "TVBSJC"},
    {"TVBS亚洲", "migu", "TVBSYZ"},
    {"中天新闻", "migu", "CTINEWS"},
    {"中天综合", "migu", "CTIZH"},
    {"中天娱乐", "migu", "CTIYL"},
    {"中天亚洲", "migu", "CTIYZ"},
    {"东森新闻", "migu", "ETTVNEWS"},
    {"东森综合", "migu", "ETTVZH"},
    {"东森电影", "migu", "ETTVDY"},
    {"东森洋片", "migu", "ETTYYP"},
    {"东森戏剧", "migu", "ETTVXJ"},
    {"东森超视", "migu", "ETTCS"},
    {"纬来日本", "migu", "WLJP"},
    {"纬来体育", "migu", "WLTY"},
    {"纬来综合", "migu", "WLZH"},
    {"纬来电影", "migu", "WLDY"},
    {"纬来戏剧", "migu", "WLXJ"},
    {"三立台湾", "migu", "SLTV"},
    {"三立新闻", "migu", "SLNEWS"},
    {"三立都会", "migu", "SLDH"},
    {"八大综合", "migu", "BADA"},
    {"八大第一", "migu", "BADA1"},
    {"八大戏剧", "migu", "BADAXJ"},
    {"八大娱乐", "migu", "BDAYL"},

    // ---------- 5. 咪咕源 - 央视备份（19个）----------
    {"CCTV1", "migu", "CCTV1"},
    {"CCTV2", "migu", "CCTV2"},
    {"CCTV3", "migu", "CCTV3"},
    {"CCTV4", "migu", "CCTV4"},
    {"CCTV5", "migu", "CCTV5"},
    {"CCTV6", "migu", "CCTV6"},
    {"CCTV7", "migu", "CCTV7"},
    {"CCTV8", "migu", "CCTV8"},
    {"CCTV9", "migu", "CCTV9"},
    {"CCTV10", "migu", "CCTV10"},
    {"CCTV11", "migu", "CCTV11"},
    {"CCTV12", "migu", "CCTV12"},
    {"CCTV13", "migu", "CCTV13"},
    {"CCTV14", "migu", "CCTV14"},
    {"CCTV15", "migu", "CCTV15"},
    {"CCTV17", "migu", "CCTV17"},
    {"CCTV5+", "migu", "CCTV5PLUS"},
    {"CGTN", "migu", "CGTN"},
    {"CGTN西班牙语", "migu", "CGTNES"},

    // ---------- 6. 教育类（8个）----------
    {"中国教育1", "migu", "CETV1"},
    {"中国教育2", "migu", "CETV2"},
    {"中国教育3", "migu", "CETV3"},
    {"中国教育4", "migu", "CETV4"},
    {"上海教育", "migu", "SETV"},
    {"山东教育", "migu", "SDETV"},
    {"早期教育", "migu", "ZQJY"},
    {"留学生", "migu", "LXS"},
}

var client = &http.Client{Timeout: 15 * time.Second}

// 缓存结构
type Cache struct {
    data []byte
    time time.Time
    mu   sync.RWMutex
}

var cache Cache

// 获取CNTV数据
func fetchCNTV(channelID, date string) ([]Programme, error) {
    url := fmt.Sprintf("https://api.cntv.cn/epg/epginfo3?serviceId=shiyi&d=%s&c=%s", date, channelID)
    
    resp, err := client.Get(url)
    if err != nil {
        return nil, fmt.Errorf("CNTV请求失败: %v", err)
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("CNTV读取失败: %v", err)
    }

    programs := gjson.Get(string(body), channelID+".program")
    if !programs.Exists() || !programs.IsArray() {
        return nil, fmt.Errorf("CNTV无节目数据")
    }

    var list []Programme
    programs.ForEach(func(_, p gjson.Result) bool {
        title := p.Get("t").String()
        if title == "" {
            return true
        }

        start := p.Get("st").Int()
        end := p.Get("et").Int()

        programme := Programme{
            Channel: getChannelName(channelID),
            Start:   time.Unix(start, 0).Format("20060102150405") + " +0800",
            Stop:    time.Unix(end, 0).Format("20060102150405") + " +0800",
            Title:   Title{Lang: "zh", Value: escapeXML(title)},
        }
        list = append(list, programme)
        return true
    })

    return list, nil
}

// 获取咪咕数据
func fetchMigu(programID, date string) ([]Programme, error) {
    url := fmt.Sprintf("https://program-sc.miguvideo.com/live/v2/tv-programs-data/%s/%s", programID, date)
    
    resp, err := client.Get(url)
    if err != nil {
        return nil, fmt.Errorf("咪咕请求失败: %v", err)
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("咪咕读取失败: %v", err)
    }

    if len(body) == 0 {
        return nil, fmt.Errorf("咪咕返回空数据")
    }

    programs := gjson.Get(string(body), "body.program.0.content")
    if !programs.Exists() || !programs.IsArray() {
        return nil, fmt.Errorf("咪咕无节目数据")
    }

    var list []Programme
    programs.ForEach(func(_, p gjson.Result) bool {
        title := p.Get("contName").String()
        if title == "" {
            return true
        }

        start := p.Get("startTime").Int()
        end := p.Get("endTime").Int()

        programme := Programme{
            Channel: getChannelName(programID),
            Start:   time.UnixMilli(start).Format("20060102150405") + " +0800",
            Stop:    time.UnixMilli(end).Format("20060102150405") + " +0800",
            Title:   Title{Lang: "zh", Value: escapeXML(title)},
        }
        list = append(list, programme)
        return true
    })

    return list, nil
}

// XML转义
func escapeXML(s string) string {
    s = strings.ReplaceAll(s, "&", "&amp;")
    s = strings.ReplaceAll(s, "<", "&lt;")
    s = strings.ReplaceAll(s, ">", "&gt;")
    s = strings.ReplaceAll(s, "\"", "&quot;")
    s = strings.ReplaceAll(s, "'", "&apos;")
    return s
}

// 根据ID获取频道名称
func getChannelName(id string) string {
    for _, c := range channels {
        if c.ID == id {
            return c.Name
        }
    }
    return id
}

// 获取频道数据的工作函数
func fetchChannelProgrammes(ch ChannelConfig, date string) []Programme {
    var programmes []Programme
    var err error
    
    if ch.Source == "cntv" {
        programmes, err = fetchCNTV(ch.ID, date)
    } else {
        programmes, err = fetchMigu(ch.ID, date)
    }
    
    if err != nil {
        log.Printf("获取失败 [%s] %s: %v", ch.Source, ch.Name, err)
        return nil
    }
    
    return programmes
}

// 生成EPG数据
func generateEPG(days int) []byte {
    log.Printf(strings.Repeat("=", 50))
    log.Printf("开始生成EPG，频道数: %d，天数: %d", len(channels), days)
    
    tv := TV{
        Generator: "EPG Server",  // generator-info-name属性
        URL:       "",             // generator-info-url属性
    }
    
    // 添加频道
    for _, ch := range channels {
        tv.Channels = append(tv.Channels, Channel{
            ID:      ch.Name,
            Display: DisplayName{Lang: "zh", Value: ch.Name},
        })
    }
    
    var mu sync.Mutex
    var wg sync.WaitGroup
    semaphore := make(chan struct{}, 10)
    
    now := time.Now()
    for day := 0; day < days; day++ {
        date := now.AddDate(0, 0, day).Format("20060102")
        log.Printf("处理日期: %s", date)
        
        for _, ch := range channels {
            wg.Add(1)
            go func(ch ChannelConfig, date string) {
                defer wg.Done()
                semaphore <- struct{}{}
                defer func() { <-semaphore }()
                
                programmes := fetchChannelProgrammes(ch, date)
                if len(programmes) > 0 {
                    mu.Lock()
                    tv.Programmes = append(tv.Programmes, programmes...)
                    mu.Unlock()
                    log.Printf("成功获取 [%s] %s - %d条节目", ch.Source, ch.Name, len(programmes))
                }
            }(ch, date)
        }
    }
    
    wg.Wait()
    log.Printf("所有频道获取完成，节目总数: %d", len(tv.Programmes))
    log.Printf(strings.Repeat("=", 50))
    
    // 生成XML
    xmlData, err := xml.MarshalIndent(tv, "", "  ")
    if err != nil {
        log.Printf("XML生成失败: %v", err)
        return []byte(xml.Header + `<tv generator-info-name="EPG Server"></tv>`)
    }
    
    return []byte(xml.Header + string(xmlData))
}

// HTTP处理器
func epgHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/xml; charset=utf-8")
    w.Header().Set("Access-Control-Allow-Origin", "*")
    
    // 从查询参数获取天数，默认为3
    days := 3
    if d := r.URL.Query().Get("days"); d != "" {
        fmt.Sscanf(d, "%d", &days)
        if days < 1 || days > 7 {
            days = 3
        }
    }
    
    // 强制刷新
    forceRefresh := r.URL.Query().Get("refresh") == "1"
    
    // 检查缓存（2小时）
    cache.mu.RLock()
    cacheValid := !forceRefresh && time.Since(cache.time) < 2*time.Hour && len(cache.data) > 0
    cache.mu.RUnlock()
    
    if cacheValid {
        cache.mu.RLock()
        w.Write(cache.data)
        cache.mu.RUnlock()
        log.Printf("返回缓存数据，年龄: %v", time.Since(cache.time))
        return
    }
    
    // 生成新数据
    xmlData := generateEPG(days)
    if len(xmlData) > 0 {
        cache.mu.Lock()
        cache.data = xmlData
        cache.time = time.Now()
        cache.mu.Unlock()
        w.Write(xmlData)
    } else {
        emptyXML := []byte(xml.Header + `<tv generator-info-name="EPG Server"></tv>`)
        w.Write(emptyXML)
    }
}

// 首页处理器
func indexHandler(w http.ResponseWriter, r *http.Request) {
    html := fmt.Sprintf(`
    <!DOCTYPE html>
    <html>
    <head>
        <title>EPG Server</title>
        <meta charset="utf-8">
        <style>
            body { font-family: Arial; padding: 20px; background: #f5f5f5; }
            .container { max-width: 800px; margin: 0 auto; background: white; padding: 30px; border-radius: 10px; }
            h2 { color: #333; }
            .info { background: #e8f4fd; padding: 15px; border-radius: 5px; margin: 20px 0; }
            .links a { display: inline-block; margin: 5px 0; color: #0066cc; }
        </style>
    </head>
    <body>
        <div class="container">
            <h2>📺 EPG 节目单服务器</h2>
            <div class="info">
                <p>✅ 频道总数: <strong>%d</strong> 个</p>
                <p>📅 默认生成天数: <strong>%d</strong> 天</p>
                <p>⏱️ 缓存时间: <strong>2小时</strong></p>
            </div>
            <div class="links">
                <p>🔗 <a href="/epg.xml" target="_blank">/epg.xml</a> - 获取EPG数据(默认3天)</p>
                <p>🔗 <a href="/epg.xml?days=1" target="_blank">/epg.xml?days=1</a> - 获取1天数据</p>
                <p>🔗 <a href="/epg.xml?days=7" target="_blank">/epg.xml?days=7</a> - 获取7天数据</p>
                <p>📊 <a href="/stats" target="_blank">/stats</a> - 查看统计</p>
            </div>
        </div>
    </body>
    </html>
    `, len(channels), 3)
    
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    fmt.Fprint(w, html)
}

// 统计处理器
func statsHandler(w http.ResponseWriter, r *http.Request) {
    cache.mu.RLock()
    cacheTime := cache.time.Format("2006-01-02 15:04:05")
    cacheAge := time.Since(cache.time).String()
    dataSize := len(cache.data)
    cache.mu.RUnlock()
    
    stats := fmt.Sprintf(`{
        "channels": %d,
        "default_days": 3,
        "cache_time": "%s",
        "cache_age": "%s",
        "cache_size_bytes": %d,
        "version": "1.0.0"
    }`, 
        len(channels),
        cacheTime,
        cacheAge,
        dataSize)
    
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    fmt.Fprint(w, stats)
}

func main() {
    http.HandleFunc("/", indexHandler)
    http.HandleFunc("/epg.xml", epgHandler)
    http.HandleFunc("/stats", statsHandler)
    
    log.Printf(strings.Repeat("=", 50))
    log.Printf("🚀 EPG服务器启动成功")
    log.Printf("📌 端口: 8082")
    log.Printf("📺 频道总数: %d", len(channels))
    log.Printf("📅 默认生成天数: 3")
    log.Printf("🔗 访问地址: http://localhost:8082/epg.xml")
    log.Printf(strings.Repeat("=", 50))
    
    log.Fatal(http.ListenAndServe(":8082", nil))
}