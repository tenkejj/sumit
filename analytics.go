package main

import (
	"encoding/csv"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

const analyticsFile = "analytics.csv"

var analyticsMu sync.Mutex

// Event to pojedynczy wpis analityki zapisywany w CSV.
type Event struct {
	Timestamp string
	Name      string
	Source    string
	Mobile    string
	UA        string
}

func isMobileUA(ua string) bool {
	u := strings.ToLower(ua)
	mobileHints := []string{
		"mobile", "android", "iphone", "ipod", "ipad",
		"windows phone", "webos", "blackberry", "opera mini", "opera mobi",
	}
	for _, hint := range mobileHints {
		if strings.Contains(u, hint) {
			return true
		}
	}
	return false
}

func truncateUA(ua string) string {
	ua = strings.TrimSpace(ua)
	if len(ua) > 80 {
		return ua[:80]
	}
	return ua
}

func appendEvent(ev Event) error {
	analyticsMu.Lock()
	defer analyticsMu.Unlock()

	_, err := os.Stat(analyticsFile)
	needHeader := os.IsNotExist(err)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	f, err := os.OpenFile(analyticsFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	if needHeader {
		if err := w.Write([]string{"timestamp", "event", "source", "mobile", "ua"}); err != nil {
			return err
		}
	}
	if err := w.Write([]string{ev.Timestamp, ev.Name, ev.Source, ev.Mobile, ev.UA}); err != nil {
		return err
	}
	w.Flush()
	return w.Error()
}

func recordEventFromRequest(name, source string, r *http.Request) error {
	ua := truncateUA(r.UserAgent())
	mobile := "nie"
	if isMobileUA(ua) {
		mobile = "tak"
	}
	ev := Event{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Name:      name,
		Source:    strings.TrimSpace(source),
		Mobile:    mobile,
		UA:        ua,
	}
	return appendEvent(ev)
}

func loadEvents() ([]Event, error) {
	analyticsMu.Lock()
	defer analyticsMu.Unlock()

	f, err := os.Open(analyticsFile)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	var events []Event
	for i, row := range records {
		if len(row) < 5 {
			continue
		}
		if i == 0 && row[0] == "timestamp" {
			continue
		}
		events = append(events, Event{
			Timestamp: row[0],
			Name:      row[1],
			Source:    row[2],
			Mobile:    row[3],
			UA:        row[4],
		})
	}
	return events, nil
}

func isLocalhost(r *http.Request) bool {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func statsAuthorized(r *http.Request) bool {
	if isLocalhost(r) {
		return true
	}
	key := os.Getenv("STATS_KEY")
	if key == "" {
		return false
	}
	return r.Header.Get("X-Stats-Key") == key
}

func handleTrack(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	r.Body = http.MaxBytesReader(w, r.Body, 4096)

	var req struct {
		Event  string `json:"event"`
		Source string `json:"source"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "nieprawidłowy JSON", http.StatusBadRequest)
		return
	}

	req.Event = strings.TrimSpace(req.Event)
	if req.Event == "" {
		http.Error(w, "brak event", http.StatusBadRequest)
		return
	}

	if err := recordEventFromRequest(req.Event, req.Source, r); err != nil {
		http.Error(w, "błąd zapisu", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func handleStatsPage(w http.ResponseWriter, r *http.Request) {
	// HTML panelu jest ukryty (brak linków w UI); dane wymagają localhost lub X-Stats-Key.
	http.ServeFile(w, r, "static/stats.html")
}

type dayEventCounts struct {
	Date          string `json:"date"`
	PageView      int    `json:"page_view"`
	PDFGenerated  int    `json:"pdf_generated"`
}

type funnelStep struct {
	Name       string  `json:"name"`
	Count      int     `json:"count"`
	Percent    float64 `json:"percent"`
	StepRate   float64 `json:"stepRate"`
}

type sourceRow struct {
	Source       string `json:"source"`
	PageViews    int    `json:"pageViews"`
	PDFGenerated int    `json:"pdfGenerated"`
}

type statsPayload struct {
	EventsPerDay []dayEventCounts `json:"eventsPerDay"`
	Funnel       []funnelStep     `json:"funnel"`
	TopSources   []sourceRow      `json:"topSources"`
	Mobile       struct {
		Mobile  int     `json:"mobile"`
		Desktop int     `json:"desktop"`
		MobileP float64 `json:"mobilePercent"`
	} `json:"mobile"`
	TotalEvents int `json:"totalEvents"`
}

func eventDay(ts string) string {
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		if len(ts) >= 10 {
			return ts[:10]
		}
		return ""
	}
	return t.UTC().Format("2006-01-02")
}

func buildStatsPayload(events []Event) statsPayload {
	today := time.Now().UTC()
	start := today.AddDate(0, 0, -29)

	dayMap := make(map[string]*dayEventCounts)
	for i := 0; i < 30; i++ {
		d := start.AddDate(0, 0, i).Format("2006-01-02")
		dayMap[d] = &dayEventCounts{Date: d}
	}

	counts := map[string]int{
		"page_view":       0,
		"link_copied":     0,
		"client_view":     0,
		"client_accepted": 0,
		"pdf_generated":   0,
		"xml_downloaded":  0,
	}

	sourceViews := make(map[string]int)
	sourcePDFs := make(map[string]int)
	mobilePV := 0
	desktopPV := 0

	for _, ev := range events {
		counts[ev.Name]++

		day := eventDay(ev.Timestamp)
		if dc, ok := dayMap[day]; ok {
			switch ev.Name {
			case "page_view":
				dc.PageView++
			case "pdf_generated":
				dc.PDFGenerated++
			}
		}

		src := ev.Source
		if src == "" {
			src = "(brak)"
		}
		switch ev.Name {
		case "page_view":
			sourceViews[src]++
			if ev.Mobile == "tak" {
				mobilePV++
			} else {
				desktopPV++
			}
		case "pdf_generated":
			sourcePDFs[src]++
		}
	}

	perDay := make([]dayEventCounts, 0, 30)
	for i := 0; i < 30; i++ {
		d := start.AddDate(0, 0, i).Format("2006-01-02")
		perDay = append(perDay, *dayMap[d])
	}

	base := counts["page_view"]
	funnelNames := []struct {
		key  string
		name string
	}{
		{"page_view", "page_view"},
		{"link_copied", "link_copied"},
		{"client_view", "client_view"},
		{"client_accepted", "client_accepted"},
	}

	var funnel []funnelStep
	var prevCount int
	for i, fn := range funnelNames {
		c := counts[fn.key]
		step := funnelStep{Name: fn.name, Count: c}
		if base > 0 {
			step.Percent = float64(c) / float64(base) * 100
		}
		if i > 0 && prevCount > 0 {
			step.StepRate = float64(c) / float64(prevCount) * 100
		} else if i > 0 {
			step.StepRate = 0
		} else {
			step.StepRate = 100
		}
		prevCount = c
		funnel = append(funnel, step)
	}

	sourceSet := make(map[string]struct{})
	for s := range sourceViews {
		sourceSet[s] = struct{}{}
	}
	for s := range sourcePDFs {
		sourceSet[s] = struct{}{}
	}

	var sources []sourceRow
	for s := range sourceSet {
		sources = append(sources, sourceRow{
			Source:       s,
			PageViews:    sourceViews[s],
			PDFGenerated: sourcePDFs[s],
		})
	}
	sort.Slice(sources, func(i, j int) bool {
		if sources[i].PageViews != sources[j].PageViews {
			return sources[i].PageViews > sources[j].PageViews
		}
		return sources[i].PDFGenerated > sources[j].PDFGenerated
	})
	if len(sources) > 20 {
		sources = sources[:20]
	}

	totalPV := mobilePV + desktopPV
	mobilePct := 0.0
	if totalPV > 0 {
		mobilePct = float64(mobilePV) / float64(totalPV) * 100
	}

	var payload statsPayload
	payload.EventsPerDay = perDay
	payload.Funnel = funnel
	payload.TopSources = sources
	payload.Mobile.Mobile = mobilePV
	payload.Mobile.Desktop = desktopPV
	payload.Mobile.MobileP = mobilePct
	payload.TotalEvents = len(events)
	return payload
}

func handleStatsData(w http.ResponseWriter, r *http.Request) {
	if !statsAuthorized(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	events, err := loadEvents()
	if err != nil {
		http.Error(w, "błąd odczytu danych", http.StatusInternalServerError)
		return
	}

	payload := buildStatsPayload(events)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, "błąd serializacji", http.StatusInternalServerError)
	}
}
