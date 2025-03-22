package stores

type StateStore interface {
	DocumentStore() DocumentStore
}
