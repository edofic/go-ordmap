package ordmap

type Key interface {
	Less(Key) bool
}

type Value interface{}
