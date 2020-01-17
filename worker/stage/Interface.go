package stage

// Stage todo
type Stage interface {
	Do() error
	Next() Stage
}
