package main

// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-4.1)#api-v283
import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

var (
	weburl           = "http://127.0.0.1:8080/"
	syncMaindata     = "api/v2/sync/maindata"
	syncTorrentPeers = "api/v2/sync/torrentPeers?hash="
	transferBanpeers = "api/v2/transfer/banPeers?peers="
	bannedClient     = []string{"XL", "Xunlei", "TorrentStorm"}
	// Xunlei hide its client name, only display version number, e.g. 7.9.45.5054
	xunleiHiddenClient = []string{"7.9"}
	suspeciousPort     = []int{12345, 2011, 2013, 54321}
	defautlPort        = []int{15000}
)

func main() {

	var interval = flag.Int("interval", 5, "The time interval in second to scan all torrents and there peers")
	flag.Parse()

	for {
		metaUrl := fmt.Sprintf("%s%s", weburl, syncMaindata)
		resp, err := http.Get(metaUrl)
		if err != nil {
			log.Fatalf("can't access %s, %s\n", metaUrl, err.Error())
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("error reading resp body: %s\n", err.Error())
		}

		var m MetaData
		err = json.Unmarshal(body, &m)
		if err != nil {
			log.Fatalln(err)
		}
		for k, v := range m.Torrents {
			torrentUrl := fmt.Sprintf("%s%s%s", weburl, syncTorrentPeers, k)
			torResp, err := http.Get(torrentUrl)
			if err != nil {
				log.Fatalf("can't access %s, %s\n", torrentUrl, err.Error())
			}
			defer torResp.Body.Close()
			body, err := io.ReadAll(torResp.Body)
			if err != nil {
				log.Fatalf("error reading torrent resp body: %s\n", err.Error())
			}
			var pd PeerData
			err = json.Unmarshal(body, &pd)
			for _, pi := range pd.Peers {
				b := 0
				if !(pi.Progress > 0 && pi.Downloaded > 0) {
					for c := range bannedClient {
						if strings.Contains(pi.Client, bannedClient[c]) {
							b += 2
						}
					}

					for _, port := range suspeciousPort {
						if pi.Port == port {
							b += 1
						}
					}

					for _, port := range defautlPort {
						if pi.Port == port {
							b += 2
						}
					}

					for _, port := range xunleiHiddenClient {
						if strings.HasPrefix(pi.Client, port) {
							b += 2
						}
					}
				}

				if pi.Progress == 0 && pi.Uploaded > 1024*1024*16 && strings.IndexByte(strings.ToLower(pi.Flags), 'u') != -1 {
					fmt.Printf("Find Progess | Uploaded: %s %f | %d\n", pi.Client, pi.Progress, +pi.Uploaded)
					b += 1
				}

				if pi.Progress == 0 && pi.Uploaded > 1024*1024*64 && strings.IndexByte(strings.ToLower(pi.Flags), 'u') != -1 {
					fmt.Printf("Find Progess | Uploaded: %s %f | %d\n", pi.Client, pi.Progress, +pi.Uploaded)
					b += 2
				}

				if b > 1 {
					banUrl := fmt.Sprintf("%s%s%s:%d", weburl, transferBanpeers, pi.Ip, pi.Port)
					_, err := http.Get(banUrl)
					if err != nil {
						log.Fatalf("can't access %s, %s\n", banUrl, err.Error())
					}
				}

				if b > 0 {
					fmt.Println(v.Name)
					fmt.Printf("%v | Find: %s:%d Value: %d Client: [%s]\n", time.Now().Local(), pi.Ip, pi.Port, b, pi.Client)
				}
			}
		}

		time.Sleep(time.Second * time.Duration(*interval))
	}
}

type PeerData struct {
	FullUpdate bool                 `json:"full_update"`
	Peers      map[string]*PeerInfo `json:"peers"`
	Rid        int                  `json:"rid"`
	ShowFlags  bool                 `json:"show_flags"`
}

type PeerInfo struct {
	Client      string  `json:"client"`
	Connection  string  `json:"connection"`
	Country     string  `json:"country"`
	CountryCode string  `json:"country_code"`
	DlSpeed     int     `json:"dl_speed"`
	Downloaded  int     `json:"downloaded"`
	Files       string  `json:"files"`
	Flags       string  `json:"flags"`
	FlagsDesc   string  `json:"flags_desc"`
	Ip          string  `json:"ip"`
	Port        int     `json:"port"`
	Progress    float64 `json:"progress"`
	Relevance   int     `json:"relevance"`
	UpSpeed     int     `json:"up_speed"`
	Uploaded    int     `json:"uploaded"`
}

