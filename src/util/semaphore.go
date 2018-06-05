package util

var sem chan struct{}

func InitSem(n int) {
	sem = make(chan struct{}, n)
}

func AcquireSem() {
	sem <- struct{}{}
}

func ReleaseSem() {
	<-sem
}
