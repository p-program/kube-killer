package killer

// TwoFace https://zh.wikipedia.org/wiki/%E5%8F%8C%E9%9D%A2%E4%BA%BA
type TwoFace interface {
	// BlackHand 是活着面对黑暗的现实，还是渴望永恒的安宁？
	BlackHand() *TwoFace
	DeserveDead(resource interface{}) bool
	DryRun() *TwoFace
	Kill() error
}
