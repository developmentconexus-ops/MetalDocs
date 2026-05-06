package application

type ListOptions struct {
	Page        int
	PageSize    int
	CreatedBy   string
	Status      []string
	AreaCode    string
	ProfileCode string
	Q           string
}

func (o ListOptions) Offset() int {
	if o.Page < 1 {
		return 0
	}
	return (o.Page - 1) * o.Limit()
}

func (o ListOptions) Limit() int {
	if o.PageSize == 0 {
		return 20
	}
	if o.PageSize < 1 {
		return 1
	}
	if o.PageSize > 50 {
		return 50
	}
	return o.PageSize
}
