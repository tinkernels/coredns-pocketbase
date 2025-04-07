package pocketbase

import (
	"github.com/pocketbase/pocketbase"
)

type Instance struct {
	pb *pocketbase.PocketBase
}

func New(pb *pocketbase.PocketBase) *Instance {
	return &Instance{
		pb: pb,
	}
}
