package gollfs

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Server struct {
	Uri  string
	Urio int
}

type config struct {
	Timeout  int
	Watchdog bool
	Interval int
	Server   []Server
}

type LLamaServers struct {
	cfg     config
	srv     []Server
	refresh func()
}

func (l *LLamaServers) Best() (Server, error) {
	if len(l.srv) == 0 {
		return Server{}, fmt.Errorf("Missing LLamaserver")
	}
	return l.srv[0], nil
}

func NewLlamaServers() *LLamaServers {
	cfg := readConfiguration()
	err := checkConfig(cfg)
	if err != nil {
		log.Fatal(err)
	}
	result := LLamaServers{cfg: *cfg, srv: make([]Server, 0, 3)}
	result.refresh = func() {
		for {
			time.Sleep(time.Second * time.Duration(cfg.Interval))
			result.contactServer()
			fmt.Println(result)
		}
	}
	if cfg.Watchdog {
		go result.refresh()
	}
	result.contactServer()
	return &result
}

const (
	SECONDPATH          = "/etc/ollfs.conf"
	FIRSTPATH           = "./ollfs.conf"
	MISSINGCONFIG       = "Config File is missing."
	WATCHDOG_REXP       = "watchdog=(yes|no)$"
	WATCHDOG_MISSING    = "watchdog=(yes|no) missing in config"
	INTERVAL_REXP       = "interval=[0-9]*$"
	SERVER_REXP         = "server=.*"
	CONFIG_ERROR_MSG    = "Could not read Server Entry in Configuration."
	SERVER_POSTFIX      = "/api/tags"
	SERVER_TIMEOUT_REXP = "timeout.*=.*$"
	COMMENT_REXP        = `^(?!\s*#).+$`
)

func comment(s string) bool {
	for i := 0; i < len(s); i++ {
		if (s)[i] == ' ' || (s)[i] == '\t' || (s)[i] == '\n' || (s)[i] == '\r' {
			continue
		}
		if (s)[i] == '#' {
			return true
		} else {
			break
		}
	}
	return false
}

func readConfiguration() *config {
	file, err := os.Open(FIRSTPATH)
	if err != nil {
		file, err = os.Open(SECONDPATH)
		if err != nil {
			panic(MISSINGCONFIG)
		}
	}
	scanner := bufio.NewScanner(file)
	cfg := config{}
	for scanner.Scan() {
		if comment(scanner.Text()) {
			continue
		}

		match, _ := regexp.MatchString(WATCHDOG_REXP, scanner.Text())
		if match {
			if strings.Contains(scanner.Text(), "yes") {
				cfg.Watchdog = true
			} else if strings.Contains(scanner.Text(), "no") {
				cfg.Watchdog = false
			}
			continue
		}

		match, _ = regexp.MatchString(SERVER_TIMEOUT_REXP, scanner.Text())
		if match {
			result := strings.Split(scanner.Text(), "=")[1]
			intres, _ := strconv.Atoi(result)
			cfg.Timeout = intres
			continue
		}

		match, _ = regexp.MatchString(INTERVAL_REXP, scanner.Text())
		if match {
			result := strings.Split(scanner.Text(), "=")[1]
			intres, _ := strconv.Atoi(result)
			cfg.Interval = intres
			continue
		}
		match, _ = regexp.MatchString(SERVER_REXP, scanner.Text())
		if match {
			result := strings.Split(scanner.Text(), "=")
			if len(result) != 2 {
				log.Fatalln(CONFIG_ERROR_MSG)
			}
			rescomma := strings.Split(result[1], ",")
			if len(rescomma) != 2 {
				log.Fatalln(CONFIG_ERROR_MSG)
			}
			prio, err := strconv.Atoi(rescomma[1])
			if err != nil {
				log.Fatalln(CONFIG_ERROR_MSG)
			}
			s := Server{uri: rescomma[0], prio: prio}
			cfg.Server = append(cfg.Server, s)
			continue
		}
	}
	return &cfg
}

func checkConfig(cfg *config) error {
	if cfg.Interval == 0 {
		return fmt.Errorf("interval is not set.")
	}
	if cfg.Timeout == 0 {
		return fmt.Errorf("timeout is not set.")
	}
	if len(cfg.Server) == 0 {
		return fmt.Errorf("no servers configured.")
	}
	return nil
}

func (l *LLamaServers) contactServer() {
	var mu sync.Mutex
	var wg sync.WaitGroup
	l.srv = nil

	client := http.Client{Timeout: time.Millisecond * time.Duration(l.cfg.Timeout)}
	for _, srv := range l.cfg.Server {
		wg.Add(1)
		go func() {
			defer wg.Done()
			uri := fmt.Sprintf("%s%s", srv.uri, SERVER_POSTFIX)
			r, err := client.Head(uri)
			if err != nil || r.StatusCode != 200 {
				return
			}
			mu.Lock()
			(l.srv) = append(l.srv, srv)
			mu.Unlock()
		}()
	}
	wg.Wait()
	sort.Slice(l.srv, func(i, j int) bool {
		return l.srv[i].prio < l.srv[j].prio
	})
}
