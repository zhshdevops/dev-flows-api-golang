/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2017 TenxCloud. All Rights Reserved.
 *
 * 2017-05-22  @author lizhen
 */

package blobs

type TypeMeta struct {
	Kind       string `json:"kind,omitempty"`
	APIVersion string `json:"apiVersion,omitempty"`
}

type Privilege int

const (
	OthersWrite   = 128
	TeammateWrite = 64
	SelfWrite     = 32

	OthersRead   = 4
	TeammateRead = 2
	SelfRead     = 1

	SelfReadWrite     = SelfRead | SelfWrite
	TeammateReadWrite = TeammateRead | TeammateWrite
	OthersReadWrite   = OthersRead | OthersWrite

	AllReadable = SelfRead | TeammateRead | OthersRead
	AllWritable = SelfWrite | TeammateWrite | OthersWrite
	All         = SelfReadWrite | TeammateReadWrite | OthersReadWrite

	Magic = -1
)

func (p Privilege) SelfReadable() bool {
	return p.AbleTo(SelfRead)
}

func (p Privilege) SelfWritable() bool {
	return p.AbleTo(SelfWrite)
}

func (p Privilege) TeammateReadable() bool {
	return p.AbleTo(TeammateRead)
}

func (p Privilege) TeammateWritable() bool {
	return p.AbleTo(TeammateWrite)
}

func (p Privilege) OthersReadable() bool {
	return p.AbleTo(OthersRead)
}

func (p Privilege) OthersWritable() bool {
	return p.AbleTo(OthersWrite)
}

func (p Privilege) AbleTo(do Privilege) bool {
	return p&do == do
}
