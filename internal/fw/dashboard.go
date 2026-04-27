package fw

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type dashPayload struct {
	Period        string         `json:"period"`
	Generated     string         `json:"generated"`
	TotalFocusMin int            `json:"total_focus_min"`
	TotalBlocks   int            `json:"total_blocks"`
	TotalSwitches int            `json:"total_switches"`
	FilesChanged  int            `json:"files_changed"`
	LinesAdded    int            `json:"lines_added"`
	LinesRemoved  int            `json:"lines_removed"`
	ByProject     map[string]int `json:"by_project"`
	ByTool        map[string]int `json:"by_tool"`
	Languages     map[string]int `json:"languages"`
	Heatmap       []hourBucket   `json:"heatmap"`     // 7×24 grid for week, 1×48 (30-min) for today
	Timeline      []tlBlock      `json:"timeline"`
	StreakDays    int            `json:"streak_days"`
	TopRepo       string         `json:"top_repo"`
	TopBranch     string         `json:"top_branch"`
}

type hourBucket struct {
	Day    string `json:"day"`    // "Mon 22"
	Hour   int    `json:"hour"`   // 0-23 (week) or 0-47 half-hour (today)
	Minute int    `json:"minute"` // focused minutes in the bucket
}

type tlBlock struct {
	Start    string `json:"start"`
	End      string `json:"end"`
	Repo     string `json:"repo"`
	Branch   string `json:"branch"`
	Focus    int    `json:"focus"`
	Switches int    `json:"switches"`
	Summary  string `json:"summary"`
}

func buildDashPayload(period string, blocks []Block) dashPayload {
	p := dashPayload{
		Period:    period,
		Generated: time.Now().Local().Format("Mon 02 Jan, 15:04"),
		ByProject: map[string]int{},
		ByTool:    map[string]int{},
		Languages: map[string]int{},
	}
	repoMin := map[string]int{}
	for _, b := range blocks {
		p.TotalFocusMin += b.FocusedMin
		p.TotalSwitches += b.Switches
		p.FilesChanged += b.FilesAdded
		p.LinesAdded += b.LinesAdded
		p.LinesRemoved += b.LinesDel
		for k, v := range b.ByProject {
			p.ByProject[k] += v
		}
		for k, v := range b.ByTool {
			p.ByTool[k] += v
		}
		for k, v := range b.Languages {
			p.Languages[k] += v
		}
		if b.Repo != "" {
			repoMin[b.Repo] += b.FocusedMin
			if b.FocusedMin > 0 && b.Branch != "" {
				p.TopBranch = b.Branch
			}
		}
		p.Timeline = append(p.Timeline, tlBlock{
			Start:    b.StartTS.Local().Format("15:04"),
			End:      b.EndTS.Local().Format("15:04"),
			Repo:     b.Repo,
			Branch:   b.Branch,
			Focus:    b.FocusedMin,
			Switches: b.Switches,
			Summary:  b.Summary,
		})
	}
	p.TotalBlocks = len(blocks)
	p.TopRepo = topKey(repoMin)
	p.Heatmap = buildHeatmap(period, blocks)
	p.StreakDays = streak(blocks)
	return p
}

func buildHeatmap(period string, blocks []Block) []hourBucket {
	if period == "week" {
		// 7 days × 24 hours
		bm := map[string]map[int]int{}
		for _, b := range blocks {
			day := b.StartTS.Local().Format("Mon 02")
			h := b.StartTS.Local().Hour()
			if bm[day] == nil {
				bm[day] = map[int]int{}
			}
			bm[day][h] += b.FocusedMin
		}
		var out []hourBucket
		// stable order: oldest day first
		now := time.Now().Local()
		for i := 6; i >= 0; i-- {
			d := now.AddDate(0, 0, -i)
			label := d.Format("Mon 02")
			row := bm[label]
			for h := 0; h < 24; h++ {
				out = append(out, hourBucket{Day: label, Hour: h, Minute: row[h]})
			}
		}
		return out
	}
	// today: 48 half-hour buckets
	buckets := make([]int, 48)
	for _, b := range blocks {
		l := b.StartTS.Local()
		idx := l.Hour()*2 + l.Minute()/30
		if idx >= 0 && idx < 48 {
			buckets[idx] += b.FocusedMin
		}
	}
	out := make([]hourBucket, 48)
	for i, m := range buckets {
		out[i] = hourBucket{Day: "Today", Hour: i, Minute: m}
	}
	return out
}

