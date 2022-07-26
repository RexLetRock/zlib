package zbandwidth

import (
  "fmt"
	"time"
	"bufio"
	"os"
	"strconv"
	"strings"

  "sync/atomic"
)

type ZBandwidth struct {
  nettop *NetTop
  rx *atomic.Value
  tx *atomic.Value

  brx *atomic.Value
  btx *atomic.Value
  dif *atomic.Value
}

func New(iface string) *ZBandwidth {
  return newOptions(iface)
}

func (f *ZBandwidth) Get() (rx float64, tx float64) {
  return f.rx.Load().(float64), f.tx.Load().(float64)
}

func (f *ZBandwidth) GetRx() (ret float64) {
  return f.rx.Load().(float64)
}

func (f *ZBandwidth) GetTx() (ret float64) {
  return f.tx.Load().(float64)
}

func (f *ZBandwidth) GetString() (ret string) {
  if dif := f.dif.Load().(float64); dif != float64(0) {
    return fmt.Sprintf("\nRx: %v, Tx: %v \n", Vsize(f.brx.Load().(uint64), dif), Vsize(f.btx.Load().(uint64), dif))
  } else {
    return fmt.Sprintf("\nRx: 0, Tx: 0 \n")
  }
}

func newOptions(iface string) *ZBandwidth {
  f := &ZBandwidth{
    nettop: NewNetTop(),
    rx: new(atomic.Value),
    tx: new(atomic.Value),

    brx: new(atomic.Value),
    btx: new(atomic.Value),
    dif: new(atomic.Value),
  }

  f.rx.Store(float64(0))
  f.tx.Store(float64(0))

  f.brx.Store(uint64(0))
  f.btx.Store(uint64(0))
  f.dif.Store(float64(0))

  f.nettop.Interface = iface

  f.bgIntervalJob()
  go f.bgInterval()
  return f
}

func (f *ZBandwidth) bgInterval() {
  ticker := time.NewTicker(100 * time.Millisecond)
  quit := make(chan struct{})
  go func() {
    for {
      select {
        case <- ticker.C:
          f.bgIntervalJob()
        case <- quit:
          ticker.Stop()
          return
      }
    }
  }()
}

func (f *ZBandwidth) bgIntervalJob() {
	delta, dt := f.nettop.Update()
  iface := f.nettop.Interface
	dtf := dt.Seconds()
	if stat, ok := delta.Stat[iface]; ok {
    f.rx.Store(float64(stat.Rx)/dtf)
    f.tx.Store(float64(stat.Tx)/dtf)

    f.brx.Store(stat.Rx)
    f.btx.Store(stat.Tx)
    f.dif.Store(dtf)
  }
}

type NetTop struct {
	delta *NetStat
	last *NetStat
	t0 time.Time
	dt time.Duration
	Interface string
}
func NewNetTop() *NetTop {
	nt := &NetTop{
		delta: NewNetStat(),
		last: NewNetStat(),
		t0: time.Now(),
		dt: 1500 * time.Millisecond,
		Interface: "*",
	}
	return nt
}

func (nt *NetTop) Update() (*NetStat, time.Duration) {
	stat1 := nt.getInfo()
	nt.dt = time.Since(nt.t0)

	for _, value := range stat1.Dev {
		t0, ok := nt.last.Stat[value]
		if !ok {
			continue
		}

		dev, ok := nt.delta.Stat[value]
		if !ok {
			nt.delta.Stat[value] = new(DevStat)
			dev = nt.delta.Stat[value]
			nt.delta.Dev = append(nt.delta.Dev, value)
		}
		t1 := stat1.Stat[value]
		dev.Rx = t1.Rx - t0.Rx
		dev.Tx = t1.Tx - t0.Tx
	}
	nt.last = &stat1
	nt.t0 = time.Now()

	return nt.delta, nt.dt
}

func (nt *NetTop) getInfo() (ret NetStat) {

	lines, _ := ReadLines("/proc/net/dev")

	ret.Dev = make([]string, 0)
	ret.Stat = make(map[string]*DevStat)

	for _, line := range lines {
		fields := strings.Split(line, ":")
		if len(fields) < 2 {
			continue
		}
		key := strings.TrimSpace(fields[0])
		value := strings.Fields(strings.TrimSpace(fields[1]))

		if nt.Interface != "*" && nt.Interface != key {
			continue
		}

		c := new(DevStat)
		c.Name = key
		r, err := strconv.ParseInt(value[0], 10, 64)
		if err != nil {
			break
		}
		c.Rx = uint64(r)

		t, err := strconv.ParseInt(value[8], 10, 64)
		if err != nil {
			break
		}
		c.Tx = uint64(t)

		ret.Dev = append(ret.Dev, key)
		ret.Stat[key] = c
	}

	return
}


type NetStat struct {
	Dev  []string
	Stat map[string]*DevStat
}
func NewNetStat() *NetStat {
	return &NetStat{
		Dev: make([]string, 0),
		Stat: make(map[string]*DevStat),
	}
}

type DevStat struct {
	Name string
	Rx   uint64
	Tx   uint64
}

func ReadLines(filename string) ([]string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return []string{""}, err
	}
	defer f.Close()
	var ret []string
	r := bufio.NewReader(f)
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			break
		}
		ret = append(ret, strings.Trim(line, "\n"))
	}
	return ret, nil
}

func Vsize(bytes uint64, delta float64) (ret string) {
	var tmp float64 = float64(bytes) / delta
	var s string = " "

	bytes = uint64(tmp)

	switch {
	case bytes < uint64(2<<9):

	case bytes < uint64(2<<19):
		tmp = tmp / float64(2<<9)
		s = "K"

	case bytes < uint64(2<<29):
		tmp = tmp / float64(2<<19)
		s = "M"

	case bytes < uint64(2<<39):
		tmp = tmp / float64(2<<29)
		s = "G"

	case bytes < uint64(2<<49):
		tmp = tmp / float64(2<<39)
		s = "T"

	}
	ret = fmt.Sprintf("%06.2f %sB/s", tmp, s)
	return
}
