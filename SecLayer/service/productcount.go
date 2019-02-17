package service

import "sync"

type ProductCountMgr struct {
	ProductCount map[int]int
	lock         sync.RWMutex
}

func (p *ProductCountMgr) Count(productId int) (count int) {
	p.lock.RLock()
	defer p.lock.RUnlock()
	count = p.ProductCount[productId]
	return
}

func (p *ProductCountMgr) Add(productId, count int) {
	p.lock.Lock()
	defer p.lock.Unlock()

	cur, ok := p.ProductCount[productId]
	if !ok {
		cur = count
	} else {
		cur += count
	}

	p.ProductCount[productId] = cur
}

func NewProductCountMgr() (productMgr *ProductCountMgr) {
	productMgr = &ProductCountMgr{
		ProductCount: make(map[int]int, 128),
	}
	return
}
