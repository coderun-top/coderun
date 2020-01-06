package model

import (
	"github.com/coderun-top/coderun/src/utils"
)

// swagger:model build
type Build struct {
    ID        string  `json:"id"                  gorm:"primary_key;column:build_id"`
	RepoID    int64   `json:"-"                   gorm:"type:integer;column:build_repo_id;unique_index:idx_build"`
	ConfigID  string  `json:"-"                   gorm:"type:varchar(250);column:build_config_id"`
	Number    int     `json:"number"              gorm:"type:integer;column:build_number;unique_index:idx_build"`
	Parent    int     `json:"parent"              gorm:"type:integer;column:build_parent"`
	Event     string  `json:"event"               gorm:"type:varchar(500);column:build_event"`
	Status    string  `json:"status"              gorm:"type:varchar(500);column:build_status"`
	Error     string  `json:"error"               gorm:"type:varchar(500);column:build_error"`
	Enqueued  int64   `json:"enqueued_at"         gorm:"type:integer;column:build_enqueued"`
	Created   int64   `json:"created_at"          gorm:"type:integer;column:build_created"`
	Started   int64   `json:"started_at"          gorm:"type:integer;column:build_started"`
	Finished  int64   `json:"finished_at"         gorm:"type:integer;column:build_finished"`
	Deploy    string  `json:"deploy_to"           gorm:"type:varchar(500);column:build_deploy"`
	Commit    string  `json:"commit"              gorm:"type:varchar(500);column:build_commit"`
	Branch    string  `json:"branch"              gorm:"type:varchar(500);column:build_branch"`
	Ref       string  `json:"ref"                 gorm:"type:varchar(500);column:build_ref"`
	Refspec   string  `json:"refspec"             gorm:"type:varchar(1000);column:build_refspec"`
	Remote    string  `json:"remote"              gorm:"type:varchar(500);column:build_remote"`
	Title     string  `json:"title"               gorm:"type:varchar(1000);column:build_title"`
	Message   string  `json:"message"             gorm:"type:varchar(2000);column:build_message"`
	Timestamp int64   `json:"timestamp"           gorm:"type:integer;column:build_timestamp"`
	Sender    string  `json:"sender"              gorm:"type:varchar(250);column:build_sender"`
	Author    string  `json:"author"              gorm:"type:varchar(500);column:build_author;index"`
	Avatar    string  `json:"author_avatar"       gorm:"type:varchar(1000);column:build_avatar"`
	Email     string  `json:"author_email"        gorm:"type:varchar(500);column:build_email"`
	Link      string  `json:"link_url"            gorm:"type:varchar(1000);column:build_link"`
	Signed    bool    `json:"signed"              gorm:"column:build_signed"`   // deprecate
	Verified  bool    `json:"verified"            gorm:"column:build_verified"` // deprecate
	Reviewer  string  `json:"reviewed_by"         gorm:"type:varchar(250);column:build_reviewer"`
	Reviewed  int64   `json:"reviewed_at"         gorm:"type:integer;column:build_reviewed"`
	Procs     []*Proc `json:"procs,omitempty"     gorm:"-"`
	Files     []*File `json:"files,omitempty"     gorm:"-"`
}

func (t *Build) BeforeCreate() (err error) {
	t.ID = utils.GeneratorId()
	return nil
}

// Trim trims string values that would otherwise exceed
// the database column sizes and fail to insert.
func (b *Build) Trim() {
	if len(b.Title) > 1000 {
		b.Title = b.Title[:1000]
	}
	if len(b.Message) > 2000 {
		b.Message = b.Message[:2000]
	}
}

type BuildRepo struct {
    *Build
    // Repo            *Repo `json:"repo"`
	RepoOwner       string `json:"repo_owner"                    gorm:"column:repo_owner"`
	RepoName        string `json:"repo_name"                     gorm:"column:repo_name"`
	RepoFullName    string `json:"repo_full_name"                gorm:"column:repo_full_name"`
	RepoAvatar      string `json:"repo_avatar_url,omitempty"     gorm:"column:repo_avatar"`
	RepoBranch      string `json:"repo_default_branch,omitempty" gorm:"column:repo_branch"`
	// User      *User `json:"user"`
	TokenName       string `json:"user_token_name"               gorm:"column:user_token_name"`
}
