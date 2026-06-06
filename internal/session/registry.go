package session

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/izzzzzi/agent-asearch/internal/state"
)

type Record struct {
	SID       string            `json:"sid"`
	Name      string            `json:"name"`
	Query     string            `json:"query"`
	Sources   []string          `json:"sources"`
	CreatedAt time.Time         `json:"created_at"`
	TotalHits int               `json:"total_hits"`
	State     string            `json:"state"`
	Meta      map[string]string `json:"meta,omitempty"`
}

func Save(r Record) error {
	_ = state.EnsureDirs()
	p := registryPath()
	var reg registry
	data, _ := os.ReadFile(p)
	_ = json.Unmarshal(data, &reg)
	reg.Sessions = append(reg.Sessions, r)
	data, err := json.MarshalIndent(reg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0600)
}

func Get(sid string) (Record, bool) {
	reg, err := load()
	if err != nil {
		return Record{}, false
	}
	for _, r := range reg.Sessions {
		if r.SID == sid {
			return r, true
		}
	}
	return Record{}, false
}

func List() []Record {
	reg, err := load()
	if err != nil {
		return nil
	}
	sort.Slice(reg.Sessions, func(i, j int) bool {
		return reg.Sessions[i].CreatedAt.After(reg.Sessions[j].CreatedAt)
	})
	return reg.Sessions
}

func UpdateState(sid, state string) error {
	reg, err := load()
	if err != nil {
		return err
	}
	for i, r := range reg.Sessions {
		if r.SID == sid {
			reg.Sessions[i].State = state
			break
		}
	}
	data, err := json.MarshalIndent(reg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(registryPath(), data, 0600)
}

func UpdateTotal(sid string, total int) error {
	reg, err := load()
	if err != nil {
		return err
	}
	for i, r := range reg.Sessions {
		if r.SID == sid {
			reg.Sessions[i].TotalHits = total
			break
		}
	}
	data, err := json.MarshalIndent(reg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(registryPath(), data, 0600)
}

func Remove(sid string) error {
	reg, err := load()
	if err != nil {
		return err
	}
	filtered := reg.Sessions[:0]
	for _, r := range reg.Sessions {
		if !strings.EqualFold(r.SID, sid) {
			filtered = append(filtered, r)
		}
	}
	reg.Sessions = filtered
	data, err := json.MarshalIndent(reg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(registryPath(), data, 0600)
}

func Lock(sid string) error {
	f := filepath.Join(state.SessionsDir(), sid+".lock")
	return os.WriteFile(f, []byte("locked"), 0600)
}

func Unlock(sid string) error {
	f := filepath.Join(state.SessionsDir(), sid+".lock")
	return os.Remove(f)
}

type registry struct {
	Sessions []Record `json:"sessions"`
}

func load() (*registry, error) {
	p := registryPath()
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return &registry{}, nil
		}
		return nil, err
	}
	var reg registry
	if err := json.Unmarshal(data, &reg); err != nil {
		return &registry{}, nil
	}
	return &reg, nil
}

func registryPath() string {
	return filepath.Join(state.StateDir(), "sessions.json")
}
