package model

// import (
// 	"github.com/coderun-top/coderun/src/utils"
// )

// ProcStore persists process information to storage.
type ProcStore interface {
	ProcLoad(int64) (*Proc, error)
	ProcFind(*Build, int) (*Proc, error)
	ProcChild(*Build, int, string) (*Proc, error)
	ProcList(*Build) ([]*Proc, error)
	ProcCreate([]*Proc) error
	ProcUpdate(*Proc) error
	ProcClear(*Build) error
}

// Proc represents a process in the build pipeline.
// swagger:model proc
type Proc struct {
    ID       int64             `json:"id"                    gorm:"AUTO_INCREMENT;primary_key;column:proc_id"`
	BuildID  string            `json:"build_id"              gorm:"column:proc_build_id;unique_index:idx_proc"`
	PID      int               `json:"pid"                   gorm:"column:proc_pid;unique_index:idx_proc"`
	PPID     int               `json:"ppid"                  gorm:"column:proc_ppid"`
	PGID     int               `json:"pgid"                  gorm:"column:proc_pgid"`
	Name     string            `json:"name"                  gorm:"column:proc_name"`
	State    string            `json:"state"                 gorm:"column:proc_state"`
	Error    string            `json:"error,omitempty"       gorm:"type:varchar(500);column:proc_error"`
	ExitCode int               `json:"exit_code"             gorm:"column:proc_exit_code"`
	Started  int64             `json:"start_time,omitempty"  gorm:"column:proc_started"`
	Stopped  int64             `json:"end_time,omitempty"    gorm:"column:proc_stopped"`
	Machine  string            `json:"machine,omitempty"     gorm:"column:proc_machine"`
	Platform string            `json:"platform,omitempty"    gorm:"column:proc_platform"`
	Environ  MapType           `json:"environ,omitempty"     gorm:"type:mediumblob;column:proc_environ;json"`
	Children []*Proc           `json:"children,omitempty"    gorm:"-"`
}

// func (t *Proc) BeforeCreate() (err error) {
// 	t.ID = utils.GeneratorId()
// 	return nil
// }

// Running returns true if the process state is pending or running.
func (p *Proc) Running() bool {
	return p.State == StatusPending || p.State == StatusRunning
}

// Failing returns true if the process state is failed, killed or error.
func (p *Proc) Failing() bool {
	return p.State == StatusError || p.State == StatusKilled || p.State == StatusFailure
}

// Tree creates a process tree from a flat process list.
func Tree(procs []*Proc) []*Proc {
	var (
		nodes  []*Proc
		parent *Proc
	)
	for _, proc := range procs {
		if proc.PPID == 0 {
			nodes = append(nodes, proc)
			parent = proc
			continue
		} else {
			parent.Children = append(parent.Children, proc)
		}
	}
	return nodes
}