func streak(blocks []Block) int {
	if len(blocks) == 0 {
		return 0
	}
	days := map[string]bool{}
	for _, b := range blocks {
		if b.FocusedMin > 0 {
			days[b.StartTS.Local().Format("2006-01-02")] = true
		}
	}
	streak := 0
	d := time.Now().Local()
	for {
		k := d.Format("2006-01-02")
		if !days[k] {
			break
		}
		streak++
		d = d.AddDate(0, 0, -1)
	}
	return streak
}

// RenderDashboard writes a self-contained HTML file.
func RenderDashboard(blocks []Block, period, outPath string) error {
	data := buildDashPayload(period, blocks)
	js, err := json.Marshal(data)
	if err != nil {
		return err
	}
	html := strings.Replace(dashboardHTML, "__FLOWD_DATA__", string(js), 1)
	if err := os.MkdirAll(filepath.Dir(outPath), 0750); err != nil {
		return err
	}
	return os.WriteFile(outPath, []byte(html), 0644)
}

// OpenInBrowser opens a path/URL in the user's default browser.
func OpenInBrowser(path string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", path).Start()
	case "linux":
		return exec.Command("xdg-open", path).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", path).Start()
	}
	return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
}

// dashboardHTML is the single-file HTML template. %s is the JSON payload.
const dashboardHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<title>flowd — dashboard</title>
<meta name="viewport" content="width=device-width,initial-scale=1">
<script src="https://cdn.jsdelivr.net/npm/chart.js@4.4.1/dist/chart.umd.min.js"></script>
<style>
:root{
  --bg:#0a0e14;--bg2:#11161d;--card:#161c25;--line:#222a35;
  --fg:#e6edf3;--mute:#8b95a6;--accent:#7c5cff;--accent2:#42d4b6;
  --warn:#ffb454;--bad:#ff6e6e;--good:#5dd87b;
  --grad:linear-gradient(135deg,#7c5cff 0%,#42d4b6 100%);
}
*{box-sizing:border-box}
html,body{margin:0;padding:0;background:var(--bg);color:var(--fg);font-family:-apple-system,BlinkMacSystemFont,"SF Pro Display","Inter",system-ui,sans-serif;font-feature-settings:"ss01","cv11"}
body{background:radial-gradient(1200px 600px at 10% -10%,rgba(124,92,255,.18),transparent 60%),radial-gradient(900px 500px at 100% 0%,rgba(66,212,182,.14),transparent 60%),var(--bg);min-height:100vh}
.wrap{max-width:1280px;margin:0 auto;padding:32px 28px 64px}
header{display:flex;align-items:center;justify-content:space-between;margin-bottom:28px}
.brand{display:flex;align-items:center;gap:14px}
.logo{width:40px;height:40px;border-radius:11px;background:var(--grad);display:grid;place-items:center;font-weight:800;font-size:18px;color:#0a0e14;letter-spacing:-.5px;box-shadow:0 8px 30px -10px rgba(124,92,255,.6)}
h1{font-size:22px;margin:0;letter-spacing:-.4px}
.muted{color:var(--mute);font-size:13px}
.tabs{display:flex;gap:6px;background:var(--card);padding:4px;border-radius:10px;border:1px solid var(--line)}
.tab{padding:7px 14px;border-radius:7px;font-size:13px;color:var(--mute);cursor:default}
.tab.on{background:var(--bg2);color:var(--fg)}
.grid{display:grid;gap:18px}
.kpis{grid-template-columns:repeat(4,1fr)}
@media(max-width:900px){.kpis{grid-template-columns:repeat(2,1fr)}}
.card{background:var(--card);border:1px solid var(--line);border-radius:16px;padding:20px;position:relative;overflow:hidden}
.card.glow::before{content:"";position:absolute;inset:0;background:var(--grad);opacity:.06;pointer-events:none}
.kpi .label{font-size:12px;color:var(--mute);text-transform:uppercase;letter-spacing:.7px;font-weight:600}
.kpi .num{font-size:34px;font-weight:700;letter-spacing:-1px;margin-top:6px;line-height:1}
.kpi .sub{font-size:12px;color:var(--mute);margin-top:6px}
.kpi.accent .num{background:var(--grad);-webkit-background-clip:text;background-clip:text;color:transparent}
.row{display:grid;gap:18px;margin-top:18px}
.row.r2{grid-template-columns:1.4fr 1fr}
.row.r3{grid-template-columns:1fr 1fr 1fr}
@media(max-width:980px){.row.r2,.row.r3{grid-template-columns:1fr}}
h2{font-size:14px;margin:0 0 14px;font-weight:600;color:var(--fg);letter-spacing:.2px}
h2 .hint{font-weight:400;color:var(--mute);font-size:12px;margin-left:6px}
.heatmap{display:grid;gap:3px}
.heatmap .row-line{display:flex;align-items:center;gap:8px}
.heatmap .lbl{font-size:11px;color:var(--mute);width:46px;text-align:right;font-variant-numeric:tabular-nums}
.heatmap .cells{display:grid;grid-template-columns:repeat(24,1fr);gap:3px;flex:1}
.heatmap.today .cells{grid-template-columns:repeat(48,1fr)}
.heatmap .cell{aspect-ratio:1;border-radius:3px;background:#1a212c;transition:transform .1s}
.heatmap .cell:hover{transform:scale(1.4);outline:1px solid var(--accent)}
.scale{display:flex;justify-content:flex-end;align-items:center;gap:8px;margin-top:10px;color:var(--mute);font-size:11px}
.scale .sw{width:11px;height:11px;border-radius:3px}
.timeline{display:flex;flex-direction:column;gap:8px;max-height:380px;overflow:auto;padding-right:6px}
.timeline::-webkit-scrollbar{width:8px}
.timeline::-webkit-scrollbar-thumb{background:#222b39;border-radius:4px}
.tl-item{display:flex;gap:12px;padding:12px 14px;background:var(--bg2);border-radius:10px;border:1px solid transparent;transition:border-color .15s}
.tl-item:hover{border-color:var(--line)}
.tl-time{font-variant-numeric:tabular-nums;color:var(--mute);font-size:12px;min-width:96px;padding-top:2px}
.tl-body{flex:1;min-width:0}
.tl-head{display:flex;align-items:baseline;gap:8px;flex-wrap:wrap}
.repo{font-weight:600;color:var(--fg)}
.branch{font-size:11px;color:var(--accent2);background:rgba(66,212,182,.1);padding:2px 7px;border-radius:99px}
.tl-stats{margin-top:6px;font-size:12px;color:var(--mute)}
.bar{height:6px;background:#1a212c;border-radius:3px;overflow:hidden;margin-top:8px}
.bar i{display:block;height:100%;background:var(--grad)}
.lang-list{display:flex;flex-direction:column;gap:10px}
.lang-row{display:flex;align-items:center;gap:10px;font-size:13px}
.lang-row .name{flex:0 0 110px;color:var(--fg)}
.lang-row .min{color:var(--mute);font-variant-numeric:tabular-nums;font-size:12px;width:60px;text-align:right}
.lang-row .barwrap{flex:1;height:8px;background:#1a212c;border-radius:4px;overflow:hidden}
.lang-row .barwrap i{display:block;height:100%;background:var(--grad)}
.ai-card{background:linear-gradient(135deg,rgba(124,92,255,.08),rgba(66,212,182,.04));border:1px dashed #2c3242;border-radius:16px;padding:24px;display:flex;align-items:center;gap:18px;min-height:140px}
.ai-icon{width:46px;height:46px;border-radius:12px;background:var(--grad);display:grid;place-items:center;font-size:22px;flex:0 0 46px}
.ai-card .title{font-weight:600;font-size:14px;margin-bottom:4px}
.ai-card .sub{color:var(--mute);font-size:13px;line-height:1.5}
.empty{color:var(--mute);text-align:center;padding:40px;font-size:13px}
canvas{max-height:260px}
footer{margin-top:36px;color:var(--mute);font-size:12px;text-align:center}
</style>
</head>
<body>
<div class="wrap">
  <header>
    <div class="brand">
      <div class="logo">fw</div>
      <div>
        <h1>flowd</h1>
        <div class="muted" id="genTime"></div>
      </div>
    </div>
    <div class="tabs"><div class="tab" id="tabT">Today</div><div class="tab" id="tabW">Week</div></div>
  </header>

  <div id="empty" class="card empty" style="display:none">No activity recorded yet for this period. Run <code>fw start</code> and check back.</div>

  <div id="content">
    <div class="grid kpis">
      <div class="card kpi accent glow"><div class="label">Focus</div><div class="num" id="kFocus">—</div><div class="sub" id="kFocusSub"></div></div>
      <div class="card kpi"><div class="label">Top repo</div><div class="num" id="kRepo" style="font-size:22px">—</div><div class="sub" id="kBranch"></div></div>
      <div class="card kpi"><div class="label">Code</div><div class="num" id="kCode" style="font-size:22px">—</div><div class="sub" id="kCodeSub"></div></div>
      <div class="card kpi"><div class="label">Streak</div><div class="num" id="kStreak">—</div><div class="sub">consecutive active days</div></div>
    </div>

    <div class="row r2">
      <div class="card">
        <h2 id="hmTitle">Activity heatmap <span class="hint" id="hmHint"></span></h2>
        <div class="heatmap" id="heatmap"></div>
        <div class="scale">less <span class="sw" style="background:#1a212c"></span><span class="sw" style="background:#2b2151"></span><span class="sw" style="background:#4a3aaa"></span><span class="sw" style="background:#7c5cff"></span><span class="sw" style="background:#42d4b6"></span> more</div>
      </div>
      <div class="ai-card">
        <div class="ai-icon">✨</div>
        <div>
          <div class="title">AI insights</div>
          <div class="sub">Coming soon. Patterns, focus tips, and weekly summaries generated from your activity will appear here.</div>
        </div>
      </div>
    </div>

    <div class="row r3">
      <div class="card"><h2>By project</h2><canvas id="chProject"></canvas></div>
      <div class="card"><h2>By tool</h2><canvas id="chTool"></canvas></div>
      <div class="card"><h2>Languages <span class="hint">(weighted by lines touched)</span></h2><div id="langList" class="lang-list"></div></div>
    </div>

    <div class="row r2">
      <div class="card">
        <h2>Timeline</h2>
        <div class="timeline" id="timeline"></div>
      </div>
      <div class="card">
        <h2>Summary</h2>
        <div id="textSummary" style="font-size:13px;color:var(--mute);line-height:1.7"></div>
      </div>
    </div>
  </div>

  <footer>flowd — local activity tracker · self-hosted</footer>
</div>

<script>
const DATA = __FLOWD_DATA__;

(function(){
  document.getElementById('genTime').textContent = "Generated " + DATA.generated;
  if (DATA.period === "week") {
    document.getElementById('tabW').classList.add('on');
    document.getElementById('hmTitle').firstChild.textContent = "Activity heatmap ";
    document.getElementById('hmHint').textContent = "(last 7 days × hour of day)";
  } else {
    document.getElementById('tabT').classList.add('on');
    document.getElementById('hmHint').textContent = "(today · 30-min buckets)";
  }

  if (!DATA.total_blocks){
    document.getElementById('empty').style.display='block';
    document.getElementById('content').style.display='none';
    return;
  }

  // KPIs
  const h = Math.floor(DATA.total_focus_min/60), m = DATA.total_focus_min%60;
  document.getElementById('kFocus').textContent = (h ? h+"h " : "") + m + "m";
  document.getElementById('kFocusSub').textContent = DATA.total_blocks + " blocks · " + DATA.total_switches + " switches";
  document.getElementById('kRepo').textContent = DATA.top_repo || "—";
  document.getElementById('kBranch').textContent = DATA.top_branch ? "branch: " + DATA.top_branch : "";
  document.getElementById('kCode').textContent = DATA.files_changed + " files";
  document.getElementById('kCodeSub').textContent = "+" + DATA.lines_added + " −" + DATA.lines_removed;
  document.getElementById('kStreak').textContent = DATA.streak_days + " day" + (DATA.streak_days===1?"":"s");

  // Heatmap
  const hm = document.getElementById('heatmap');
  const max = Math.max(1, ...DATA.heatmap.map(c=>c.minute));
  const colorFor = v => {
    if (!v) return "#1a212c";
    const t = v/max;
    if (t<.25) return "#2b2151";
    if (t<.5)  return "#4a3aaa";
    if (t<.75) return "#7c5cff";
    return "#42d4b6";
  };
  if (DATA.period === "week") {
    const byDay = {};
    DATA.heatmap.forEach(c=>{(byDay[c.day]=byDay[c.day]||[]).push(c)});
    Object.keys(byDay).forEach(day=>{
      const row = document.createElement('div'); row.className='row-line';
      const lbl = document.createElement('div'); lbl.className='lbl'; lbl.textContent=day;
      const cells = document.createElement('div'); cells.className='cells';
      byDay[day].sort((a,b)=>a.hour-b.hour).forEach(c=>{
        const el = document.createElement('div'); el.className='cell';
        el.style.background = colorFor(c.minute);
        el.title = day + " " + String(c.hour).padStart(2,'0')+":00 · "+c.minute+"m";
        cells.appendChild(el);
      });
      row.appendChild(lbl); row.appendChild(cells); hm.appendChild(row);
    });
  } else {
    hm.classList.add('today');
    const row = document.createElement('div'); row.className='row-line';
    const lbl = document.createElement('div'); lbl.className='lbl'; lbl.textContent='Today';
    const cells = document.createElement('div'); cells.className='cells';
    DATA.heatmap.forEach(c=>{
      const el = document.createElement('div'); el.className='cell';
      el.style.background = colorFor(c.minute);
      const hh=Math.floor(c.hour/2), mm=(c.hour%2)*30;
      el.title = String(hh).padStart(2,'0')+":"+String(mm).padStart(2,'0')+" · "+c.minute+"m";
      cells.appendChild(el);
    });
    row.appendChild(lbl); row.appendChild(cells); hm.appendChild(row);
  }

  // Charts
  Chart.defaults.color = "#8b95a6";
  Chart.defaults.font.family = "-apple-system,BlinkMacSystemFont,Inter,system-ui,sans-serif";
  Chart.defaults.borderColor = "#222a35";
  const palette = ["#7c5cff","#42d4b6","#ffb454","#ff6e6e","#5dd87b","#5b9dff","#e667d4","#a884ff"];

  function donut(id, data){
    const labels = Object.keys(data).filter(k=>data[k]>0).sort((a,b)=>data[b]-data[a]);
    const vals = labels.map(k=>data[k]);
    if (!labels.length){
      document.getElementById(id).parentElement.querySelector('canvas').replaceWith(emptyMsg());
      return;
    }
    new Chart(document.getElementById(id),{
      type:'doughnut',
      data:{labels,datasets:[{data:vals,backgroundColor:labels.map((_,i)=>palette[i%palette.length]),borderWidth:0}]},
      options:{
        plugins:{legend:{position:'bottom',labels:{boxWidth:10,padding:10,font:{size:11}}},
        tooltip:{callbacks:{label:c=>c.label+': '+c.parsed+'m'}}},
        cutout:'62%'
      }
    });
  }
  function emptyMsg(){const d=document.createElement('div');d.className='empty';d.textContent='No data';return d}

  donut('chProject', DATA.by_project);
  donut('chTool',    DATA.by_tool);

  // Languages list
  const ll = document.getElementById('langList');
  const langs = Object.entries(DATA.languages).filter(([,v])=>v>0).sort((a,b)=>b[1]-a[1]).slice(0,8);
  if (!langs.length){ ll.innerHTML='<div class="empty">No languages detected.<br><span class="muted">Languages are inferred from files changed in git during the window.</span></div>'; }
  const lmax = Math.max(1, ...langs.map(l=>l[1]));
  langs.forEach(([k,v],i)=>{
    const r=document.createElement('div'); r.className='lang-row';
    r.innerHTML = '<div class="name">'+k+'</div><div class="barwrap"><i style="width:'+(100*v/lmax)+'%;background:'+palette[i%palette.length]+'"></i></div><div class="min">'+v+'m</div>';
    ll.appendChild(r);
  });

  // Timeline
  const tl = document.getElementById('timeline');
  const tlMax = Math.max(1, ...DATA.timeline.map(b=>b.focus));
  if (!DATA.timeline.length){ tl.innerHTML='<div class="empty">No blocks.</div>'; }
  DATA.timeline.slice().reverse().forEach(b=>{
    const el=document.createElement('div'); el.className='tl-item';
    el.innerHTML = '<div class="tl-time">'+b.start+' → '+b.end+'</div>'+
      '<div class="tl-body">'+
      '<div class="tl-head"><span class="repo">'+(b.repo||"—")+'</span>'+
      (b.branch?'<span class="branch">'+b.branch+'</span>':'')+
      '</div>'+
      '<div class="tl-stats">'+b.focus+'m focus · '+b.switches+' switches</div>'+
      '<div class="bar"><i style="width:'+(100*b.focus/tlMax)+'%"></i></div>'+
      '</div>';
    tl.appendChild(el);
  });

  // Text summary
  const top = arr => arr.slice(0,3).map(([k,v])=>k+' '+v+'m').join(' · ') || '—';
  const ts = document.getElementById('textSummary');
  const projTop = Object.entries(DATA.by_project).sort((a,b)=>b[1]-a[1]);
  const toolTop = Object.entries(DATA.by_tool).sort((a,b)=>b[1]-a[1]);
  ts.innerHTML =
    '<p><strong style="color:var(--fg)">Focus:</strong> '+DATA.total_focus_min+' min across '+DATA.total_blocks+' blocks.</p>'+
    '<p><strong style="color:var(--fg)">Projects:</strong> '+top(projTop)+'</p>'+
    '<p><strong style="color:var(--fg)">Tools:</strong> '+top(toolTop)+'</p>'+
    '<p><strong style="color:var(--fg)">Code:</strong> '+DATA.files_changed+' files (+'+DATA.lines_added+' −'+DATA.lines_removed+').</p>';
})();
</script>
</body>
</html>`
