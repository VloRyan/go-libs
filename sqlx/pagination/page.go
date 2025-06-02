package pagination

type Page struct {
	Offset     int      `json:"offset"`
	Limit      int      `json:"limit"`
	Sort       []string `json:"sort"`
	TotalCount int      `json:"totalCount"`
}

func NewPage(offset, limit int, sorts ...string) *Page {
	return &Page{
		Offset: offset,
		Limit:  limit,
		Sort:   sorts,
	}
}

func First(sorts ...string) *Page {
	return &Page{
		Offset: 0,
		Limit:  1,
		Sort:   sorts,
	}
}

func SortBy(sorts ...string) *Page {
	return &Page{
		Offset: 0,
		Limit:  -1,
		Sort:   sorts,
	}
}

var Limitless *Page = nil

func (p *Page) Next() *Page {
	if p.Offset == p.TotalCount {
		return nil
	}
	return &Page{
		Offset: p.Offset + 1,
		Limit:  p.Limit,
		Sort:   p.Sort,
	}
}

func (p *Page) Prev() *Page {
	if p.Offset == 0 {
		return nil
	}
	return &Page{
		Offset: p.Offset - 1,
		Limit:  p.Limit,
		Sort:   p.Sort,
	}
}

func (p *Page) SortBy(name string) *Page {
	p.Sort = append(p.Sort, name)
	return p
}
