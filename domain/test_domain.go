package domain

import "fmt"

type Data struct {
	Id  int64
	Val string
}

func (d *Data) GetId() int64 {
	return d.Id
}

func (d *Data) ToString() string {
	return fmt.Sprintf("id: %d, val: %s", d.Id, d.Val)
}
