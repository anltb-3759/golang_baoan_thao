package utils

type Map map[string]interface{}

type Pagination struct {
	Page     int   `json:"page"`
	Limit    int   `json:"limit"`
	Total    int64 `json:"total"`
	From     int   `json:"from"`
	To       int   `json:"to"`
	PrevPage int   `json:"prev_page"`
	NextPage int   `json:"next_page"`
}

type PaginationData struct {
	Page     int   `json:"page"`
	Limit    int   `json:"limit"`
	Total    int64 `json:"total"`
	From     int   `json:"from"`
	To       int   `json:"to"`
	PrevPage int   `json:"prev_page"`
	NextPage int   `json:"next_page"`
	HasPrev  bool  `json:"has_prev"`
	HasNext  bool  `json:"has_next"`
}

func NewPagination(page, limit int, total int64) PaginationData {
	from := (page-1)*limit + 1
	to := page * limit
	if int64(to) > total {
		to = int(total)
	}
	if total == 0 {
		from = 0
	}
	totalPages := int((total + int64(limit) - 1) / int64(limit))
	prevPage := 0
	if page > 1 {
		prevPage = page - 1
	}
	nextPage := 999999
	if page < totalPages {
		nextPage = page + 1
	}
	return PaginationData{
		Page:     page,
		Limit:    limit,
		Total:    total,
		From:     from,
		To:       to,
		PrevPage: prevPage,
		NextPage: nextPage,
		HasPrev:  page > 1,
		HasNext:  page < totalPages,
	}
}
