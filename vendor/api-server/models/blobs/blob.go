/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2017 TenxCloud. All Rights Reserved.
 *
 * 2017-05-22  @author lizhen
 */

package blobs

import (
	"api-server/models/team2user"
	"errors"
	"github.com/astaxie/beego/orm"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"time"
)

var AccessDenied error = errors.New("access denied")
var ContentOversize error = errors.New("content oversize")

var MaxBlobSizeInByte int64 = 1024 * 1024 * 10 // 10MB

func init() {
	var sizeStr string
	if sizeStr = os.Getenv("MAX_DB_BLOB_SIZE_IN_BYTE"); sizeStr == "" {
		return
	}
	var err error
	if MaxBlobSizeInByte, err = strconv.ParseInt(sizeStr, 10, 64); err != nil {
		panic(err)
	}
}

//const createTable = `CREATE TABLE IF NOT EXISTS tenx_blobs (
//  id INT AUTO_INCREMENT PRIMARY KEY,
//  content MEDIUMBLOB NOT NULL,
//  metadata varchar(128) NOT NULL,
//  privilege INT NOT NULL DEFAULT 231 COMMENT 'OthersWrite = 128, TeammateWrite = 64, SelfWrite = 32, OthersRead = 4, TeammateRead = 2, SelfRead = 1',
//  create_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
//  create_by INT NOT NULL COMMENT 'tenx_users.user_id'
//) ENGINE=InnoDB DEFAULT CHARSET=utf8;`

const tableName = "tenx_blobs"

type Blob struct {
	ID        int       `orm:"pk;column(id)"`
	Content   string    `orm:"column(content)"`
	Metadata  string    `orm:"column(metadata)"`
	Privilege Privilege `orm:"column(privilege)"`
	CreateAt  time.Time `orm:"column(create_at)"`
	CreateBy  int32     `orm:"column(create_by)"`
}

func (Blob) TableName() string {
	return tableName
}

func (Blob) TableEngine() string {
	return "InnoDB"
}

func InsertBlob(stream io.Reader, metadata string, privilege Privilege, userID int32) (id int, err error) {
	var content []byte
	if content, err = ioutil.ReadAll(stream); err != nil {
		return
	}
	if int64(len(content)) > MaxBlobSizeInByte {
		err = ContentOversize
		return
	}
	var id64 int64
	if id64, err = orm.NewOrm().Insert(&Blob{
		Content:   string(content),
		Metadata:  metadata,
		Privilege: privilege,
		CreateBy:  userID,
	}); err != nil {
		return
	}
	id = int(id64)
	return
}

func GetBlobByID(id int, userID int32) (content []byte, err error) {
	var blob *Blob
	if blob, err = getByID(id); err != nil {
		return
	}
	var readable bool
	if readable, err = checkReadable(blob, userID); err != nil {
		return
	} else if !readable {
		err = AccessDenied
		return
	}
	content = []byte(blob.Content)
	return
}

func DeleteBlobByID(id int, userID int32) (err error) {
	var blob *Blob
	if blob, err = getByID(id); err != nil {
		return
	}
	var writable bool
	if writable, err = checkWritable(blob, userID); err != nil {
		return
	} else if !writable {
		err = AccessDenied
		return
	}
	_, err = orm.NewOrm().Delete(blob)
	return
}

func ModifyBlobByID(id int, stream io.Reader, userID int32, privilege Privilege) (err error) {
	var content []byte
	if content, err = ioutil.ReadAll(stream); err != nil {
		return
	}
	if int64(len(content)) > MaxBlobSizeInByte {
		err = ContentOversize
		return
	}
	var blob *Blob
	if blob, err = getByID(id); err != nil {
		return
	}
	var writable bool
	if writable, err = checkWritable(blob, userID); err != nil {
		return
	} else if !writable {
		err = AccessDenied
		return
	}
	blob.Content = string(content)
	var column []string
	if privilege == Magic {
		column = []string{"content"}
	} else {
		blob.Privilege = privilege
		column = []string{"content", "privilege"}
	}
	_, err = orm.NewOrm().Update(blob, column...)
	return
}

func getByID(id int) (blob *Blob, err error) {
	blob = new(Blob)
	err = orm.NewOrm().QueryTable(tableName).Filter("id", id).One(blob)
	return
}

func GetContentFormatByID(id int) (blob *Blob, err error) {
	blob = new(Blob)
	err = orm.NewOrm().QueryTable(tableName).Filter("id", id).One(blob)
	return
}


func checkReadable(blob *Blob, userID int32) (bool, error) {
	if blob.Privilege.AbleTo(AllReadable) {
		return true, nil
	}
	if blob.CreateBy == userID {
		return blob.Privilege.SelfReadable(), nil
	}
	isTeammate, err := team2user.IsTeammate(userID, blob.CreateBy)
	if err != nil {
		return false, err
	}
	if isTeammate {
		return blob.Privilege.TeammateReadable(), nil
	}
	return blob.Privilege.OthersReadable(), nil
}

func checkWritable(blob *Blob, userID int32) (bool, error) {
	if blob.Privilege.AbleTo(AllWritable) {
		return true, nil
	}
	if blob.CreateBy == userID {
		return blob.Privilege.SelfWritable(), nil
	}
	isTeammate, err := team2user.IsTeammate(userID, blob.CreateBy)
	if err != nil {
		return false, err
	}
	if isTeammate {
		return blob.Privilege.TeammateWritable(), nil
	}
	return blob.Privilege.OthersWritable(), nil
}
