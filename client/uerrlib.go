package client

func (p *Proc) Errstr() string {
	res := p.errstr
	p.errstr = ""
	return res
}
