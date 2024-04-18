package ihui

// Paginator is a paginator. 
// It is used to split a list of items into pages.

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

// SetTotal sets the total number of items
func (p *Paginator) SetTotal(size int) {
	if p.Size != size {
		p.Size = size
		p.SetPage(1)
	}
}

// Pages returns the number of pages
func (p *Paginator) Pages() int {
	nb := int(math.Ceil(float64(p.Size) / float64(p.PageSize)))
	p.Fragments = nil
	for i := 0; i < nb; i++ {
		p.Fragments = append(p.Fragments, &Fragment{pagination: p, Number: i + 1, Index: p.PageSize * i})
	}
	return nb
}

// SetPage sets the current page
func (p *Paginator) SetPage(index int) {
	if index > 0 && index <= p.Pages() {
		p.iCurrent = index
		p.Current = p.Fragments[p.iCurrent-1]
	}
}

// PreviousPage sets the current page to the previous page
func (p *Paginator) PreviousPage() {
	p.SetPage(p.iCurrent - 1)
}

// NextPage sets the current page to the next page
func (p *Paginator) NextPage() {
	p.SetPage(p.iCurrent + 1)
}
