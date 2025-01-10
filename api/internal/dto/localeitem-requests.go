package dto

type CreateRequest struct {
	Lang    string
	Context string
	Content string
}

type UpdateRequest struct {
	AggregateId string
	Lang        string
	Content     string
}

type GetContextRequest struct {
	Context string
}

type GetDetailRequest struct {
	AggregateId string
}

type SearchRequest struct {
	Lang           string
	Context        string
	PartialContent string
}
