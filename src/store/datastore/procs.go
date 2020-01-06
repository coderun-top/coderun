// Copyright 2018 Drone.IO Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package datastore

import (
	"github.com/coderun-top/coderun/src/model"
	// "github.com/coderun-top/coderun/src/store/datastore/sql"
	// "github.com/russross/meddler"
	
)

func (db *datastore) ProcLoad(id int64) (*model.Proc, error) {
	// stmt := sql.Lookup(db.driver, "procs-find-id")
	// proc := new(model.Proc)
	// err := meddler.QueryRow(db, proc, stmt, id)
	// return proc, err

	proc := new(model.Proc)
	err := db.Where("proc_id = ?",id).First(&proc).Error
	return proc, err
}

func (db *datastore) ProcFind(build *model.Build, pid int) (*model.Proc, error) {
	// stmt := sql.Lookup(db.driver, "procs-find-build-pid")
	// proc := new(model.Proc)
	// err := meddler.QueryRow(db, proc, stmt, build.ID, pid)
	// return proc, err
	proc := new(model.Proc)
	err := db.Where("proc_build_id = ? AND proc_pid = ?",build.ID, pid).First(&proc).Error
	return proc, err
}

func (db *datastore) ProcChild(build *model.Build, pid int, procName string) (*model.Proc, error) {
	// stmt := sql.Lookup(db.driver, "procs-find-build-ppid")
	// proc := new(model.Proc)
	// err := meddler.QueryRow(db, proc, stmt, build.ID, pid, child)
	// return proc, err

	proc := new(model.Proc)
	err := db.Where("proc_build_id = ? AND proc_ppid = ? AND proc_name = ?", build.ID, pid, procName).First(&proc).Error
	return proc, err
}

func (db *datastore) ProcList(build *model.Build) ([]*model.Proc, error) {
	// stmt := sql.Lookup(db.driver, "procs-find-build")
	// list := []*model.Proc{}
	// err := meddler.QueryAll(db, &list, stmt, build.ID)
	// return list, err

	list := []*model.Proc{}
	err := db.Where("proc_build_id = ?",build.ID).Order("proc_id ASC").Find(&list).Error
	return list, err
}

func (db *datastore) ProcCreate(procs []*model.Proc) error {
	// for _, proc := range procs {
	// 	if err := meddler.Insert(db, "procs", proc); err != nil {
	// 		return err
	// 	}
	// }
	// return nil

	for _, proc := range procs {
		if err := db.Create(proc).Error; err != nil {
			return err
		}
	}
	return nil

}

func (db *datastore) ProcUpdate(proc *model.Proc) error {
	// return meddler.Update(db, "procs", proc)

	// 由于.save 在不存在的时候，会创建新的数据，所以需要判断是否存在
	var count int
	err := db.Model(&model.Proc{}).Where("proc_id = ?",proc.ID).Count(&count).Error
	if err != nil || count == 0{
		return nil
	}

	return db.Save(proc).Error
}

func (db *datastore) ProcClear(build *model.Build) (err error) {
	// stmt1 := sql.Lookup(db.driver, "files-delete-build")
	// stmt2 := sql.Lookup(db.driver, "procs-delete-build")
	// _, err = db.Exec(stmt1, build.ID)
	// if err != nil {
	// 	return
	// }
	// _, err = db.Exec(stmt2, build.ID)
	// return


	// 开启事务
	tx := db.Begin()
	
	err = tx.Where("file_build_id = ?", build.ID).Delete(model.File{}).Error
	if err != nil{
		tx.Rollback()
		return err
	}
	err = tx.Where("proc_build_id = ?", build.ID).Delete(model.Proc{}).Error
	if err != nil{
		tx.Rollback()
		return err
	}

	// 提交事务
	tx.Commit()
	return nil
}
