package reservation

type Reserver interface {
	Reserve(userID string, tokens int) (bool, error)
	Commit(userID string, tokens int) error
	Rollback(userID string, tokens int) error
}
