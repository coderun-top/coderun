package model


type LogData struct {
	ID     int64  `meddler:"log_id,pk"    gorm:"AUTO_INCREMENT;primary_key;column:log_id"`
	ProcID int64  `meddler:"log_job_id"   gorm:"type:integer;column:log_job_id"`
	Data   []byte `meddler:"log_data"     gorm:"type:mediumblob;column:log_data"`
}

// 定义生成表的名称
func (LogData) TableName() string {
	return "logs"
  }
