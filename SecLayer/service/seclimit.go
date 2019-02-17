package service

/*type TimeLimit interface {
	Count(nowTime int64) (curCount int)
	Check(nowTime int64) int
}*/

type SecLimit struct {
	count   int
	curTime int64
}

func (p *SecLimit) Count(nowTime int64) (curCount int) {
	if nowTime-p.curTime > 60 {
		p.count = 1
		p.curTime = nowTime
		curCount = p.count
		return
	}

	p.count++
	curCount = p.count
	return
}

func (p *SecLimit) Check(nowTime int64) int {
	if nowTime-p.curTime > 60 {
		return 0
	}

	return p.count
}
