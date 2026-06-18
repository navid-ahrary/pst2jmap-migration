package model

type Job struct {
	Email    string `csv:"email"`
	Password string `csv:"password"`
	PSTFile  string `csv:"pstfile"`
	Row      int
}
