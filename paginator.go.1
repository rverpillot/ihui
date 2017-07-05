package ihui

import "math"

type Fragment struct {
	pagination *Paginator
	Number     int
	Index      int
}

func (f *Fragment) Active() string {
	if f.pagination.Current == f {
		return "active"
	}
	return ""
}

type Paginator struct {
	Size      int
	PageSize  int
	Fragments []*Fragment
	Current   *Fragment
	iCurrent  int
}

func NewPaginator(pagesize int) *Paginator {
	p := &Paginator{Size: pagesize, PageSize: pagesize}
	p.SetPage(1)
	return p
}

func (p *Paginator) SetTotal(size int) {
	if p.Size != size {
		p.Size = size
		p.SetPage(1)
	}
}

func (p *Paginator) Pages() int {
	nb := int(math.Ceil(float64(p.Size) / float64(p.PageSize)))
	p.Fragments = nil
	for i := 0; i < nb; i++ {
		p.Fragments = append(p.Fragments, &Fragment{pagination: p, Number: i + 1, Index: p.PageSize * i})
	}
	return nb
}

func (p *Paginator) SetPage(index int) {
	if index > 0 && index <= p.Pages() {
		p.iCurrent = index
		p.Current = p.Fragments[p.iCurrent-1]
	}
}

func (p *Paginator) PreviousPage() {
	p.SetPage(p.iCurrent - 1)
}

func (p *Paginator) NextPage() {
	p.SetPage(p.iCurrent + 1)
}
