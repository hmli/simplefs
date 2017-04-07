package core

type Directory interface {
	Get(id uint64) (n *Needle, err error)
	Set(n *Needle) (err error)
	Has(id uint64) (has bool)
	Del(id uint64) (err error)
	Next() (id uint64, hasNext bool)
}
