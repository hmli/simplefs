package core

type Directory interface {
	Get(id uint64) (n *Needle, err error)
	New(n *Needle) (err error)
	Has(id uint64) (has bool)
	Del(id uint64) (err error)
	Set(id uint64, n *Needle) (err error)
	Iter() (iter Iterator)
}

type Iterator interface {
	Next() (key []byte, exists bool)
	Release()
}