type MetaData struct {
	Categories  Categories           `json:"categories"`
	FullUpdate  bool                 `json:"full_update"`
	Rid         int                  `json:"rid"`
	ServerState ServerState          `json:"server_state"`
	Tags        []interface{}        `json:"tags"`
	Torrents    map[string]*Torrents `json:"torrents"`
	Trackers    map[string][]string  `json:"trackers"`
}
type Categories struct {
}
type ServerState struct {
	AlltimeDl            int64  `json:"alltime_dl"`
	AlltimeUl            int64  `json:"alltime_ul"`
	AverageTimeQueue     int    `json:"average_time_queue"`
	ConnectionStatus     string `json:"connection_status"`
	DhtNodes             int    `json:"dht_nodes"`
	DlInfoData           int    `json:"dl_info_data"`
	DlInfoSpeed          int    `json:"dl_info_speed"`
	DlRateLimit          int    `json:"dl_rate_limit"`
	FreeSpaceOnDisk      int64  `json:"free_space_on_disk"`
	GlobalRatio          string `json:"global_ratio"`
	QueuedIoJobs         int    `json:"queued_io_jobs"`
	Queueing             bool   `json:"queueing"`
	ReadCacheHits        string `json:"read_cache_hits"`
	ReadCacheOverload    string `json:"read_cache_overload"`
	RefreshInterval      int    `json:"refresh_interval"`
	TotalBuffersSize     int    `json:"total_buffers_size"`
	TotalPeerConnections int    `json:"total_peer_connections"`
	TotalQueuedSize      int    `json:"total_queued_size"`
	TotalWastedSession   int    `json:"total_wasted_session"`
	UpInfoData           int    `json:"up_info_data"`
	UpInfoSpeed          int    `json:"up_info_speed"`
	UpRateLimit          int    `json:"up_rate_limit"`
	UseAltSpeedLimits    bool   `json:"use_alt_speed_limits"`
	WriteCacheOverload   string `json:"write_cache_overload"`
}
type Torrents struct {
	AddedOn           int     `json:"added_on"`
	AmountLeft        int     `json:"amount_left"`
	AutoTmm           bool    `json:"auto_tmm"`
	Availability      float64 `json:"availability"`
	Category          string  `json:"category"`
	Completed         int64   `json:"completed"`
	CompletionOn      int     `json:"completion_on"`
	ContentPath       string  `json:"content_path"`
	DlLimit           int     `json:"dl_limit"`
	Dlspeed           int     `json:"dlspeed"`
	Downloaded        int64   `json:"downloaded"`
	DownloadedSession int     `json:"downloaded_session"`
	Eta               int     `json:"eta"`
	FLPiecePrio       bool    `json:"f_l_piece_prio"`
	ForceStart        bool    `json:"force_start"`
	LastActivity      int     `json:"last_activity"`
	MagnetURI         string  `json:"magnet_uri"`
	MaxRatio          int     `json:"max_ratio"`
	MaxSeedingTime    int     `json:"max_seeding_time"`
	Name              string  `json:"name"`
	NumComplete       int     `json:"num_complete"`
	NumIncomplete     int     `json:"num_incomplete"`
	NumLeechs         int     `json:"num_leechs"`
	NumSeeds          int     `json:"num_seeds"`
	Priority          int     `json:"priority"`
	Progress          float64 `json:"progress"`
	Ratio             float64 `json:"ratio"`
	RatioLimit        int     `json:"ratio_limit"`
	SavePath          string  `json:"save_path"`
	SeedingTime       int     `json:"seeding_time"`
	SeedingTimeLimit  int     `json:"seeding_time_limit"`
	SeenComplete      int     `json:"seen_complete"`
	SeqDl             bool    `json:"seq_dl"`
	Size              int64   `json:"size"`
	State             string  `json:"state"`
	SuperSeeding      bool    `json:"super_seeding"`
	Tags              string  `json:"tags"`
	TimeActive        int     `json:"time_active"`
	TotalSize         int64   `json:"total_size"`
	Tracker           string  `json:"tracker"`
	TrackersCount     int     `json:"trackers_count"`
	UpLimit           int     `json:"up_limit"`
	Uploaded          int     `json:"uploaded"`
	UploadedSession   int     `json:"uploaded_session"`
	Upspeed           int     `json:"upspeed"`
}
