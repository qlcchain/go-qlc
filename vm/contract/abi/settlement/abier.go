package settlement

type ABIer interface {
	ToABI() ([]byte, error)
	FromABI(data []byte) error
}

type Verifier interface {
	Verify() error
}
